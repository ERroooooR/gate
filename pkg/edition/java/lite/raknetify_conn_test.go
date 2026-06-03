package lite

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.minekube.com/gate/pkg/edition/java/proto/util"
	"go.minekube.com/gate/pkg/internal/raknetify/raknet"
)

type fakeRaknetPacketConn struct {
	readFrames  []*raknet.Frame
	writeFrames []*raknet.Frame
}

func (f *fakeRaknetPacketConn) Read([]byte) (int, error) { panic("Read should not be called") }

func (f *fakeRaknetPacketConn) ReadPacket() ([]byte, error) {
	frame, err := f.ReadFrame()
	if err != nil {
		return nil, err
	}
	return frame.Payload, nil
}

func (f *fakeRaknetPacketConn) ReadFrame() (*raknet.Frame, error) {
	frame := f.readFrames[0]
	f.readFrames = f.readFrames[1:]
	return frame, nil
}

func (f *fakeRaknetPacketConn) Write(p []byte) (int, error) {
	_, _ = f.WriteFrame(&raknet.Frame{
		Payload:      append([]byte(nil), p...),
		Reliability:  raknet.ReliabilityReliableOrdered,
		OrderChannel: 0,
	})
	return len(p), nil
}

func (f *fakeRaknetPacketConn) WriteFrame(frame *raknet.Frame) (int, error) {
	payload := append([]byte(nil), frame.Payload...)
	f.writeFrames = append(f.writeFrames, &raknet.Frame{
		Payload:      payload,
		Reliability:  frame.Reliability,
		OrderChannel: frame.OrderChannel,
	})
	return len(payload), nil
}

func (f *fakeRaknetPacketConn) Close() error                     { return nil }
func (f *fakeRaknetPacketConn) LocalAddr() net.Addr              { return fakeAddr("local") }
func (f *fakeRaknetPacketConn) RemoteAddr() net.Addr             { return fakeAddr("remote") }
func (f *fakeRaknetPacketConn) SetDeadline(time.Time) error      { return nil }
func (f *fakeRaknetPacketConn) SetReadDeadline(time.Time) error  { return nil }
func (f *fakeRaknetPacketConn) SetWriteDeadline(time.Time) error { return nil }

type fakeAddr string

func (a fakeAddr) Network() string { return "raknet" }
func (a fakeAddr) String() string  { return string(a) }

func TestRaknetifyConnReadUnwrapsGamePacket(t *testing.T) {
	raw := &fakeRaknetPacketConn{
		readFrames: []*raknet.Frame{
			{
				Payload:      []byte{raknetifyStreamingCompressionHandshakePacketID, 0x01},
				Reliability:  raknet.ReliabilityReliable,
				OrderChannel: 7,
			},
			{Payload: []byte{raknetifyGamePacketID, 0x00, 0x01, 0x02}},
		},
	}
	conn := newRaknetifyConn(raw)

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	require.NoError(t, err)

	var expected bytes.Buffer
	require.NoError(t, util.WriteVarInt(&expected, 3))
	expected.Write([]byte{0x00, 0x01, 0x02})
	require.Equal(t, expected.Bytes(), buf[:n])

	buffered := conn.(interface{ DrainBufferedFrames() []*raknet.Frame }).DrainBufferedFrames()
	require.Len(t, buffered, 1)
	require.Equal(t, []byte{raknetifyStreamingCompressionHandshakePacketID, 0x01}, buffered[0].Payload)
	require.Equal(t, raknet.ReliabilityReliable, buffered[0].Reliability)
	require.Equal(t, byte(7), buffered[0].OrderChannel)
	require.Empty(t, conn.(interface{ DrainBufferedFrames() []*raknet.Frame }).DrainBufferedFrames())
}

func TestRaknetifyConnCanDisableControlFrameBuffering(t *testing.T) {
	raw := &fakeRaknetPacketConn{
		readFrames: []*raknet.Frame{
			{Payload: []byte{raknetifyPingPacketID, 0x00}},
			{Payload: []byte{raknetifyGamePacketID, 0x00}},
		},
	}
	conn := newRaknetifyConn(raw)
	conn.(interface{ SetBufferedFrameCapture(bool) }).SetBufferedFrameCapture(false)

	buf := make([]byte, 2)
	_, err := conn.Read(buf)
	require.NoError(t, err)
	require.Empty(t, conn.(interface{ DrainBufferedFrames() []*raknet.Frame }).DrainBufferedFrames())
}

func TestRaknetifyConnWriteWrapsMinecraftFrames(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw)

	var stream bytes.Buffer
	require.NoError(t, util.WriteVarInt(&stream, 2))
	stream.Write([]byte{0x01, 0x02})
	require.NoError(t, util.WriteVarInt(&stream, 1))
	stream.WriteByte(0x03)

	n, err := conn.Write(stream.Bytes())
	require.NoError(t, err)
	require.Equal(t, stream.Len(), n)
	var writePackets [][]byte
	for _, frame := range raw.writeFrames {
		writePackets = append(writePackets, frame.Payload)
		require.Equal(t, raknet.ReliabilityReliableOrdered, frame.Reliability)
		require.Equal(t, byte(0), frame.OrderChannel)
	}
	require.Equal(t, [][]byte{
		{raknetifyGamePacketID, 0x01, 0x02},
		{raknetifyGamePacketID, 0x03},
	}, writePackets)
}

// ---------------------------------------------------------------------------
// Write gate after detach (matches raknetify MixinClientConnection isClosing fix)
// ---------------------------------------------------------------------------

