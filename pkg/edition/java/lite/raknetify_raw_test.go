package lite

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/go-logr/logr"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
	"go.minekube.com/gate/pkg/util/configutil"
)

func rawRaknetifyHintPacket(host string) []byte {
	hostBytes := []byte(host)
	packet := make([]byte, 1+len(rawRaknetifyRouteHintMagic)+1+2+len(hostBytes))
	offset := 0
	packet[offset] = rawRaknetifyRouteHintPacketID
	offset++
	copy(packet[offset:], rawRaknetifyRouteHintMagic)
	offset += len(rawRaknetifyRouteHintMagic)
	packet[offset] = rawRaknetifyRouteHintVersion
	offset++
	binary.BigEndian.PutUint16(packet[offset:offset+2], uint16(len(hostBytes)))
	offset += 2
	copy(packet[offset:], hostBytes)
	return packet
}

func rawRaknetifyHintPacketV2(host string, token []byte) []byte {
	hostBytes := []byte(host)
	packet := make([]byte, 1+len(rawRaknetifyRouteHintMagic)+1+rawRaknetifyRouteTokenLen+2+len(hostBytes))
	offset := 0
	packet[offset] = rawRaknetifyRouteHintPacketID
	offset++
	copy(packet[offset:], rawRaknetifyRouteHintMagic)
	offset += len(rawRaknetifyRouteHintMagic)
	packet[offset] = rawRaknetifyRouteHintVersion2
	offset++
	copy(packet[offset:], token)
	offset += rawRaknetifyRouteTokenLen
	binary.BigEndian.PutUint16(packet[offset:offset+2], uint16(len(hostBytes)))
	offset += 2
	copy(packet[offset:], hostBytes)
	return packet
}

func TestDecodeRawRaknetifyRouteHint(t *testing.T) {
	hint, ok, err := decodeRawRaknetifyRouteHint(rawRaknetifyHintPacket("Example.COM:25566"))
	if err != nil {
		t.Fatalf("decodeRawRaknetifyRouteHint returned error: %v", err)
	}
	if !ok {
		t.Fatal("decodeRawRaknetifyRouteHint did not recognize the hint packet")
	}
	if hint.host != "Example.COM:25566" {
		t.Fatalf("unexpected host: %q", hint.host)
	}
	if hint.hasToken {
		t.Fatal("legacy route hint should not include a token")
	}
}

func TestDecodeRawRaknetifyRouteHintV2(t *testing.T) {
	token := []byte("0123456789abcdef")
	hint, ok, err := decodeRawRaknetifyRouteHint(rawRaknetifyHintPacketV2("Example.COM:25566", token))
	if err != nil {
		t.Fatalf("decodeRawRaknetifyRouteHint returned error: %v", err)
	}
	if !ok {
		t.Fatal("decodeRawRaknetifyRouteHint did not recognize the hint packet")
	}
	if hint.host != "Example.COM:25566" {
		t.Fatalf("unexpected host: %q", hint.host)
	}
	if !hint.hasToken || hint.token != string(token) {
		t.Fatalf("unexpected token: has=%v token=%q", hint.hasToken, hint.token)
	}
}

func TestDecodeRawRaknetifyRouteHintIgnoresNonHintPackets(t *testing.T) {
	hint, ok, err := decodeRawRaknetifyRouteHint([]byte{0x05, 0x00, 0x00})
	if err != nil {
		t.Fatalf("decodeRawRaknetifyRouteHint returned error: %v", err)
	}
	if ok || hint.host != "" {
		t.Fatalf("expected non-hint packet to be ignored, got host=%q ok=%v", hint.host, ok)
	}
}

func TestCleanRawRaknetifyHost(t *testing.T) {
	host := cleanRawRaknetifyHost(" Example.COM:25566\x00FML2 ")
	if host != "Example.COM" {
		t.Fatalf("unexpected cleaned host: %q", host)
	}
}

