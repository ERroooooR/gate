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
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	rawRaknetifyRouteHintPacketID = byte(0xfe)
	rawRaknetifyRouteHintVersion  = byte(1)
	rawRaknetifyRouteHintVersion2 = byte(2)
	rawRaknetifyRouteTokenLen     = 16
	rawRaknetifyMaxHintHostLen    = 1024
	// rawRaknetifyDefaultIdleTimeout is the default idle timeout for raw sessions
	// when not configured in the route. It is set to 2x the server-side RakNet read
	// timeout (15s) to give enough margin for ping/pong exchanges while cleaning up
	// stale sessions promptly.
	rawRaknetifyDefaultIdleTimeout = 30 * time.Second
	rawRaknetifySweepInterval     = 15 * time.Second
	rawRaknetifyMaxSessions       = 4096
	rawRaknetifySocketBufferSize  = 4 * 1024 * 1024
	rawRaknetifyDefaultIPTOS      = 0xA0
	// rawRaknetifyDefaultReadDeadline is the maximum time to wait for a backend read.
	// This prevents copyBackendToClient from hanging indefinitely on a dead backend.
	// Set to sweep interval + 5s so normal pings (every 200ms) won't trigger it.
	rawRaknetifyDefaultReadDeadline = 20 * time.Second
)

var rawRaknetifyRouteHintMagic = []byte("GATE_RAKNET_ROUTE")

type rawRaknetifySession struct {
	host        string
	routeToken  string
	hasToken    bool
	tokenKey    string
	backendConn *net.UDPConn
	decrement   func()
	lastSeen    atomic.Int64
	closeOnce   sync.Once
	mu          sync.Mutex
	clientAddr  net.Addr
	clientKey   string
	// idleTimeout is the per-session idle timeout derived from the route config.
	// If the route has no configured IdleTimeout, rawRaknetifyDefaultIdleTimeout is used.
	idleTimeout time.Duration
	// writeDeadline is the per-session write deadline derived from the route config.
	// 0 means no deadline.
	writeDeadline time.Duration
}

func (s *rawRaknetifySession) touch(now time.Time) {
	s.lastSeen.Store(now.UnixNano())
}

func (s *rawRaknetifySession) close() bool {
	closed := false
	s.closeOnce.Do(func() {
		closed = true
		_ = s.backendConn.Close()
		if s.decrement != nil {
			s.decrement()
		}
	})
	return closed
}

func (s *rawRaknetifySession) setClientAddr(clientAddr net.Addr) (oldKey, newKey string, changed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newKey = clientAddr.String()
	if s.clientKey == newKey {
		return s.clientKey, newKey, false
	}
	oldKey = s.clientKey
	s.clientAddr = clientAddr
	s.clientKey = newKey
	return oldKey, newKey, true
}

func (s *rawRaknetifySession) currentClientAddr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clientAddr
}

func (s *rawRaknetifySession) currentClientKey() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clientKey
}

