package lite

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
	"go.minekube.com/gate/pkg/util/netutil"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	rawRaknetifyRouteHintPacketID = byte(0xfe)
	rawRaknetifyRouteHintVersion  = byte(1)
	rawRaknetifyRouteHintVersion2 = byte(2)
	rawRaknetifyRouteTokenLen     = 16
	rawRaknetifyMaxHintHostLen    = 1024
	rawRaknetifyIdleTimeout       = time.Minute
	rawRaknetifySweepInterval     = 15 * time.Second
	rawRaknetifyMaxSessions       = 4096
	rawRaknetifySocketBufferSize  = 4 * 1024 * 1024
	rawRaknetifyDefaultIPTOS      = 0x18
	rawRaknetifyWriteTimeout      = 10 * time.Millisecond
)

var rawRaknetifyRouteHintMagic = []byte("GATE_RAKNET_ROUTE")

type rawRaknetifySession struct {
	host        string
	routeToken  string
	hasToken    bool
	backendAddr *net.UDPAddr
	backendConn *net.UDPConn
	decrement   func()
	lastSeen    atomic.Int64
	closeOnce   sync.Once
}

type rawRaknetifyRouteHint struct {
	host     string
	token    string
	hasToken bool
}

func (s *rawRaknetifySession) touch(now time.Time) {
	s.lastSeen.Store(now.UnixNano())
}

func (s *rawRaknetifySession) close() {
	s.closeOnce.Do(func() {
		_ = s.backendConn.Close()
		if s.decrement != nil {
			s.decrement()
		}
	})
}

// ServeRaknetifyRaw starts a UDP route-hint based raw Raknetify passthrough listener.
func ServeRaknetifyRaw(ctx context.Context, opts RaknetifyOptions) error {
	if opts.Routes == nil {
		return fmt.Errorf("raknetify routes provider is nil")
	}
	if opts.StrategyManager == nil {
		opts.StrategyManager = NewStrategyManager()
	}

	var lc net.ListenConfig
	pc, err := lc.ListenPacket(ctx, "udp", opts.Bind)
	if err != nil {
		return err
	}
	defer func() { _ = pc.Close() }()

	log := opts.Logger.WithName("lite").WithName("raknetify").WithName("raw").WithValues("bind", pc.LocalAddr())
	log.Info("raw raknetify lite listener started")
	if udpConn, ok := pc.(*net.UDPConn); ok {
		tuneRawRaknetifyUDPConn(log, "listener", udpConn)
	}

	srv := &rawRaknetifyServer{
		conn:            pc,
		routes:          opts.Routes,
		strategyManager: opts.StrategyManager,
		log:             log,
	}

	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()
	go srv.sweep(ctx)

	return srv.serve(ctx)
}

type rawRaknetifyServer struct {
	conn            net.PacketConn
	routes          func() []config.Route
	strategyManager *StrategyManager
	log             logr.Logger
	sessions        sync.Map // map[clientAddr string]*rawRaknetifySession
	sessionCount    atomic.Int64
}

func (s *rawRaknetifyServer) serve(ctx context.Context) error {
	buf := make([]byte, 64*1024)
	for {
		n, clientAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		packet := buf[:n]
		if hint, ok, err := decodeRawRaknetifyRouteHint(packet); ok {
			if err != nil {
				rawRaknetifyMetrics.recordDroppedPacket("client_to_backend", "invalid_route_hint")
				s.log.V(1).Info("dropping invalid raw raknetify route hint", "clientAddr", clientAddr, "error", err)
				continue
			}
			hint.host = cleanRawRaknetifyHost(hint.host)
			if _, err = s.ensureSession(clientAddr, hint); err != nil {
				rawRaknetifyMetrics.recordSessionEvent("rejected", "ensure_failed")
				s.log.Info("failed to create raw raknetify session", "clientAddr", clientAddr, "host", hint.host, "error", err)
			}
			continue
		}

		session, ok := s.loadSession(clientAddr)
		if !ok {
			rawRaknetifyMetrics.recordDroppedPacket("client_to_backend", "no_route_hint")
			s.log.V(1).Info("dropping raw raknetify packet without route hint", "clientAddr", clientAddr)
			continue
		}
		session.touch(time.Now())
		if err = s.writeToBackend(clientAddr, session, packet); err != nil {
			if isRawRaknetifyWriteTimeout(err) {
				rawRaknetifyMetrics.recordDroppedPacket("client_to_backend", "write_timeout")
				rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "timeout")
				s.log.V(1).Info("dropping raw raknetify packet after backend write timed out", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
				continue
			}
			rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "error")
			s.log.V(1).Info("closing raw raknetify session after backend write failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(clientAddr.String(), session, "backend_write_error")
		}
	}
}

