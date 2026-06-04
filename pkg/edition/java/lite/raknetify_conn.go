package lite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
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

	readMu         sync.Mutex
	readBuf        bytes.Buffer
	preRead        []*raknet.Frame
	capturePreRead bool
	readAborted    atomic.Bool // set by DetachFrameConn to stop the read loop

	writeMu  sync.Mutex
	writeBuf []byte

	detached atomic.Bool // set by DetachFrameConn so Close() leaves underlying conn alive
}

func newRaknetifyConn(conn raknetFrameConn) net.Conn {
	return &raknetifyConn{conn: conn, capturePreRead: true}
}

func (c *raknetifyConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	if c.readAborted.Load() {
		c.readMu.Unlock()
		return 0, io.EOF
	}
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
			if c.capturePreRead {
				c.preRead = append(c.preRead, cloneRaknetFrame(frame))
			}
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

	// Prevent writes on a detached connection —the underlying frame conn
	// has been handed off to the passthrough pipe and writing here would
	// race with copyFrames. Matches the !isClosing gate in raknetify's
	// MixinClientConnection.redirectIsOpen.
	if c.detached.Load() {
		return 0, io.EOF
	}

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
	// Acquire writeMu to serialize with DetachFrameConn and Write.
	// Prevents racing: Close reads detached=false, DetachFrameConn sets
	// detached=true and hands off conn, then Close calls c.conn.Close()
	// on the already-transferred connection.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.detached.Load() {
		return nil // underlying conn was detached, don't close it
	}
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

// DetachFrameConn stops the Minecraft read loop and returns the underlying
// RakNet frame connection. After detach, Close() is a no-op and Read() returns
// io.EOF, so the caller can take over the frame connection for passthrough.
//
// Acquires both readMu and writeMu: readMu stops new Reads, writeMu drains
// any in-flight Write before the handoff. This guarantees that by the time
// this method returns, no Read, Write, or Close on this raknetifyConn will
// touch the underlying connection — it belongs exclusively to the caller.
func (c *raknetifyConn) DetachFrameConn() raknetFrameConn {
	c.readMu.Lock()
	c.detached.Store(true)
	c.readAborted.Store(true)
	c.readMu.Unlock()

	// Drain in-flight writes: a Write may have passed the detached check
	// before we set the flag. Holding writeMu ensures it has completed
	// (or will see detached=true and return io.EOF) before we return.
	c.writeMu.Lock()
	_ = c.detached.Load()
	c.writeMu.Unlock()

	return c.conn
}

func (c *raknetifyConn) DrainBufferedFrames() []*raknet.Frame {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	frames := c.preRead
	c.preRead = nil
	c.capturePreRead = false
	return frames
}

func (c *raknetifyConn) SetBufferedFrameCapture(enabled bool) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	c.capturePreRead = enabled
	if !enabled {
		c.preRead = nil
	}
}

func cloneRaknetFrame(frame *raknet.Frame) *raknet.Frame {
	if frame == nil {
		return nil
	}
	payload := make([]byte, len(frame.Payload))
	copy(payload, frame.Payload)
	return &raknet.Frame{
		Payload:      payload,
		Reliability:  frame.Reliability,
		OrderChannel: frame.OrderChannel,
	}
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