type rawRaknetifyRouteHint struct {
	host     string
	token    string
	hasToken bool
}

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
		tuneRawRaknetifyUDPConn(log, "listener", udpConn, rawRaknetifyDefaultIPTOS)
	}

	srv := &rawRaknetifyServer{
		conn:            pc,
		routes:          opts.Routes,
		strategyManager: opts.StrategyManager,
		log:             log,
		clientTOS:       -1, // force TOS set on first write
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
	sessions        sync.Map
	tokenSessions   sync.Map
	sessionCount    atomic.Int64
	tosMu           sync.Mutex
	clientTOS       int
	migrationMu     sync.Mutex
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

		if session.writeDeadline > 0 {
			_ = session.backendConn.SetWriteDeadline(time.Now().Add(session.writeDeadline))
		}

		if _, err := session.backendConn.Write(packet); err != nil {
			rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "write_error")
			s.log.V(1).Info("raw raknetify backend write failed, closing session", "clientAddr", clientAddr, "backendAddr", session.backendConn.RemoteAddr(), "error", err)
			s.closeSession(session, "backend_write_error")
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
	tokenKey := rawRaknetifyTokenKey(hint)
	if tokenKey != "" {
		if value, ok := s.tokenSessions.Load(tokenKey); ok {
			if existing, ok := value.(*rawRaknetifySession); ok && strings.EqualFold(existing.host, hint.host) {
				if current, ok := s.loadSession(clientAddr); ok && current != existing {
					rawRaknetifyMetrics.recordSessionEvent("replaced", "migration_conflict")
					s.closeSession(current, "migration_conflict")
				}
				if oldKey, newKey, migrated := s.migrateSessionClient(existing, clientAddr); migrated {
					rawRaknetifyMetrics.recordSessionEvent("migrated", "")
					s.log.Info("migrated raw raknetify session", "oldClientAddr", oldKey, "clientAddr", newKey, "backendAddr", existing.backendConn.RemoteAddr())
				}
				existing.touch(time.Now())
				return existing, nil
			}
		}
	}
	if existing, ok := s.loadSession(clientAddr); ok {
		if strings.EqualFold(existing.host, hint.host) {
			if existing.hasToken && hint.hasToken && existing.routeToken != hint.token {
			} else {
				existing.touch(time.Now())
				return existing, nil
			}
		}
		rawRaknetifyMetrics.recordSessionEvent("replaced", "")
		s.closeSession(existing, "replaced")
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
	tuneRawRaknetifyUDPConn(log, "backend", backendConn, rawRaknetifyQOSTOSForRoute(route))

	// Determine per-session timeouts from route config, with sensible defaults.
	sessionIdleTimeout := rawRaknetifyDefaultIdleTimeout
	sessionWriteDeadline := time.Duration(0) // no deadline by default
	raw := route.Raknetify.RawPassthrough
	if raw.IdleTimeout > 0 {
		sessionIdleTimeout = time.Duration(raw.IdleTimeout)
	}
	if raw.WriteTimeout > 0 {
		sessionWriteDeadline = time.Duration(raw.WriteTimeout)
	}

	session := &rawRaknetifySession{
		host:           hint.host,
		routeToken:     hint.token,
		hasToken:       hint.hasToken,
		tokenKey:       tokenKey,
		backendConn:    backendConn,
		clientAddr:     clientAddr,
		clientKey:      key,
		idleTimeout:    sessionIdleTimeout,
		writeDeadline:  sessionWriteDeadline,
	}
	if route.Strategy == config.StrategyLeastConnections {
		session.decrement = s.strategyManager.IncrementConnection(backendKey)
	}
	session.touch(time.Now())
	s.sessions.Store(key, session)
	if tokenKey != "" {
		s.tokenSessions.Store(tokenKey, session)
	}
	s.sessionCount.Add(1)
	rawRaknetifyMetrics.addActiveSessions(1)
	rawRaknetifyMetrics.recordSessionEvent("created", "")

	log.Info("created raw raknetify session", "clientAddr", clientAddr, "backendAddr", backendAddr)
	go s.copyBackendToClient(session)
	return session, nil
}

func (s *rawRaknetifyServer) migrateSessionClient(session *rawRaknetifySession, clientAddr net.Addr) (oldKey, newKey string, migrated bool) {
	s.migrationMu.Lock()
	defer s.migrationMu.Unlock()

	oldKey, newKey, migrated = session.setClientAddr(clientAddr)
	if !migrated {
		return oldKey, newKey, false
	}
	if value, ok := s.sessions.Load(newKey); ok {
		if existing, ok := value.(*rawRaknetifySession); ok && existing != session {
			s.sessions.CompareAndDelete(newKey, existing)
			s.finalizeSessionClose(existing, "migration_conflict")
		}
	}
	s.sessions.Store(newKey, session)
	if oldKey != "" {
		s.sessions.CompareAndDelete(oldKey, session)
	}
	return oldKey, newKey, true
}

func (s *rawRaknetifyServer) copyBackendToClient(session *rawRaknetifySession) {
	buf := make([]byte, 64*1024)
	for {
		// Apply a read deadline to prevent hanging on a dead backend.
		// The deadline is set to rawRaknetifyDefaultReadDeadline, which is longer than
		// the server-side RakNet ping interval (200ms) so normal traffic won't trigger it,
		// but short enough to detect a truly dead backend within one sweep cycle.
		_ = session.backendConn.SetReadDeadline(time.Now().Add(rawRaknetifyDefaultReadDeadline))

		n, err := session.backendConn.Read(buf)
		if err != nil {
			clientAddr := session.currentClientAddr()
			s.log.V(1).Info("closing raw raknetify session after backend read failed", "clientAddr", clientAddr, "backendAddr", session.backendConn.RemoteAddr(), "error", err)
			s.closeSession(session, "backend_read_error")
			return
		}
		session.touch(time.Now())

		clientAddr := session.currentClientAddr()
		if clientAddr == nil {
			s.closeSession(session, "no_client_addr")
			return
		}
		if _, err = s.writeToClient(clientAddr, buf[:n]); err != nil {
			rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "write_error")
			s.log.V(1).Info("closing raw raknetify session after client write failed", "clientAddr", clientAddr, "backendAddr", session.backendConn.RemoteAddr(), "error", err)
			s.closeSession(session, "client_write_error")
			return
		}
	}
}