func TestRaknetifyConnWriteAfterDetachReturnsEOF(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	// Detach the underlying frame connection.
	conn.DetachFrameConn()

	// Write should return io.EOF to prevent racing with the passthrough pipe.
	n, err := conn.Write([]byte{0x01})
	if n != 0 {
		t.Fatalf("Write after detach returned n=%d, want 0", n)
	}
	if err != io.EOF {
		t.Fatalf("Write after detach returned err=%v, want io.EOF", err)
	}

	// Verify no frames were written to the underlying connection.
	if len(raw.writeFrames) != 0 {
		t.Fatalf("Write after detach wrote %d frames, want 0", len(raw.writeFrames))
	}
}

func TestRaknetifyConnWriteBeforeDetachSucceeds(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	// Write a valid Minecraft packet: VarInt(1) + payload{0x00}.
	// The Write method requires a Minecraft-style length-prefixed frame.
	var stream bytes.Buffer
	require.NoError(t, util.WriteVarInt(&stream, 1))
	stream.WriteByte(0x00)

	n, err := conn.Write(stream.Bytes())
	if n != stream.Len() {
		t.Fatalf("Write before detach returned n=%d, want %d", n, stream.Len())
	}
	if err != nil {
		t.Fatalf("Write before detach returned err=%v, want nil", err)
	}
	if len(raw.writeFrames) != 1 {
		t.Fatalf("Write before detach wrote %d frames, want 1", len(raw.writeFrames))
	}
}

func TestRaknetifyConnReadAfterDetachReturnsEOF(t *testing.T) {
	raw := &fakeRaknetPacketConn{
		readFrames: []*raknet.Frame{
			{Payload: []byte{raknetifyGamePacketID, 0x00}},
		},
	}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	conn.DetachFrameConn()

	// Read should return io.EOF after detach (readAborted gate).
	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if n != 0 {
		t.Fatalf("Read after detach returned n=%d, want 0", n)
	}
	if err != io.EOF {
		t.Fatalf("Read after detach returned err=%v, want io.EOF", err)
	}
}

func TestRaknetifyConnCloseAfterDetachIsNoOp(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	conn.DetachFrameConn()

	// Close should be a no-op after detach —the underlying conn is preserved.
	err := conn.Close()
	if err != nil {
		t.Fatalf("Close after detach returned err=%v, want nil", err)
	}
	// The underlying fake conn doesn't track closes, but the detached flag
	// should prevent double-close issues on the real connection.
	if !conn.detached.Load() {
		t.Fatal("detached flag was not set by DetachFrameConn")
	}
}

func TestRaknetifyConnDetachReturnsFrameConn(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	fc := conn.DetachFrameConn()
	if fc == nil {
		t.Fatal("DetachFrameConn returned nil")
	}
	// The returned frame conn should be the original.
	if fc != raw {
		t.Fatal("DetachFrameConn did not return the original frame connection")
	}
}

func TestRaknetifyConnDetachIsIdempotent(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	fc1 := conn.DetachFrameConn()
	fc2 := conn.DetachFrameConn()
	if fc1 != fc2 {
		t.Fatal("repeated DetachFrameConn returned different values")
	}
	if !conn.detached.Load() {
		t.Fatal("detached flag was not set after DetachFrameConn")
	}
	// Write should still be blocked.
	_, err := conn.Write([]byte{0x02})
	if err != io.EOF {
		t.Fatalf("Write after multiple detach calls returned err=%v, want io.EOF", err)
	}
}

// TestRaknetifyConnConcurrentWriteDetach verifies that concurrent Write and
// DetachFrameConn do not race. The detached flag (atomic.Bool) is read under
// writeMu in Write and written under readMu in DetachFrameConn - different
// locks - so both the atomic store/load AND correct happens-before ordering
// are required for correctness. Run with -race.
func TestRaknetifyConnConcurrentWriteDetach(t *testing.T) {
	raw := &fakeRaknetPacketConn{}
	conn := newRaknetifyConn(raw).(*raknetifyConn)

	// Build a valid Minecraft packet: VarInt(1) + payload{0x00}.
	var stream bytes.Buffer
	require.NoError(t, util.WriteVarInt(&stream, 1))
	stream.WriteByte(0x00)
	pkt := stream.Bytes()

	var writesOK, writesEOF int
	done := make(chan struct{})

	// Goroutine 1: write repeatedly until detach cuts us off.
	go func() {
		defer close(done)
		for {
			_, err := conn.Write(pkt)
			if err == nil {
				writesOK++
			} else if err == io.EOF {
				writesEOF++
				return
			} else {
				t.Errorf("unexpected write error: %v", err)
				return
			}
		}
	}()

	// Goroutine 2: detach after a brief moment to ensure some writes have landed.
	time.Sleep(5 * time.Millisecond)
	fc := conn.DetachFrameConn()
	if fc == nil {
		t.Fatal("DetachFrameConn returned nil")
	}

	<-done

	// At least one write should have succeeded before detach.
	if writesOK == 0 {
		t.Fatal("no writes succeeded before detach")
	}
	// At least one write should have returned io.EOF after detach.
	if writesEOF == 0 {
		t.Fatal("no writes returned io.EOF after detach")
	}

	// Verify the underlying connection received exactly the successful writes.
	if len(raw.writeFrames) != writesOK {
		t.Fatalf("underlying writeFrames = %d, want %d (writesOK)", len(raw.writeFrames), writesOK)
	}
}