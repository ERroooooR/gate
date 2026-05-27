package lite

import (
	"bytes"
	"context"
	"encoding/binary"
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
)

const (
	rawRaknetifyRouteHintPacketID = byte(0xfe)
	rawRaknetifyRouteHintVersion  = byte(1)
	rawRaknetifyMaxHintHostLen    = 1024
	rawRaknetifyIdleTimeout       = time.Minute
	rawRaknetifySweepInterval     = 15 * time.Second
	rawRaknetifyMaxSessions       = 4096
)

var rawRaknetifyRouteHintMagic = []byte("GATE_RAKNET_ROUTE")

type rawRaknetifySession struct {
	host        string
	backendAddr *net.UDPAddr
	backendConn *net.UDPConn
	decrement   func()
	lastSeen    atomic.Int64
	closeOnce   sync.Once
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
		if host, ok, err := decodeRawRaknetifyRouteHint(packet); ok {
			if err != nil {
				s.log.V(1).Info("dropping invalid raw raknetify route hint", "clientAddr", clientAddr, "error", err)
				continue
			}
			if _, err = s.ensureSession(clientAddr, cleanRawRaknetifyHost(host)); err != nil {
				s.log.Info("failed to create raw raknetify session", "clientAddr", clientAddr, "host", host, "error", err)
			}
			continue
		}

		session, ok := s.loadSession(clientAddr)
		if !ok {
			s.log.V(1).Info("dropping raw raknetify packet without route hint", "clientAddr", clientAddr)
			continue
		}
		session.touch(time.Now())
		if _, err = session.backendConn.Write(packet); err != nil {
			s.log.V(1).Info("closing raw raknetify session after backend write failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(clientAddr.String())
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

func (s *rawRaknetifyServer) ensureSession(clientAddr net.Addr, host string) (*rawRaknetifySession, error) {
	if host == "" {
		return nil, fmt.Errorf("empty route hint host")
	}
	if existing, ok := s.loadSession(clientAddr); ok {
		if strings.EqualFold(existing.host, host) {
			existing.touch(time.Now())
			return existing, nil
		}
		s.closeSession(clientAddr.String())
	}
	if s.sessionCount.Load() >= rawRaknetifyMaxSessions {
		return nil, fmt.Errorf("raw raknetify session limit reached")
	}

	backendAddr, backendKey, route, log, err := s.resolveBackend(host, clientAddr)
	if err != nil {
		return nil, err
	}
	backendConn, err := net.DialUDP("udp", nil, backendAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial raw raknetify backend %s: %w", backendAddr, err)
	}

	session := &rawRaknetifySession{
		host:        host,
		backendAddr: backendAddr,
		backendConn: backendConn,
	}
	if route.Strategy == config.StrategyLeastConnections {
		session.decrement = s.strategyManager.IncrementConnection(backendKey)
	}
	session.touch(time.Now())
	s.sessions.Store(clientAddr.String(), session)
	s.sessionCount.Add(1)

	log.Info("created raw raknetify session", "clientAddr", clientAddr, "backendAddr", backendAddr)
	go s.copyBackendToClient(clientAddr.String(), clientAddr, session)
	return session, nil
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
			s.closeSession(key)
			return
		}
		session.touch(time.Now())
		if _, err = s.conn.WriteTo(buf[:n], clientAddr); err != nil {
			s.log.V(1).Info("closing raw raknetify session after client write failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(key)
			return
		}
	}
}

func (s *rawRaknetifyServer) closeSession(key string) {
	value, ok := s.sessions.LoadAndDelete(key)
	if !ok {
		return
	}
	if session, ok := value.(*rawRaknetifySession); ok {
		session.close()
		s.sessionCount.Add(-1)
	}
}

func (s *rawRaknetifyServer) sweep(ctx context.Context) {
	ticker := time.NewTicker(rawRaknetifySweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.sessions.Range(func(key, _ any) bool {
				if keyStr, ok := key.(string); ok {
					s.closeSession(keyStr)
				}
				return true
			})
			return
		case now := <-ticker.C:
			cutoff := now.Add(-rawRaknetifyIdleTimeout).UnixNano()
			s.sessions.Range(func(key, value any) bool {
				session, ok := value.(*rawRaknetifySession)
				keyStr, keyOK := key.(string)
				if ok && keyOK && session.lastSeen.Load() < cutoff {
					s.closeSession(keyStr)
				}
				return true
			})
		}
	}
}

func decodeRawRaknetifyRouteHint(packet []byte) (host string, ok bool, err error) {
	if len(packet) == 0 || packet[0] != rawRaknetifyRouteHintPacketID {
		return "", false, nil
	}
	offset := 1
	if len(packet) < offset+len(rawRaknetifyRouteHintMagic)+1+2 {
		return "", true, fmt.Errorf("route hint packet is too short")
	}
	if !bytes.Equal(packet[offset:offset+len(rawRaknetifyRouteHintMagic)], rawRaknetifyRouteHintMagic) {
		return "", true, fmt.Errorf("route hint magic mismatch")
	}
	offset += len(rawRaknetifyRouteHintMagic)
	if packet[offset] != rawRaknetifyRouteHintVersion {
		return "", true, fmt.Errorf("unsupported route hint version %d", packet[offset])
	}
	offset++
	hostLen := int(binary.BigEndian.Uint16(packet[offset : offset+2]))
	offset += 2
	if hostLen == 0 || hostLen > rawRaknetifyMaxHintHostLen {
		return "", true, fmt.Errorf("invalid route hint host length %d", hostLen)
	}
	if len(packet) != offset+hostLen {
		return "", true, fmt.Errorf("route hint length mismatch")
	}
	hostBytes := packet[offset:]
	if !utf8.Valid(hostBytes) {
		return "", true, fmt.Errorf("route hint host is not valid UTF-8")
	}
	return string(hostBytes), true, nil
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
