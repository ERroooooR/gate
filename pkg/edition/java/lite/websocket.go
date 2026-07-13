package lite

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
	"go.minekube.com/gate/pkg/internal/tcpbrutal"
	"go.minekube.com/gate/pkg/util/netutil"
)

type WebSocketOptions struct {
	Config           func() config.Config
	StrategyManager  *StrategyManager
	HandleConn       func(net.Conn)
	ClientTCPBrutal  func() tcpbrutal.Options
	BackendTCPBrutal func() tcpbrutal.Options
	Logger           logr.Logger
}

type webSocketServer struct {
	ctx       context.Context
	opts      WebSocketOptions
	trusted   []netip.Prefix
	quotaMu   sync.Mutex
	activeIPs map[string]int
}

// ServeWebSocket serves WSMC-compatible binary WebSocket streams on a dedicated TCP listener.
func ServeWebSocket(ctx context.Context, opts WebSocketOptions) error {
	if opts.Config == nil || opts.HandleConn == nil {
		return errors.New("websocket config provider and connection handler are required")
	}
	if opts.StrategyManager == nil {
		opts.StrategyManager = NewStrategyManager()
	}
	if opts.ClientTCPBrutal == nil {
		opts.ClientTCPBrutal = func() tcpbrutal.Options { return tcpbrutal.Options{} }
	}
	if opts.BackendTCPBrutal == nil {
		opts.BackendTCPBrutal = func() tcpbrutal.Options { return tcpbrutal.Options{} }
	}

	cfg := opts.Config().WebSocket
	trusted, err := parseTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return err
	}
	s := &webSocketServer{ctx: ctx, opts: opts, trusted: trusted, activeIPs: make(map[string]int)}

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", cfg.Bind)
	if err != nil {
		return err
	}
	defer ln.Close()
	ln = &webSocketListener{Listener: ln, apply: opts.ClientTCPBrutal, cfg: cfg, log: opts.Logger}

	httpServer := &http.Server{
		Handler:           s,
		ReadHeaderTimeout: cfg.EffectiveHandshakeTimeout(),
		MaxHeaderBytes:    32 << 10,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	opts.Logger.WithName("lite").WithName("websocket").Info("websocket listener started", "bind", ln.Addr())
	err = httpServer.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) || ctx.Err() != nil {
		return nil
	}
	return err
}

func (s *webSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestStarted := time.Now()
	cfg := s.opts.Config()
	wsCfg := cfg.WebSocket
	if !wsCfg.Enabled {
		wsmcMetrics.request("disabled", "unknown", time.Since(requestStarted))
		http.Error(w, "WebSocket transport disabled", http.StatusServiceUnavailable)
		return
	}
	if path := wsCfg.EffectivePath(); path != "" && r.URL.Path != path {
		wsmcMetrics.request("invalid_path", "unknown", time.Since(requestStarted))
		http.NotFound(w, r)
		return
	}

	virtualHost := webSocketHost(r.Host)
	matchedHost, route, groups := FindRouteWithGroups(virtualHost, cfg.Routes...)
	if route == nil || !route.WebSocket.Enabled {
		wsmcMetrics.request("no_route", "unknown", time.Since(requestStarted))
		http.Error(w, "No WebSocket route", http.StatusNotFound)
		return
	}

	remoteAddr, clientIP := s.clientAddr(r, wsCfg)
	if !s.acquireIP(clientIP, wsCfg.EffectiveMaxConnectionsPerIP()) {
		wsmcMetrics.request("quota_rejected", string(route.WebSocket.EffectiveMode()), time.Since(requestStarted))
		http.Error(w, "Too many WebSocket connections", http.StatusTooManyRequests)
		return
	}
	defer s.releaseIP(clientIP)

	compression := websocket.CompressionDisabled
	if wsCfg.Compression {
		compression = websocket.CompressionContextTakeover
	}
	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: compression})
	if err != nil {
		wsmcMetrics.request("handshake_failed", string(route.WebSocket.EffectiveMode()), time.Since(requestStarted))
		s.opts.Logger.V(1).Info("websocket handshake failed", "error", err, "remoteAddr", r.RemoteAddr)
		return
	}
	connCtx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	mode := string(route.WebSocket.EffectiveMode())
	client := newWebSocketStreamConn(connCtx, wsConn, wsCfg, remoteAddr, virtualHost, mode)
	wsmcMetrics.request("accepted", mode, time.Since(requestStarted))
	wsmcMetrics.addActive(mode, 1)
	defer wsmcMetrics.addActive(mode, -1)
	sessionStarted := time.Now()
	defer func() { wsmcMetrics.session(mode, time.Since(sessionStarted)) }()

	log := s.opts.Logger.WithName("lite").WithName("websocket").WithValues(
		"clientAddr", remoteAddr, "virtualHost", virtualHost, "route", matchedHost,
	)
	if route.WebSocket.EffectiveMode() == config.WebSocketModeRawPassthrough {
		defer client.Close()
		s.serveRaw(connCtx, log, r, client, route, matchedHost, groups, clientIP, wsCfg)
		return
	}

	// HandleConn owns the stream until the translated Minecraft connection closes.
	s.opts.HandleConn(client)
}