func (s *rawRaknetifyServer) loadSession(clientAddr net.Addr) (*rawRaknetifySession, bool) {
	value, ok := s.sessions.Load(clientAddr.String())
	if !ok {
		return nil, false
	}
	session, ok := value.(*rawRaknetifySession)
	return session, ok
}

func (s *rawRaknetifyServer) ensureSession(clientAddr net.Addr, hint rawRaknetifyRouteHint) (*rawRaknetifySession, error) {
	if hint.host == "" {
		return nil, fmt.Errorf("empty route hint host")
	}
	key := clientAddr.String()
	if existing, ok := s.loadSession(clientAddr); ok {
		if existing.sameRouteHint(hint) {
			existing.touch(time.Now())
			return existing, nil
		}
		rawRaknetifyMetrics.recordSessionEvent("replaced", "")
		s.closeSession(key, existing, "replaced")
	}
	if s.sessionCount.Load() >= rawRaknetifyMaxSessions {
		rawRaknetifyMetrics.recordSessionEvent("rejected", "session_limit")
		return nil, fmt.Errorf("raw raknetify session limit reached")
	}

	backendAddr, backendKey, route, log, err := s.resolveBackend(hint.host, clientAddr)
	if err != nil {
		return nil, err
	}
	backendConn, err := net.DialUDP("udp", nil, backendAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial raw raknetify backend %s: %w", backendAddr, err)
	}
	tuneRawRaknetifyUDPConn(log, "backend", backendConn)

	session := &rawRaknetifySession{
		host:        hint.host,
		routeToken:  hint.token,
		hasToken:    hint.hasToken,
		backendAddr: backendAddr,
		backendConn: backendConn,
	}
	if route.Strategy == config.StrategyLeastConnections {
		session.decrement = s.strategyManager.IncrementConnection(backendKey)
	}
	session.touch(time.Now())
	s.sessions.Store(key, session)
	s.sessionCount.Add(1)
	rawRaknetifyMetrics.addActiveSessions(1)
	rawRaknetifyMetrics.recordSessionEvent("created", "")

	log.Info("created raw raknetify session", "clientAddr", clientAddr, "backendAddr", backendAddr)
	go s.copyBackendToClient(key, clientAddr, session)
	return session, nil
}

func (s *rawRaknetifySession) sameRouteHint(hint rawRaknetifyRouteHint) bool {
	if !strings.EqualFold(s.host, hint.host) {
		return false
	}
	if s.hasToken && hint.hasToken {
		return s.routeToken == hint.token
	}
	return false
}

func tuneRawRaknetifyUDPConn(log logr.Logger, name string, conn *net.UDPConn) {
	setRawRaknetifyUDPQoS(log, name, conn, rawRaknetifyDefaultIPTOS)
	if err := conn.SetReadBuffer(rawRaknetifySocketBufferSize); err != nil {
		log.V(1).Info("failed to set raw raknetify UDP read buffer", "socket", name, "bytes", rawRaknetifySocketBufferSize, "error", err)
	}
	if err := conn.SetWriteBuffer(rawRaknetifySocketBufferSize); err != nil {
		log.V(1).Info("failed to set raw raknetify UDP write buffer", "socket", name, "bytes", rawRaknetifySocketBufferSize, "error", err)
	}
}

func setRawRaknetifyUDPQoS(log logr.Logger, name string, conn *net.UDPConn, tos int) {
	if err := ipv4.NewPacketConn(conn).SetTOS(tos); err != nil {
		log.V(1).Info("failed to set raw raknetify IPv4 TOS", "socket", name, "tos", tos, "error", err)
	}
	if err := ipv6.NewPacketConn(conn).SetTrafficClass(tos); err != nil {
		log.V(1).Info("failed to set raw raknetify IPv6 traffic class", "socket", name, "trafficClass", tos, "error", err)
	}
}

