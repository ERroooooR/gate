package lite

import (
	"encoding/binary"
	"testing"
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

func TestDecodeRawRaknetifyRouteHint(t *testing.T) {
	host, ok, err := decodeRawRaknetifyRouteHint(rawRaknetifyHintPacket("Example.COM:25566"))
	if err != nil {
		t.Fatalf("decodeRawRaknetifyRouteHint returned error: %v", err)
	}
	if !ok {
		t.Fatal("decodeRawRaknetifyRouteHint did not recognize the hint packet")
	}
	if host != "Example.COM:25566" {
		t.Fatalf("unexpected host: %q", host)
	}
}

func TestDecodeRawRaknetifyRouteHintIgnoresNonHintPackets(t *testing.T) {
	host, ok, err := decodeRawRaknetifyRouteHint([]byte{0x05, 0x00, 0x00})
	if err != nil {
		t.Fatalf("decodeRawRaknetifyRouteHint returned error: %v", err)
	}
	if ok || host != "" {
		t.Fatalf("expected non-hint packet to be ignored, got host=%q ok=%v", host, ok)
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