func (s *webSocketServer) serveRaw(
	ctx context.Context,
	log logr.Logger,
	r *http.Request,
	client net.Conn,
	route *config.Route,
	matchedHost string,
	groups []string,
	clientIP string,
	listenerCfg config.WebSocketListenerConfig,
) {
	backends := route.Backend.Copy()
	for len(backends) > 0 {
		backend, backendLog, ok := s.opts.StrategyManager.GetNextBackend(log, route, matchedHost, backends)
		if !ok {
			break
		}
		backend = SubstituteBackendParams(backend, groups)
		backends = removeWebSocketBackend(backends, backend, groups)
		dst, err := s.dialRawBackend(ctx, r, backend, route.WebSocket, clientIP, listenerCfg)
		if err != nil {
			backendLog.V(1).Info("failed to connect websocket backend", "backendAddr", backend, "error", err)
			continue
		}
		defer dst.Close()
		var decrement func()
		if route.Strategy == config.StrategyLeastConnections {
			decrement = s.opts.StrategyManager.IncrementConnection(backend)
			defer decrement()
		}
		log.Info("forwarding raw websocket stream", "backendAddr", backend)
		pipe(log, client, dst)
		return
	}
	log.Info("all websocket backends failed")
}

func (s *webSocketServer) dialRawBackend(
	ctx context.Context,
	inbound *http.Request,
	backend string,
	routeCfg config.WebSocketRouteConfig,
	clientIP string,
	listenerCfg config.WebSocketListenerConfig,
) (net.Conn, error) {
	addr, err := netutil.Parse(backend, "tcp")
	if err != nil {
		return nil, err
	}
	host := addr.String()
	if _, port := netutil.HostPort(addr); port == 0 {
		host = net.JoinHostPort(addr.String(), "25565")
	}
	scheme := routeCfg.BackendScheme
	if scheme == "" {
		scheme = "ws"
	}
	path := routeCfg.BackendPath
	if path == "" {
		path = inbound.URL.Path
		if path == "" {
			path = "/"
		}
	}
	target := (&url.URL{Scheme: scheme, Host: host, Path: path, RawQuery: inbound.URL.RawQuery}).String()

	header := make(http.Header)
	header.Set("X-Forwarded-For", clientIP)
	if name := listenerCfg.EffectiveForwardedForHeader(); !strings.EqualFold(name, "X-Forwarded-For") {
		header.Set(name, clientIP)
	}
	header.Set("X-Forwarded-Host", inbound.Host)
	header.Set("X-Forwarded-Proto", forwardedProto(inbound, s.trusted))
	compression := websocket.CompressionDisabled
	if listenerCfg.Compression {
		compression = websocket.CompressionContextTakeover
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	dialer := &net.Dialer{Timeout: listenerCfg.EffectiveHandshakeTimeout()}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, address)
		if err == nil {
			if tuneErr := tuneWebSocketTCP(conn, listenerCfg); tuneErr != nil {
				s.opts.Logger.Info("failed to tune websocket backend TCP socket", "error", tuneErr, "backendAddr", address)
			}
			options := tcpbrutal.Options{}
			if s.opts.BackendTCPBrutal != nil {
				options = s.opts.BackendTCPBrutal()
			}
			if applyErr := tcpbrutal.Apply(conn, options); applyErr != nil {
				s.opts.Logger.Info("failed to apply TCP Brutal to websocket backend", "error", applyErr, "backendAddr", address)
			}
		}
		return conn, err
	}
	httpClient := &http.Client{Transport: transport}
	dialCtx, cancel := context.WithTimeout(ctx, listenerCfg.EffectiveHandshakeTimeout())
	defer cancel()
	wsConn, _, err := websocket.Dial(dialCtx, target, &websocket.DialOptions{
		HTTPClient:      httpClient,
		HTTPHeader:      header,
		Host:            firstNonEmpty(routeCfg.BackendHost, inbound.Host),
		CompressionMode: compression,
	})
	if err != nil {
		return nil, err
	}
	return newWebSocketStreamConn(ctx, wsConn, listenerCfg, nil, "", "raw-backend"), nil
}