func (s *rawRaknetifyServer) writeToBackend(clientAddr net.Addr, session *rawRaknetifySession, packet []byte) error {
	if err := session.backendConn.SetWriteDeadline(time.Now().Add(rawRaknetifyWriteTimeout)); err != nil {
		rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "deadline_error")
		s.log.V(1).Info("failed to set raw raknetify backend write deadline", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "timeout", rawRaknetifyWriteTimeout, "error", err)
	}
	_, err := session.backendConn.Write(packet)
	return err
}

func (s *rawRaknetifyServer) writeToClient(clientAddr net.Addr, session *rawRaknetifySession, packet []byte) error {
	if err := s.conn.SetWriteDeadline(time.Now().Add(rawRaknetifyWriteTimeout)); err != nil {
		rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "deadline_error")
		s.log.V(1).Info("failed to set raw raknetify client write deadline", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "timeout", rawRaknetifyWriteTimeout, "error", err)
	}
	_, err := s.conn.WriteTo(packet, clientAddr)
	return err
}

func isRawRaknetifyWriteTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func (s *rawRaknetifyServer) resolveBackend(host string, clientAddr net.Addr) (*net.UDPAddr, string, *config.Route, logr.Logger, error) {
	log := s.log.WithValues("clientAddr", clientAddr, "virtualHost", host)
	matchedHost, route, groups := FindRouteWithGroups(host, s.routes()...)
	if route == nil {
		return nil, "", nil, log.V(1), fmt.Errorf("no route configured for host %s", host)
	}
	log = log.WithValues("route", matchedHost)
	if !route.Raknetify.Enabled || route.RaknetifyMode() != config.RaknetifyModeRawPassthrough {
		return nil, "", route, log, fmt.Errorf("route %s is not configured for raw raknetify passthrough", matchedHost)
	}
	if len(route.Backend) == 0 {
		return nil, "", route, log, fmt.Errorf("no backend configured for route %s", matchedHost)
	}

	tryBackendList := route.Backend.Copy()
	nextBackend := func() (string, logr.Logger, bool) {
		if len(tryBackendList) == 0 {
			return "", log, false
		}
		backendAddr, newLog, ok := s.strategyManager.GetNextBackend(log, route, matchedHost, tryBackendList)
		if !ok {
			return "", log, false
		}
		if len(groups) > 0 {
			backendAddr = SubstituteBackendParams(backendAddr, groups)
		}
		for i, backend := range tryBackendList {
			original := backend
			if len(groups) > 0 {
				original = SubstituteBackendParams(backend, groups)
			}
			if rawRaknetifyBackendEqual(original, backendAddr) {
				tryBackendList = append(tryBackendList[:i], tryBackendList[i+1:]...)
				break
			}
		}
		return backendAddr, newLog.WithValues("backendAddr", backendAddr), true
	}

	backendKey, log, backendAddr, err := tryBackends(nextBackend, func(log logr.Logger, backendKey string) (logr.Logger, *net.UDPAddr, error) {
		addr, err := resolveRawRaknetifyBackend(backendKey)
		return log, addr, err
	})
	if err != nil {
		return nil, "", route, log, err
	}
	return backendAddr, backendKey, route, log, nil
}

func (s *rawRaknetifyServer) copyBackendToClient(key string, clientAddr net.Addr, session *rawRaknetifySession) {
	buf := make([]byte, 64*1024)
	for {
		n, err := session.backendConn.Read(buf)
		if err != nil {
			s.log.V(1).Info("closing raw raknetify session after backend read failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(key, session, "backend_read_error")
			return
		}
		session.touch(time.Now())
		if err = s.writeToClient(clientAddr, session, buf[:n]); err != nil {
			if isRawRaknetifyWriteTimeout(err) {
				rawRaknetifyMetrics.recordDroppedPacket("backend_to_client", "write_timeout")
				rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "timeout")
				s.log.V(1).Info("dropping raw raknetify packet after client write timed out", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
				continue
			}
			rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "error")
			s.log.V(1).Info("closing raw raknetify session after client write failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(key, session, "client_write_error")
			return
		}
	}
}

func (s *rawRaknetifyServer) closeSession(key string, expected *rawRaknetifySession, reason string) {
	var session *rawRaknetifySession
	if expected != nil {
		if !s.sessions.CompareAndDelete(key, expected) {
			return
		}
		session = expected
	} else {
		value, ok := s.sessions.LoadAndDelete(key)
		if !ok {
			return
		}
		var valid bool
		session, valid = value.(*rawRaknetifySession)
		if !valid {
			return
		}
	}
	session.close()
	s.sessionCount.Add(-1)
	rawRaknetifyMetrics.addActiveSessions(-1)
	rawRaknetifyMetrics.recordSessionEvent("closed", reason)
}

func (s *rawRaknetifyServer) closeAllSessions() {
	s.sessions.Range(func(key, value any) bool {
		keyStr, keyOK := key.(string)
		session, sessionOK := value.(*rawRaknetifySession)
		if keyOK && sessionOK {
			s.closeSession(keyStr, session, "shutdown")
		}
		return true
	})
}

func (s *rawRaknetifyServer) closeIdleSessions(cutoff int64) {
	s.sessions.Range(func(key, value any) bool {
		session, sessionOK := value.(*rawRaknetifySession)
		keyStr, keyOK := key.(string)
		if sessionOK && keyOK && session.lastSeen.Load() < cutoff {
			s.closeSession(keyStr, session, "idle_timeout")
		}
		return true
	})
}

func (s *rawRaknetifyServer) sweep(ctx context.Context) {
	ticker := time.NewTicker(rawRaknetifySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.closeAllSessions()
			return
		case now := <-ticker.C:
			s.closeIdleSessions(now.Add(-rawRaknetifyIdleTimeout).UnixNano())
		}
	}
}

