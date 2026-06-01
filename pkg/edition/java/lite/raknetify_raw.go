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
	rawRaknetifyPacingInterval    = 100 * time.Microsecond
	rawRaknetifyBackendQueueSize  = 256
)

var rawRaknetifyRouteHintMagic = []byte("GATE_RAKNET_ROUTE")

type rawRaknetifySession struct {
	host        string
	routeToken  string
	hasToken    bool
	tokenKey    string
	backendAddr *net.UDPAddr
	backendConn *net.UDPConn
	options     rawRaknetifyRouteOptions
	toBackend   chan []byte
	decrement   func()
	lastSeen    atomic.Int64
	closeOnce   sync.Once
	mu          sync.RWMutex
	clientAddr  net.Addr
	clientKey   string
	queueMu     sync.RWMutex
	closed      bool
}

type rawRaknetifyRouteHint struct {
	host     string
	token    string
	hasToken bool
}

type rawRaknetifyRouteOptions struct {
	qosTOS         int
	idleTimeout    time.Duration
	writeTimeout   time.Duration
	pacingInterval time.Duration
	queueSize      int
}

func (s *rawRaknetifySession) touch(now time.Time) {
	s.lastSeen.Store(now.UnixNano())
}

func (s *rawRaknetifySession) close() bool {
	closed := false
	s.closeOnce.Do(func() {
		closed = true
		s.queueMu.Lock()
		s.closed = true
		if s.toBackend != nil {
			close(s.toBackend)
		}
		s.queueMu.Unlock()
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientAddr
}

func (s *rawRaknetifySession) currentClientKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientKey
}

func (s *rawRaknetifySession) enqueueToBackend(packet []byte) bool {
	copied := make([]byte, len(packet))
	copy(copied, packet)
	s.queueMu.RLock()
	defer s.queueMu.RUnlock()
	if s.closed {
		return false
	}
	select {
	case s.toBackend <- copied:
		return true
	default:
		return false
	}
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
		tuneRawRaknetifyUDPConn(log, "listener", udpConn, rawRaknetifyDefaultIPTOS)
	}

	srv := &rawRaknetifyServer{
		conn:            pc,
		routes:          opts.Routes,
		strategyManager: opts.StrategyManager,
		log:             log,
		clientTOS:       rawRaknetifyDefaultIPTOS,
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
	tokenSessions   sync.Map // map[host+token string]*rawRaknetifySession
	sessionCount    atomic.Int64
	clientWriteMu   sync.Mutex
	clientTOS       int
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
		if !session.enqueueToBackend(packet) {
			rawRaknetifyMetrics.recordDroppedPacket("client_to_backend", "backend_queue_full")
			s.log.V(1).Info("dropping raw raknetify packet because backend queue is full", "clientAddr", clientAddr, "backendAddr", session.backendAddr)
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
			if existing, ok := value.(*rawRaknetifySession); ok && existing.sameRouteHint(hint) {
				if current, ok := s.loadSession(clientAddr); ok && current != existing {
					rawRaknetifyMetrics.recordSessionEvent("replaced", "migration_conflict")
					s.closeSession(current, "migration_conflict")
				}
				if oldKey, newKey, migrated := s.migrateSessionClient(existing, clientAddr); migrated {
					rawRaknetifyMetrics.recordSessionEvent("migrated", "")
					s.log.Info("migrated raw raknetify session", "oldClientAddr", oldKey, "clientAddr", newKey, "backendAddr", existing.backendAddr)
				}
				existing.touch(time.Now())
				return existing, nil
			}
		}
	}
	if existing, ok := s.loadSession(clientAddr); ok {
		if existing.sameRouteHint(hint) {
			existing.touch(time.Now())
			return existing, nil
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
	options := rawRaknetifyOptionsForRoute(route)
	tuneRawRaknetifyUDPConn(log, "backend", backendConn, options.qosTOS)

	session := &rawRaknetifySession{
		host:        hint.host,
		routeToken:  hint.token,
		hasToken:    hint.hasToken,
		tokenKey:    tokenKey,
		backendAddr: backendAddr,
		backendConn: backendConn,
		options:     options,
		toBackend:   make(chan []byte, options.queueSize),
		clientAddr:  clientAddr,
		clientKey:   key,
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
	go s.copyClientToBackend(session)
	go s.copyBackendToClient(session)
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

func (s *rawRaknetifyServer) migrateSessionClient(session *rawRaknetifySession, clientAddr net.Addr) (oldKey, newKey string, migrated bool) {
	oldKey, newKey, migrated = session.setClientAddr(clientAddr)
	if !migrated {
		return oldKey, newKey, false
	}
	if value, ok := s.sessions.Load(newKey); ok {
		if existing, ok := value.(*rawRaknetifySession); ok && existing != session {
			s.closeSession(existing, "migration_conflict")
		}
	}
	s.sessions.Store(newKey, session)
	if oldKey != "" {
		s.sessions.CompareAndDelete(oldKey, session)
	}
	return oldKey, newKey, true
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

func (s *rawRaknetifyServer) writeToBackend(session *rawRaknetifySession, packet []byte) error {
	clientAddr := session.currentClientAddr()
	if err := session.backendConn.SetWriteDeadline(time.Now().Add(session.options.writeTimeout)); err != nil {
		rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "deadline_error")
		s.log.V(1).Info("failed to set raw raknetify backend write deadline", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "timeout", session.options.writeTimeout, "error", err)
	}
	_, err := session.backendConn.Write(packet)
	return err
}

func (s *rawRaknetifyServer) writeToClient(session *rawRaknetifySession, packet []byte) error {
	clientAddr := session.currentClientAddr()
	if clientAddr == nil {
		return net.ErrClosed
	}
	s.clientWriteMu.Lock()
	defer s.clientWriteMu.Unlock()
	if udpConn, ok := s.conn.(*net.UDPConn); ok && s.clientTOS != session.options.qosTOS {
		setRawRaknetifyUDPQoS(s.log, "listener", udpConn, session.options.qosTOS)
		s.clientTOS = session.options.qosTOS
	}
	if err := s.conn.SetWriteDeadline(time.Now().Add(session.options.writeTimeout)); err != nil {
		rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "deadline_error")
		s.log.V(1).Info("failed to set raw raknetify client write deadline", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "timeout", session.options.writeTimeout, "error", err)
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

func (s *rawRaknetifyServer) copyClientToBackend(session *rawRaknetifySession) {
	var nextWrite time.Time
	for packet := range session.toBackend {
		nextWrite = paceRawRaknetifyWrite(nextWrite, session.options.pacingInterval)
		session.touch(time.Now())
		if err := s.writeToBackend(session, packet); err != nil {
			clientAddr := session.currentClientAddr()
			if isRawRaknetifyWriteTimeout(err) {
				rawRaknetifyMetrics.recordDroppedPacket("client_to_backend", "write_timeout")
				rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "timeout")
				s.log.V(1).Info("dropping raw raknetify packet after backend write timed out", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
				continue
			}
			rawRaknetifyMetrics.recordWriteFailure("client_to_backend", "error")
			s.log.V(1).Info("closing raw raknetify session after backend write failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(session, "backend_write_error")
			return
		}
	}
}

func (s *rawRaknetifyServer) copyBackendToClient(session *rawRaknetifySession) {
	buf := make([]byte, 64*1024)
	var nextWrite time.Time
	for {
		n, err := session.backendConn.Read(buf)
		if err != nil {
			clientAddr := session.currentClientAddr()
			s.log.V(1).Info("closing raw raknetify session after backend read failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(session, "backend_read_error")
			return
		}
		nextWrite = paceRawRaknetifyWrite(nextWrite, session.options.pacingInterval)
		session.touch(time.Now())
		if err = s.writeToClient(session, buf[:n]); err != nil {
			clientAddr := session.currentClientAddr()
			if isRawRaknetifyWriteTimeout(err) {
				rawRaknetifyMetrics.recordDroppedPacket("backend_to_client", "write_timeout")
				rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "timeout")
				s.log.V(1).Info("dropping raw raknetify packet after client write timed out", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
				continue
			}
			rawRaknetifyMetrics.recordWriteFailure("backend_to_client", "error")
			s.log.V(1).Info("closing raw raknetify session after client write failed", "clientAddr", clientAddr, "backendAddr", session.backendAddr, "error", err)
			s.closeSession(session, "client_write_error")
			return
		}
	}
}

func paceRawRaknetifyWrite(nextWrite time.Time, interval time.Duration) time.Time {
	if interval <= 0 {
		return time.Time{}
	}
	now := time.Now()
	if now.Before(nextWrite) {
		time.Sleep(nextWrite.Sub(now))
		now = nextWrite
	}
	return now.Add(interval)
}

func (s *rawRaknetifyServer) closeSession(session *rawRaknetifySession, reason string) {
	if session == nil {
		return
	}
	key := session.currentClientKey()
	if key != "" && !s.sessions.CompareAndDelete(key, session) {
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
		if sessionOK && keyOK && session.lastSeen.Load() < now.Add(-session.options.idleTimeout).UnixNano() {
			s.closeSession(session, "idle_timeout")
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

func rawRaknetifyOptionsForRoute(route *config.Route) rawRaknetifyRouteOptions {
	raw := route.Raknetify.RawPassthrough
	options := rawRaknetifyRouteOptions{
		qosTOS:         rawRaknetifyDefaultIPTOS,
		idleTimeout:    rawRaknetifyIdleTimeout,
		writeTimeout:   rawRaknetifyWriteTimeout,
		pacingInterval: rawRaknetifyPacingInterval,
		queueSize:      rawRaknetifyBackendQueueSize,
	}
	switch raw.QOS.Mode {
	case config.RaknetifyQOSModeClear:
		options.qosTOS = 0
	case config.RaknetifyQOSModeCustom:
		if raw.QOS.TOS != nil {
			options.qosTOS = *raw.QOS.TOS
		}
	}
	if raw.IdleTimeout > 0 {
		options.idleTimeout = time.Duration(raw.IdleTimeout)
	}
	if raw.WriteTimeout > 0 {
		options.writeTimeout = time.Duration(raw.WriteTimeout)
	}
	if raw.PacingInterval > 0 {
		options.pacingInterval = time.Duration(raw.PacingInterval)
	}
	if raw.QueueSize > 0 {
		options.queueSize = raw.QueueSize
	}
	return options
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
