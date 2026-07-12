package config

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"regexp"
	"slices"
	"strconv"
	"time"

	"go.minekube.com/gate/pkg/edition/java/forge/modinfo"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/gate/proto"
	"go.minekube.com/gate/pkg/util/configutil"
	"go.minekube.com/gate/pkg/util/favicon"
	"go.minekube.com/gate/pkg/util/netutil"
)

// DefaultConfig is the default configuration for Lite mode.
var DefaultConfig = Config{
	Enabled: false,
	Routes:  []Route{},
}

type (
	// Config is the configuration for Lite mode.
	Config struct {
		Enabled   bool                    `yaml:"enabled,omitempty" json:"enabled,omitempty"`
		WebSocket WebSocketListenerConfig `yaml:"websocket,omitempty" json:"websocket,omitempty"`
		Routes    []Route                 `yaml:"routes,omitempty" json:"routes,omitempty"`
	}
	Route struct {
		Host          configutil.SingleOrMulti[string] `json:"host,omitempty" yaml:"host,omitempty"`
		Backend       configutil.SingleOrMulti[string] `json:"backend,omitempty" yaml:"backend,omitempty"`
		CachePingTTL  configutil.Duration              `json:"cachePingTTL,omitempty" yaml:"cachePingTTL,omitempty"` // 0 = default, < 0 = disabled
		Fallback      *Status                          `json:"fallback,omitempty" yaml:"fallback,omitempty"`         // nil = disabled
		ProxyProtocol bool                             `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
		// Deprecated: use TCPShieldRealIP instead.
		RealIP            bool                 `json:"realIP,omitempty" yaml:"realIP,omitempty"`
		TCPShieldRealIP   bool                 `json:"tcpShieldRealIP,omitempty" yaml:"tcpShieldRealIP,omitempty"`
		ModifyVirtualHost bool                 `json:"modifyVirtualHost,omitempty" yaml:"modifyVirtualHost,omitempty"`
		Strategy          Strategy             `json:"strategy,omitempty" yaml:"strategy,omitempty"`
		Raknetify         RaknetifyConfig      `json:"raknetify,omitempty" yaml:"raknetify,omitempty"`
		WebSocket         WebSocketRouteConfig `json:"websocket,omitempty" yaml:"websocket,omitempty"`
	}
	WebSocketListenerConfig struct {
		Enabled              bool                `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Bind                 string              `json:"bind,omitempty" yaml:"bind,omitempty"`
		Path                 string              `json:"path,omitempty" yaml:"path,omitempty"`
		TrustedProxies       []string            `json:"trustedProxies,omitempty" yaml:"trustedProxies,omitempty"`
		ForwardedForHeader   string              `json:"forwardedForHeader,omitempty" yaml:"forwardedForHeader,omitempty"`
		ReadLimit            int64               `json:"readLimit,omitempty" yaml:"readLimit,omitempty"`
		FramePayloadSize     int                 `json:"framePayloadSize,omitempty" yaml:"framePayloadSize,omitempty"`
		MinFramePayloadSize  int                 `json:"minFramePayloadSize,omitempty" yaml:"minFramePayloadSize,omitempty"`
		AdaptiveFraming      *bool               `json:"adaptiveFraming,omitempty" yaml:"adaptiveFraming,omitempty"`
		CoalesceWindow       configutil.Duration `json:"coalesceWindow,omitempty" yaml:"coalesceWindow,omitempty"`
		CoalesceLimit        int                 `json:"coalesceLimit,omitempty" yaml:"coalesceLimit,omitempty"`
		MaxPendingBytes      int                 `json:"maxPendingBytes,omitempty" yaml:"maxPendingBytes,omitempty"`
		BackpressureTimeout  configutil.Duration `json:"backpressureTimeout,omitempty" yaml:"backpressureTimeout,omitempty"`
		IdleTimeout          configutil.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
		HandshakeTimeout     configutil.Duration `json:"handshakeTimeout,omitempty" yaml:"handshakeTimeout,omitempty"`
		MaxConnectionsPerIP  int                 `json:"maxConnectionsPerIP,omitempty" yaml:"maxConnectionsPerIP,omitempty"`
		Compression          bool                `json:"compression,omitempty" yaml:"compression,omitempty"`
		TCPNoDelay           *bool               `json:"tcpNoDelay,omitempty" yaml:"tcpNoDelay,omitempty"`
		SocketReadBuffer     int                 `json:"socketReadBuffer,omitempty" yaml:"socketReadBuffer,omitempty"`
		SocketWriteBuffer    int                 `json:"socketWriteBuffer,omitempty" yaml:"socketWriteBuffer,omitempty"`
		TCPKeepAlive         configutil.Duration `json:"tcpKeepAlive,omitempty" yaml:"tcpKeepAlive,omitempty"`
		TCPKeepAliveInterval configutil.Duration `json:"tcpKeepAliveInterval,omitempty" yaml:"tcpKeepAliveInterval,omitempty"`
		TCPKeepAliveCount    int                 `json:"tcpKeepAliveCount,omitempty" yaml:"tcpKeepAliveCount,omitempty"`
		TCPNotSentLowAt      int                 `json:"tcpNotSentLowAt,omitempty" yaml:"tcpNotSentLowAt,omitempty"`
	}
	WebSocketRouteConfig struct {
		Enabled       bool          `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Mode          WebSocketMode `json:"mode,omitempty" yaml:"mode,omitempty"`
		BackendScheme string        `json:"backendScheme,omitempty" yaml:"backendScheme,omitempty"`
		BackendPath   string        `json:"backendPath,omitempty" yaml:"backendPath,omitempty"`
		BackendHost   string        `json:"backendHost,omitempty" yaml:"backendHost,omitempty"`
	}
	RaknetifyConfig struct {
		Enabled        bool                          `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Mode           RaknetifyMode                 `json:"mode,omitempty" yaml:"mode,omitempty"`
		RawPassthrough RaknetifyRawPassthroughConfig `json:"rawPassthrough,omitempty" yaml:"rawPassthrough,omitempty"`
	}
	RaknetifyRawPassthroughConfig struct {
		QOS              RaknetifyQOSConfig  `json:"qos,omitempty" yaml:"qos,omitempty"`
		IdleTimeout      configutil.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
		WriteTimeout     configutil.Duration `json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
		PacingInterval   configutil.Duration `json:"pacingInterval,omitempty" yaml:"pacingInterval,omitempty"`
		QueueSize        int                 `json:"queueSize,omitempty" yaml:"queueSize,omitempty"`
		MaxSessionsPerIP int                 `json:"maxSessionsPerIP,omitempty" yaml:"maxSessionsPerIP,omitempty"`
	}
	RaknetifyQOSConfig struct {
		Mode RaknetifyQOSMode `json:"mode,omitempty" yaml:"mode,omitempty"`
		TOS  *int             `json:"tos,omitempty" yaml:"tos,omitempty"`
	}
	Status struct {
		MOTD    *configutil.Component `yaml:"motd,omitempty" json:"motd,omitempty"`
		Version ping.Version          `yaml:"version,omitempty" json:"version,omitempty"`
		Players *ping.Players         `json:"players,omitempty" yaml:"players,omitempty"`
		Favicon favicon.Favicon       `yaml:"favicon,omitempty" json:"favicon,omitempty"`
		ModInfo modinfo.ModInfo       `yaml:"modInfo,omitempty" json:"modInfo,omitempty"`
	}
)

type RaknetifyMode string
type RaknetifyQOSMode string
type WebSocketMode string

const (
	RaknetifyModeRawPassthrough RaknetifyMode = "raw-passthrough"

	RaknetifyQOSModeDefault RaknetifyQOSMode = "default"
	RaknetifyQOSModeClear   RaknetifyQOSMode = "clear"
	RaknetifyQOSModeCustom  RaknetifyQOSMode = "custom"

	DefaultRawRaknetifyQueueSize     = 256
	MaxRawRaknetifyQueueSize         = 4096
	DefaultRawRaknetifySessionsPerIP = 64
	MaxRawRaknetifySessionsPerIP     = 1024

	WebSocketModeTranslate      WebSocketMode = "translate"
	WebSocketModeRawPassthrough WebSocketMode = "raw-passthrough"
)

func (c WebSocketListenerConfig) EffectivePath() string {
	return c.Path
}
func (c WebSocketListenerConfig) EffectiveForwardedForHeader() string {
	if c.ForwardedForHeader == "" {
		return "X-Forwarded-For"
	}
	return c.ForwardedForHeader
}
func (c WebSocketListenerConfig) EffectiveReadLimit() int64 {
	if c.ReadLimit == 0 {
		return 16 << 20
	}
	return c.ReadLimit
}
func (c WebSocketListenerConfig) EffectiveFramePayloadSize() int {
	if c.FramePayloadSize == 0 {
		return 64 << 10
	}
	return c.FramePayloadSize
}
func (c WebSocketListenerConfig) EffectiveMinFramePayloadSize() int {
	if c.MinFramePayloadSize == 0 {
		return 4 << 10
	}
	return c.MinFramePayloadSize
}
func (c WebSocketListenerConfig) EffectiveAdaptiveFraming() bool {
	return c.AdaptiveFraming == nil || *c.AdaptiveFraming
}
func (c WebSocketListenerConfig) EffectiveCoalesceWindow() time.Duration {
	if c.CoalesceWindow == 0 {
		return 200 * time.Microsecond
	}
	return time.Duration(c.CoalesceWindow)
}
func (c WebSocketListenerConfig) EffectiveCoalesceLimit() int {
	if c.CoalesceLimit == 0 {
		return c.EffectiveFramePayloadSize()
	}
	return c.CoalesceLimit
}
func (c WebSocketListenerConfig) EffectiveMaxPendingBytes() int {
	if c.MaxPendingBytes == 0 {
		return 4 << 20
	}
	return c.MaxPendingBytes
}
func (c WebSocketListenerConfig) EffectiveBackpressureTimeout() time.Duration {
	if c.BackpressureTimeout == 0 {
		return 10 * time.Second
	}
	return time.Duration(c.BackpressureTimeout)
}
func (c WebSocketListenerConfig) EffectiveIdleTimeout() time.Duration {
	if c.IdleTimeout == 0 {
		return 90 * time.Second
	}
	return time.Duration(c.IdleTimeout)
}
func (c WebSocketListenerConfig) EffectiveTCPNoDelay() bool {
	return c.TCPNoDelay == nil || *c.TCPNoDelay
}
func (c WebSocketListenerConfig) EffectiveTCPKeepAlive() time.Duration {
	if c.TCPKeepAlive == 0 {
		return 30 * time.Second
	}
	return time.Duration(c.TCPKeepAlive)
}
func (c WebSocketListenerConfig) EffectiveTCPKeepAliveInterval() time.Duration {
	if c.TCPKeepAliveInterval == 0 {
		return 10 * time.Second
	}
	return time.Duration(c.TCPKeepAliveInterval)
}
func (c WebSocketListenerConfig) EffectiveTCPKeepAliveCount() int {
	if c.TCPKeepAliveCount == 0 {
		return 3
	}
	return c.TCPKeepAliveCount
}
func (c WebSocketListenerConfig) EffectiveHandshakeTimeout() time.Duration {
	if c.HandshakeTimeout == 0 {
		return 10 * time.Second
	}
	return time.Duration(c.HandshakeTimeout)
}
func (c WebSocketListenerConfig) EffectiveMaxConnectionsPerIP() int {
	if c.MaxConnectionsPerIP == 0 {
		return 64
	}
	return c.MaxConnectionsPerIP
}
func (c WebSocketRouteConfig) EffectiveMode() WebSocketMode {
	if c.Mode == "" {
		return WebSocketModeTranslate
	}
	return c.Mode
}

// Response returns the configured status response.
func (s *Status) Response(proto.Protocol) (*ping.ServerPing, error) {
	return &ping.ServerPing{
		Version:     s.Version,
		Players:     s.Players,
		Description: s.MOTD.C(),
		Favicon:     s.Favicon,
		ModInfo:     &s.ModInfo,
	}, nil
}

// GetCachePingTTL returns the configured ping cache TTL or a default duration if not set.
func (r *Route) GetCachePingTTL() time.Duration {
	const defaultTTL = time.Second * 10
	if r.CachePingTTL == 0 {
		return defaultTTL
	}
	return time.Duration(r.CachePingTTL)
}

// CachePingEnabled returns true if the route has a ping cache enabled.
func (r *Route) CachePingEnabled() bool { return r.GetCachePingTTL() > 0 }

// GetTCPShieldRealIP returns the configured TCPShieldRealIP or deprecated RealIP value.
func (r *Route) GetTCPShieldRealIP() bool { return r.TCPShieldRealIP || r.RealIP }

// RaknetifyMode returns the configured Raknetify mode or the only supported raw mode.
func (r *Route) RaknetifyMode() RaknetifyMode {
	if r.Raknetify.Mode == "" {
		return RaknetifyModeRawPassthrough
	}
	return r.Raknetify.Mode
}

// Strategy represents a load balancing strategy for lite mode routes.
type Strategy string

const (
	// StrategySequential selects backends in config order for each connection attempt.
	StrategySequential Strategy = "sequential"

	// StrategyRandom selects a random backend from available options.
	StrategyRandom Strategy = "random"

	// StrategyRoundRobin cycles through backends in order for each new connection.
	StrategyRoundRobin Strategy = "round-robin"

	// StrategyLeastConnections selects the backend with the fewest active connections.
	StrategyLeastConnections Strategy = "least-connections"

	// StrategyLowestLatency selects the backend with the lowest ping response time.
	StrategyLowestLatency Strategy = "lowest-latency"
)

var allowedStrategies = []Strategy{
	StrategySequential,
	StrategyRandom,
	StrategyRoundRobin,
	StrategyLeastConnections,
	StrategyLowestLatency,
}

var allowedRaknetifyModes = []RaknetifyMode{
	RaknetifyModeRawPassthrough,
}

var allowedRaknetifyQOSModes = []RaknetifyQOSMode{
	RaknetifyQOSModeDefault,
	RaknetifyQOSModeClear,
	RaknetifyQOSModeCustom,
}

func (c Config) Validate() (warns []error, errs []error) {
	e := func(m string, args ...any) { errs = append(errs, fmt.Errorf(m, args...)) }
	w := func(m string, args ...any) { warns = append(warns, fmt.Errorf(m, args...)) }
	if c.WebSocket.Enabled {
		if c.WebSocket.Bind == "" {
			e("WebSocket: bind is required when enabled")
		}
		if path := c.WebSocket.EffectivePath(); path != "" && path[0] != '/' {
			e("WebSocket: path must start with '/'")
		}
		if c.WebSocket.EffectiveReadLimit() < 1024 {
			e("WebSocket: readLimit must be at least 1024")
		}
		if size := c.WebSocket.EffectiveFramePayloadSize(); size < 1 || int64(size) > c.WebSocket.EffectiveReadLimit() {
			e("WebSocket: framePayloadSize must be positive and not exceed readLimit")
		}
		if minSize := c.WebSocket.EffectiveMinFramePayloadSize(); minSize < 1 || minSize > c.WebSocket.EffectiveFramePayloadSize() {
			e("WebSocket: minFramePayloadSize must be positive and not exceed framePayloadSize")
		}
		if c.WebSocket.EffectiveCoalesceWindow() < 0 {
			e("WebSocket: coalesceWindow must be >= 0")
		}
		if limit := c.WebSocket.EffectiveCoalesceLimit(); limit < 1 || limit > c.WebSocket.EffectiveMaxPendingBytes() {
			e("WebSocket: coalesceLimit must be positive and not exceed maxPendingBytes")
		}
		if c.WebSocket.EffectiveMaxPendingBytes() < c.WebSocket.EffectiveFramePayloadSize() {
			e("WebSocket: maxPendingBytes must be at least framePayloadSize")
		}
		if c.WebSocket.EffectiveBackpressureTimeout() <= 0 {
			e("WebSocket: backpressureTimeout must be positive")
		}
		if c.WebSocket.EffectiveIdleTimeout() <= 0 {
			e("WebSocket: idleTimeout must be positive")
		}
		if c.WebSocket.EffectiveHandshakeTimeout() <= 0 {
			e("WebSocket: handshakeTimeout must be positive")
		}
		if c.WebSocket.EffectiveMaxConnectionsPerIP() < 1 {
			e("WebSocket: maxConnectionsPerIP must be positive")
		}
		if c.WebSocket.SocketReadBuffer < 0 || c.WebSocket.SocketWriteBuffer < 0 || c.WebSocket.TCPNotSentLowAt < 0 {
			e("WebSocket: socket buffer and tcpNotSentLowAt values must be >= 0")
		}
		if c.WebSocket.EffectiveTCPKeepAlive() <= 0 || c.WebSocket.EffectiveTCPKeepAliveInterval() <= 0 || c.WebSocket.EffectiveTCPKeepAliveCount() < 1 {
			e("WebSocket: TCP keepalive values must be positive")
		}
		for _, prefix := range c.WebSocket.TrustedProxies {
			if _, err := netip.ParsePrefix(prefix); err != nil {
				e("WebSocket: invalid trusted proxy CIDR %q: %v", prefix, err)
			}
		}
	}

	if len(c.Routes) == 0 {
		e("No routes configured")
		return
	}

	for i, ep := range c.Routes {
		if len(ep.Host) == 0 {
			e("Route %d: no host configured", i)
		}
		if len(ep.Backend) == 0 {
			e("Route %d: no backend configured", i)
		}
		if !slices.Contains(allowedStrategies, ep.Strategy) && ep.Strategy != "" {
			e("Route %d: invalid strategy '%s', allowed: %v", i, ep.Strategy, allowedStrategies)
		}
		if ep.Raknetify.Enabled {
			if ep.Raknetify.Mode == "" {
				ep.Raknetify.Mode = RaknetifyModeRawPassthrough
			}
			if !slices.Contains(allowedRaknetifyModes, ep.Raknetify.Mode) {
				e("Route %d: invalid raknetify mode '%s', allowed: %v", i, ep.Raknetify.Mode, allowedRaknetifyModes)
			}
			qos := ep.Raknetify.RawPassthrough.QOS
			if qos.Mode != "" && !slices.Contains(allowedRaknetifyQOSModes, qos.Mode) {
				e("Route %d: invalid raknetify raw passthrough qos mode '%s', allowed: %v", i, qos.Mode, allowedRaknetifyQOSModes)
			}
			if qos.Mode == RaknetifyQOSModeCustom && qos.TOS == nil {
				e("Route %d: raknetify raw passthrough qos mode 'custom' requires tos", i)
			}
			if qos.TOS != nil && (*qos.TOS < 0 || *qos.TOS > 255) {
				e("Route %d: raknetify raw passthrough qos tos must be between 0 and 255", i)
			}
			raw := ep.Raknetify.RawPassthrough
			if raw.IdleTimeout < 0 {
				e("Route %d: raknetify raw passthrough idleTimeout must be >= 0", i)
			}
			if raw.WriteTimeout < 0 {
				e("Route %d: raknetify raw passthrough writeTimeout must be >= 0", i)
			}
			if raw.PacingInterval < 0 {
				e("Route %d: raknetify raw passthrough pacingInterval must be >= 0", i)
			}
			if raw.QueueSize < 0 {
				e("Route %d: raknetify raw passthrough queueSize must be >= 0", i)
			}
			if raw.QueueSize > MaxRawRaknetifyQueueSize {
				e("Route %d: raknetify raw passthrough queueSize must be <= %d", i, MaxRawRaknetifyQueueSize)
			}
			if raw.MaxSessionsPerIP < 0 || raw.MaxSessionsPerIP > MaxRawRaknetifySessionsPerIP {
				e("Route %d: raknetify raw passthrough maxSessionsPerIP must be between 0 and %d", i, MaxRawRaknetifySessionsPerIP)
			}
		}
		if ep.WebSocket.Enabled {
			mode := ep.WebSocket.EffectiveMode()
			if mode != WebSocketModeTranslate && mode != WebSocketModeRawPassthrough {
				e("Route %d: invalid websocket mode %q", i, mode)
			}
			if mode == WebSocketModeRawPassthrough {
				scheme := ep.WebSocket.BackendScheme
				if scheme == "" {
					scheme = "ws"
				}
				if scheme != "ws" && scheme != "wss" {
					e("Route %d: websocket backendScheme must be ws or wss", i)
				}
				if ep.WebSocket.BackendPath != "" && ep.WebSocket.BackendPath[0] != '/' {
					e("Route %d: websocket backendPath must start with '/'", i)
				}
			}
		}

		// Validate parameter usage in backend addresses
		for hostIdx, host := range ep.Host {
			wildcardCount := countWildcards(host)
			for backendIdx, addr := range ep.Backend {
				paramIndices := extractParameterIndices(addr)
				if len(paramIndices) > 0 {
					// Check if parameters exceed available wildcards
					maxParam := 0
					for _, idx := range paramIndices {
						if idx > maxParam {
							maxParam = idx
						}
					}
					if maxParam > wildcardCount {
						w("Route %d: host %d '%s' has %d wildcard(s) but backend %d '%s' uses parameter $%d (parameters will not be substituted)",
							i, hostIdx, host, wildcardCount, backendIdx, addr, maxParam)
					}
					// Warn if no wildcards but parameters are used
					if wildcardCount == 0 {
						w("Route %d: host %d '%s' has no wildcards but backend %d '%s' uses parameters (parameters will not be substituted)",
							i, hostIdx, host, backendIdx, addr)
					}
				}

				// Validate address parsing (after parameter substitution would happen)
				// We can't fully validate addresses with parameters, but we can check the structure
				_, err := netutil.Parse(addr, "tcp")
				if err != nil {
					// If it contains parameters, it might be valid after substitution
					if !containsParameters(addr) {
						e("Route %d: backend %d: failed to parse address: %w", i, backendIdx, err)
					}
				}
			}
		}
	}

	return
}

// countWildcards counts the number of wildcard characters (* and ?) in a pattern.
func countWildcards(pattern string) int {
	count := 0
	escapeNext := false
	for _, r := range pattern {
		if escapeNext {
			escapeNext = false
			continue
		}
		if r == '\\' {
			escapeNext = true
			continue
		}
		if r == '*' || r == '?' {
			count++
		}
	}
	return count
}

// extractParameterIndices extracts all parameter indices ($1, $2, etc.) from a string.
// Returns a slice of unique parameter indices found.
func extractParameterIndices(s string) []int {
	// Match $ followed by one or more digits
	re := regexp.MustCompile(`\$(\d+)`)
	matches := re.FindAllStringSubmatch(s, -1)

	indices := make(map[int]bool)
	for _, match := range matches {
		if len(match) > 1 {
			if idx, err := strconv.Atoi(match[1]); err == nil {
				indices[idx] = true
			}
		}
	}

	result := make([]int, 0, len(indices))
	for idx := range indices {
		result = append(result, idx)
	}
	return result
}

// containsParameters returns true if the string contains parameter placeholders like $1, $2, etc.
func containsParameters(s string) bool {
	matched, _ := regexp.MatchString(`\$\d+`, s)
	return matched
}

// Equal returns true if the Routes are equal.
func (r *Route) Equal(other *Route) bool {
	j, err := json.Marshal(r)
	if err != nil {
		return false
	}
	o, err := json.Marshal(other)
	if err != nil {
		return false
	}
	return string(j) == string(o)
}
