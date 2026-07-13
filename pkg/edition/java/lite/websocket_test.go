package lite

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
)

func TestWebSocketTranslateUsesHostXFFAndStreamFraming(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	metadata := make(chan struct {
		host string
		addr string
	}, 1)
	cfg := config.Config{
		WebSocket: config.WebSocketListenerConfig{
			Enabled: true, Path: "/mc", TrustedProxies: []string{"127.0.0.0/8"},
			FramePayloadSize: 3, ReadLimit: 1024, MaxConnectionsPerIP: 4,
		},
		Routes: []config.Route{{
			Host: []string{"play.example.com"}, Backend: []string{"127.0.0.1:25565"},
			WebSocket: config.WebSocketRouteConfig{Enabled: true, Mode: config.WebSocketModeTranslate},
		}},
	}
	server := &webSocketServer{
		ctx: ctx,
		opts: WebSocketOptions{
			Config:          func() config.Config { return cfg },
			StrategyManager: NewStrategyManager(),
			Logger:          logr.Discard(),
			HandleConn: func(conn net.Conn) {
				defer conn.Close()
				metadata <- struct{ host, addr string }{conn.(interface{ WebSocketVirtualHost() string }).WebSocketVirtualHost(), conn.RemoteAddr().String()}
				payload := make([]byte, 8)
				_, _ = io.ReadFull(conn, payload)
				_, _ = conn.Write(payload)
			},
		},
		trusted: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}, activeIPs: make(map[string]int),
	}
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/mc"
	ws, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		Host:       "play.example.com",
		HTTPHeader: http.Header{"X-Forwarded-For": []string{"203.0.113.9"}},
	})
	require.NoError(t, err)
	conn := websocket.NetConn(ctx, ws, websocket.MessageBinary)
	defer conn.Close()
	_, err = conn.Write([]byte("12345678"))
	require.NoError(t, err)
	got := make([]byte, 8)
	_, err = io.ReadFull(conn, got)
	require.NoError(t, err)
	require.Equal(t, "12345678", string(got))
	info := <-metadata
	require.Equal(t, "play.example.com", info.host)
	require.Contains(t, info.addr, "203.0.113.9")
}

func TestWebSocketRawPassthrough(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		require.NoError(t, err)
		stream := websocket.NetConn(ctx, conn, websocket.MessageBinary)
		defer stream.Close()
		_, _ = io.Copy(stream, stream)
	}))
	defer backend.Close()
	backendAddr := strings.TrimPrefix(backend.URL, "http://")
	cfg := config.Config{
		WebSocket: config.WebSocketListenerConfig{Enabled: true, Path: "/mc", FramePayloadSize: 4, ReadLimit: 1024, MaxConnectionsPerIP: 4},
		Routes: []config.Route{{
			Host: []string{"raw.example.com"}, Backend: []string{backendAddr},
			WebSocket: config.WebSocketRouteConfig{Enabled: true, Mode: config.WebSocketModeRawPassthrough, BackendScheme: "ws", BackendPath: "/mc"},
		}},
	}
	server := &webSocketServer{
		ctx:       ctx,
		opts:      WebSocketOptions{Config: func() config.Config { return cfg }, StrategyManager: NewStrategyManager(), HandleConn: func(net.Conn) {}, Logger: logr.Discard()},
		activeIPs: make(map[string]int),
	}
	gate := httptest.NewServer(server)
	defer gate.Close()
	wsURL := "ws" + strings.TrimPrefix(gate.URL, "http") + "/mc"
	ws, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{Host: "raw.example.com"})
	require.NoError(t, err)
	stream := websocket.NetConn(ctx, ws, websocket.MessageBinary)
	defer stream.Close()
	want := []byte("raw-websocket-payload")
	_, err = stream.Write(want)
	require.NoError(t, err)
	got := make([]byte, len(want))
	_, err = io.ReadFull(stream, got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

// WSMC's current client (WebSocketClientHandler/WebSocketHandler) uses RFC6455
// V13, sends every Minecraft ByteBuf as a binary message and exposes received
// binary message payloads as one continuous Netty byte stream. Verify both
// boundaries against that contract rather than only using websocket.NetConn.
func TestWebSocketTranslateMatchesWSMCBinaryFrameContract(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg := config.Config{
		WebSocket: config.WebSocketListenerConfig{
			Enabled: true, Path: "/mc", FramePayloadSize: 65536,
			ReadLimit: 1 << 20, MaxConnectionsPerIP: 4,
		},
		Routes: []config.Route{{
			Host: []string{"wsmc.example.com"}, Backend: []string{"127.0.0.1:25565"},
			WebSocket: config.WebSocketRouteConfig{Enabled: true, Mode: config.WebSocketModeTranslate},
		}},
	}
	server := &webSocketServer{
		ctx: ctx,
		opts: WebSocketOptions{
			Config: func() config.Config { return cfg }, StrategyManager: NewStrategyManager(), Logger: logr.Discard(),
			HandleConn: func(conn net.Conn) {
				defer conn.Close()
				in := make([]byte, 5)
				_, _ = io.ReadFull(conn, in)
				_, _ = conn.Write(make([]byte, 65536*2+7))
			},
		},
		activeIPs: make(map[string]int),
	}
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	ws, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(httpServer.URL, "http")+"/mc", &websocket.DialOptions{Host: "wsmc.example.com"})
	require.NoError(t, err)
	defer ws.CloseNow()
	ws.SetReadLimit(1 << 20)
	require.NoError(t, ws.Write(ctx, websocket.MessageBinary, []byte{1, 2}))
	require.NoError(t, ws.Write(ctx, websocket.MessageBinary, []byte{3, 4, 5}))
	total := 0
	frames := 0
	for total < 65536*2+7 {
		typ, payload, readErr := ws.Read(ctx)
		require.NoError(t, readErr)
		require.Equal(t, websocket.MessageBinary, typ)
		require.LessOrEqual(t, len(payload), 65536)
		total += len(payload)
		frames++
	}
	require.Equal(t, 65536*2+7, total)
	require.Equal(t, 3, frames)
}

func TestUntrustedProxyCannotSpoofForwardedFor(t *testing.T) {
	s := &webSocketServer{}
	r := httptest.NewRequest(http.MethodGet, "http://example.com/mc", nil)
	r.RemoteAddr = "198.51.100.2:1234"
	r.Header.Set("X-Forwarded-For", "203.0.113.9")
	addr, ip := s.clientAddr(r, config.WebSocketListenerConfig{})
	require.Equal(t, "198.51.100.2", ip)
	require.Contains(t, addr.String(), "198.51.100.2")
}