type webSocketListener struct {
	net.Listener
	apply func() tcpbrutal.Options
	cfg   config.WebSocketListenerConfig
	log   logr.Logger
}

func (l *webSocketListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err == nil {
		if tuneErr := tuneWebSocketTCP(conn, l.cfg); tuneErr != nil {
			l.log.Info("failed to tune websocket client TCP socket", "error", tuneErr, "remoteAddr", conn.RemoteAddr())
		}
		if applyErr := tcpbrutal.Apply(conn, l.apply()); applyErr != nil {
			l.log.Info("failed to apply TCP Brutal to websocket client", "error", applyErr, "remoteAddr", conn.RemoteAddr())
		}
	}
	return conn, err
}

func parseTrustedProxies(values []string) ([]netip.Prefix, error) {
	result := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", value, err)
		}
		result = append(result, prefix)
	}
	return result, nil
}

func (s *webSocketServer) clientAddr(r *http.Request, cfg config.WebSocketListenerConfig) (net.Addr, string) {
	host, port, _ := net.SplitHostPort(r.RemoteAddr)
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return &net.TCPAddr{}, host
	}
	client := peer.Unmap()
	forwarded := false
	if prefixContains(s.trusted, client) {
		parts := strings.Split(r.Header.Get(cfg.EffectiveForwardedForHeader()), ",")
		for i := len(parts) - 1; i >= 0 && prefixContains(s.trusted, client); i-- {
			candidate, parseErr := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if parseErr != nil {
				break
			}
			client = candidate.Unmap()
			forwarded = true
		}
	}
	portNumber := 0
	if !forwarded {
		_, _ = fmt.Sscanf(port, "%d", &portNumber)
	}
	return &net.TCPAddr{IP: net.IP(client.AsSlice()), Port: portNumber}, client.String()
}

func prefixContains(prefixes []netip.Prefix, addr netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func forwardedProto(r *http.Request, trusted []netip.Prefix) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	peer, _ := netip.ParseAddr(host)
	if peer.IsValid() && prefixContains(trusted, peer.Unmap()) {
		if value := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]); value == "http" || value == "https" || value == "ws" || value == "wss" {
			return value
		}
	}
	if r.TLS != nil {
		return "wss"
	}
	return "ws"
}

func webSocketHost(hostport string) string {
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return strings.Trim(host, "[]")
	}
	return strings.Trim(hostport, "[]")
}

func (s *webSocketServer) acquireIP(ip string, limit int) bool {
	s.quotaMu.Lock()
	defer s.quotaMu.Unlock()
	if s.activeIPs[ip] >= limit {
		return false
	}
	s.activeIPs[ip]++
	return true
}
func (s *webSocketServer) releaseIP(ip string) {
	s.quotaMu.Lock()
	defer s.quotaMu.Unlock()
	if s.activeIPs[ip] <= 1 {
		delete(s.activeIPs, ip)
	} else {
		s.activeIPs[ip]--
	}
}

func removeWebSocketBackend(backends []string, selected string, groups []string) []string {
	for i, backend := range backends {
		if SubstituteBackendParams(backend, groups) == selected {
			return append(backends[:i], backends[i+1:]...)
		}
	}
	return backends[1:]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
