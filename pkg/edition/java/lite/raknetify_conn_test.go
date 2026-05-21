package lite

import (
	"bytes"
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
			{Payload: []byte{raknetifyStreamingCompressionHandshakePacketID, 0x01}},
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