func TestNormalizeRawRaknetifyBackendAddsDefaultPort(t *testing.T) {
	backend, err := normalizeRawRaknetifyBackend("127.0.0.1")
	if err != nil {
		t.Fatalf("normalizeRawRaknetifyBackend returned error: %v", err)
	}
	if backend != "127.0.0.1:25565" {
		t.Fatalf("unexpected backend: %q", backend)
	}
}

func TestRawRaknetifyLegacyHintReplacesExistingSession(t *testing.T) {
	srv, clientAddr, cleanup := newRawRaknetifyTestServer(t)
	defer cleanup()

	first, err := srv.ensureSession(clientAddr, rawRaknetifyRouteHint{host: "example.com"})
	if err != nil {
		t.Fatalf("first ensureSession returned error: %v", err)
	}
	second, err := srv.ensureSession(clientAddr, rawRaknetifyRouteHint{host: "example.com"})
	if err != nil {
		t.Fatalf("second ensureSession returned error: %v", err)
	}
	if first == second {
		t.Fatal("legacy route hint reused an existing session")
	}
	if srv.sessionCount.Load() != 1 {
		t.Fatalf("unexpected session count: %d", srv.sessionCount.Load())
	}
	if loaded, ok := srv.loadSession(clientAddr); !ok || loaded != second {
		t.Fatalf("expected the replacement session to remain loaded")
	}
}

func TestRawRaknetifyV2HintTokenDeduplicatesRetries(t *testing.T) {
	srv, clientAddr, cleanup := newRawRaknetifyTestServer(t)
	defer cleanup()

	token := string([]byte("0123456789abcdef"))
	first, err := srv.ensureSession(clientAddr, rawRaknetifyRouteHint{
		host:     "example.com",
		token:    token,
		hasToken: true,
	})
	if err != nil {
		t.Fatalf("first ensureSession returned error: %v", err)
	}
	second, err := srv.ensureSession(clientAddr, rawRaknetifyRouteHint{
		host:     "example.com",
		token:    token,
		hasToken: true,
	})
	if err != nil {
		t.Fatalf("second ensureSession returned error: %v", err)
	}
	if first != second {
		t.Fatal("duplicate v2 route hint did not reuse the existing session")
	}
	if srv.sessionCount.Load() != 1 {
		t.Fatalf("unexpected session count: %d", srv.sessionCount.Load())
	}

	replacement, err := srv.ensureSession(clientAddr, rawRaknetifyRouteHint{
		host:     "example.com",
		token:    string([]byte("fedcba9876543210")),
		hasToken: true,
	})
	if err != nil {
		t.Fatalf("replacement ensureSession returned error: %v", err)
	}
	if replacement == first {
		t.Fatal("new v2 route token reused the previous session")
	}
	srv.closeSession(clientAddr.String(), first, "test")
	if loaded, ok := srv.loadSession(clientAddr); !ok || loaded != replacement {
		t.Fatal("closing the old session removed the replacement session")
	}
}

func newRawRaknetifyTestServer(t *testing.T) (*rawRaknetifyServer, net.Addr, func()) {
	t.Helper()

	gateConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen for gate test socket: %v", err)
	}
	backendConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		_ = gateConn.Close()
		t.Fatalf("failed to listen for backend test socket: %v", err)
	}
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:45678")
	if err != nil {
		_ = gateConn.Close()
		_ = backendConn.Close()
		t.Fatalf("failed to resolve client test address: %v", err)
	}

	srv := &rawRaknetifyServer{
		conn: gateConn,
		routes: func() []config.Route {
			return []config.Route{
				{
					Host:    configutil.SingleOrMulti[string]{"example.com"},
					Backend: configutil.SingleOrMulti[string]{backendConn.LocalAddr().String()},
					Raknetify: config.RaknetifyConfig{
						Enabled: true,
						Mode:    config.RaknetifyModeRawPassthrough,
					},
				},
			}
		},
		strategyManager: NewStrategyManager(),
		log:             logr.Discard(),
	}
	cleanup := func() {
		srv.closeAllSessions()
		_ = gateConn.Close()
		_ = backendConn.Close()
	}
	return srv, clientAddr, cleanup
}
