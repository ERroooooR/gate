package lite

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.minekube.com/gate/pkg/edition/java/proto/util"
	"go.minekube.com/gate/pkg/internal/raknetify/raknet"
)

const (
	raknetifyGamePacketID                          = byte(0xfd)
	raknetifyPingPacketID                          = byte(0xfa)
	raknetifySyncPacketID                          = byte(0xfc)
	raknetifyMetricsSyncPacketID                   = byte(0xfb)
	raknetifyStreamingCompressionPacketID          = byte(0xed)
	raknetifyStreamingCompressionHandshakePacketID = byte(0xec)
)

type raknetPacketConn interface {
	net.Conn
	ReadPacket() ([]byte, error)
}

type raknetFrameConn interface {
	net.Conn
	ReadFrame() (*raknet.Frame, error)
	WriteFrame(*raknet.Frame) (int, error)
}

type raknetSyncFrameConn interface {
	RaknetifySyncFrame() *raknet.Frame
}

// raknetifyConn adapts Raknetify's RakNet packet payloads to the vanilla Java
// TCP byte stream expected by Gate's Lite forwarding code.
type raknetifyConn struct {
	conn raknetFrameConn

	readMu  sync.Mutex
	readBuf bytes.Buffer

	writeMu  sync.Mutex
	writeBuf []byte
}

func newRaknetifyConn(conn raknetFrameConn) net.Conn {
	return &raknetifyConn{conn: conn}
}

func (c *raknetifyConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	for c.readBuf.Len() == 0 {
		frame, err := c.conn.ReadFrame()
		if err != nil {
			return 0, err
		}
		packet := frame.Payload
		if len(packet) == 0 {
			continue
		}
		switch packet[0] {
		case raknetifyGamePacketID:
			payload := packet[1:]
			if err := util.WriteVarInt(&c.readBuf, len(payload)); err != nil {
				return 0, err
			}
			_, _ = c.readBuf.Write(payload)
		case raknetifyPingPacketID,
			raknetifySyncPacketID,
			raknetifyMetricsSyncPacketID,
			raknetifyStreamingCompressionPacketID,
			raknetifyStreamingCompressionHandshakePacketID:
			// Gate Lite deliberately does not start Raknetify multi-channeling or
			// streaming compression. Control packets from the client are ignored.
			continue
		default:
			continue
		}
	}

	return c.readBuf.Read(p)
}

func (c *raknetifyConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.writeBuf = append(c.writeBuf, p...)
	consumed := 0
	for {
		length, varIntLen, ok, err := peekVarInt(c.writeBuf[consumed:])
		if err != nil {
			return 0, err
		}
		if !ok {
			break
		}
		frameEnd := consumed + varIntLen + length
		if len(c.writeBuf) < frameEnd {
			break
		}
		payload := c.writeBuf[consumed+varIntLen : frameEnd]
		out := make([]byte, 1+len(payload))
		out[0] = raknetifyGamePacketID
		copy(out[1:], payload)
		if _, err := c.conn.Write(out); err != nil {
			return 0, err
		}
		consumed = frameEnd
	}
	if consumed != 0 {
		copy(c.writeBuf, c.writeBuf[consumed:])
		c.writeBuf = c.writeBuf[:len(c.writeBuf)-consumed]
	}
	return len(p), nil
}

func peekVarInt(buf []byte) (value int, bytesRead int, complete bool, err error) {
	for bytesRead < 5 {
		if bytesRead >= len(buf) {
			return 0, bytesRead, false, nil
		}
		b := buf[bytesRead]
		value |= int(b&0x7f) << (7 * bytesRead)
		bytesRead++
		if b&0x80 == 0 {
			return value, bytesRead, true, nil
		}
	}
	return 0, bytesRead, false, fmt.Errorf("minecraft packet length VarInt is too large")
}

func (c *raknetifyConn) Close() error {
	return c.conn.Close()
}

func (c *raknetifyConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *raknetifyConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *raknetifyConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *raknetifyConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *raknetifyConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *raknetifyConn) FrameConn() raknetFrameConn {
	return c.conn
}

var _ net.Conn = (*raknetifyConn)(nil)

func dialRaknetify(ctx context.Context, address string) (net.Conn, error) {
	conn, err := dialRaknetifyFrame(ctx, address)
	if err != nil {
		return nil, err
	}
	return newRaknetifyConn(conn), nil
}

func dialRaknetifyFrame(ctx context.Context, address string) (raknetFrameConn, error) {
	conn, err := raknet.DialContext(ctx, address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