func decodeRawRaknetifyRouteHint(packet []byte) (hint rawRaknetifyRouteHint, ok bool, err error) {
	if len(packet) == 0 || packet[0] != rawRaknetifyRouteHintPacketID {
		return rawRaknetifyRouteHint{}, false, nil
	}
	offset := 1
	if len(packet) < offset+len(rawRaknetifyRouteHintMagic)+1 {
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("route hint packet is too short")
	}
	if !bytes.Equal(packet[offset:offset+len(rawRaknetifyRouteHintMagic)], rawRaknetifyRouteHintMagic) {
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("route hint magic mismatch")
	}
	offset += len(rawRaknetifyRouteHintMagic)
	version := packet[offset]
	offset++
	switch version {
	case rawRaknetifyRouteHintVersion:
	case rawRaknetifyRouteHintVersion2:
		if len(packet) < offset+rawRaknetifyRouteTokenLen {
			return rawRaknetifyRouteHint{}, true, fmt.Errorf("route hint packet is too short")
		}
		hint.token = string(packet[offset : offset+rawRaknetifyRouteTokenLen])
		hint.hasToken = true
		offset += rawRaknetifyRouteTokenLen
	default:
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("unsupported route hint version %d", version)
	}
	if len(packet) < offset+2 {
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("route hint packet is too short")
	}
	hostLen := int(binary.BigEndian.Uint16(packet[offset : offset+2]))
	offset += 2
	if hostLen == 0 || hostLen > rawRaknetifyMaxHintHostLen {
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("invalid route hint host length %d", hostLen)
	}
	if len(packet) != offset+hostLen {
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("route hint length mismatch")
	}
	hostBytes := packet[offset:]
	if !utf8.Valid(hostBytes) {
		return rawRaknetifyRouteHint{}, true, fmt.Errorf("route hint host is not valid UTF-8")
	}
	hint.host = string(hostBytes)
	return hint, true, nil
}

func cleanRawRaknetifyHost(host string) string {
	host = strings.TrimSpace(ClearVirtualHost(host))
	if hostOnly := netutil.HostStr(host); hostOnly != "" {
		host = hostOnly
	}
	return strings.Trim(host, ".")
}

func resolveRawRaknetifyBackend(backendAddr string) (*net.UDPAddr, error) {
	normalized, err := normalizeRawRaknetifyBackend(backendAddr)
	if err != nil {
		return nil, err
	}
	return net.ResolveUDPAddr("udp", normalized)
}

func normalizeRawRaknetifyBackend(backendAddr string) (string, error) {
	addr, err := netutil.Parse(backendAddr, "udp")
	if err != nil {
		return "", err
	}
	normalized := addr.String()
	if _, port := netutil.HostPort(addr); port == 0 {
		normalized = net.JoinHostPort(addr.String(), "25565")
	}
	return normalized, nil
}

func rawRaknetifyBackendEqual(a, b string) bool {
	normalizedA, errA := normalizeRawRaknetifyBackend(a)
	normalizedB, errB := normalizeRawRaknetifyBackend(b)
	return errA == nil && errB == nil && normalizedA == normalizedB
}