func (s *rawRaknetifyServer) writeToClient(clientAddr net.Addr, packet []byte) (int, error) {
	if udpConn, ok := s.conn.(*net.UDPConn); ok {
		// clientTOS is initialized to -1, so the first call always sets TOS.
		// Subsequent calls only update TOS if it has changed.
		s.tosMu.Lock()
		if s.clientTOS != rawRaknetifyDefaultIPTOS {
			setRawRaknetifyUDPQoS(s.log, "listener", udpConn, rawRaknetifyDefaultIPTOS)
			s.clientTOS = rawRaknetifyDefaultIPTOS
		}
		s.tosMu.Unlock()
	}
	return s.conn.WriteTo(packet, clientAddr)
}

func (s *rawRaknetifyServer) closeSession(session *rawRaknetifySession, reason string) {
	if session == nil {
		return
	}
	s.migrationMu.Lock()
	key := session.currentClientKey()
	if key != "" {
		s.sessions.CompareAndDelete(key, session)
	}
	if session.tokenKey != "" {
		s.tokenSessions.CompareAndDelete(session.tokenKey, session)
	}
	s.migrationMu.Unlock()

	s.finalizeSessionClose(session, reason)
}

func (s *rawRaknetifyServer) finalizeSessionClose(session *rawRaknetifySession, reason string) {
	if session == nil {
		return
	}
	if session.tokenKey != "" {
		s.tokenSessions.CompareAndDelete(session.tokenKey, session)
	}
	if session.close() {
		s.sessionCount.Add(-1)
		rawRaknetifyMetrics.addActiveSessions(-1)
		rawRaknetifyMetrics.recordSessionEvent("closed", reason)
	}
}

func (s *rawRaknetifyServer) closeAllSessions() {
	s.sessions.Range(func(key, value any) bool {
		_, keyOK := key.(string)
		session, sessionOK := value.(*rawRaknetifySession)
		if keyOK && sessionOK {
			s.closeSession(session, "shutdown")
		}
		return true
	})
}

func (s *rawRaknetifyServer) closeIdleSessions(now time.Time) {
	s.sessions.Range(func(key, value any) bool {
		session, sessionOK := value.(*rawRaknetifySession)
		_, keyOK := key.(string)
		if sessionOK && keyOK {
			// Use per-session idle timeout, falling back to the default.
			idleTimeout := session.idleTimeout
			if idleTimeout <= 0 {
				idleTimeout = rawRaknetifyDefaultIdleTimeout
			}
			if session.lastSeen.Load() < now.Add(-idleTimeout).UnixNano() {
				s.log.V(1).Info("closing idle raw raknetify session",
					"clientAddr", session.currentClientAddr(),
					"backendAddr", session.backendConn.RemoteAddr(),
					"idleTimeout", idleTimeout,
					"host", session.host)
				s.closeSession(session, "idle_timeout")
			}
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
			s.closeIdleSessions(now)
		}
	}
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

func tuneRawRaknetifyUDPConn(log logr.Logger, name string, conn *net.UDPConn, tos int) {
	setRawRaknetifyUDPQoS(log, name, conn, tos)
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

func rawRaknetifyTokenKey(hint rawRaknetifyRouteHint) string {
	if !hint.hasToken {
		return ""
	}
	return strings.ToLower(hint.host) + "\x00" + hint.token
}

func rawRaknetifyQOSTOSForRoute(route *config.Route) int {
	raw := route.Raknetify.RawPassthrough
	switch raw.QOS.Mode {
	case config.RaknetifyQOSModeClear:
		return 0
	case config.RaknetifyQOSModeCustom:
		if raw.QOS.TOS != nil {
			return *raw.QOS.TOS
		}
	}
	return rawRaknetifyDefaultIPTOS
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
