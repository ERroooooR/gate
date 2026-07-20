package lite

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"go.minekube.com/gate/pkg/edition/java/lite/config"
)

type webSocketStreamConn struct {
	net.Conn
	remoteAddr  net.Addr
	virtualHost string
}

func newWebSocketStreamConn(
	ctx context.Context,
	wsConn *websocket.Conn,
	cfg config.WebSocketListenerConfig,
	remoteAddr net.Addr,
	virtualHost string,
	mode string,
) *webSocketStreamConn {
	stream := websocket.NetConn(ctx, wsConn, websocket.MessageBinary)
	// NetConn disables the read limit, so restore it after construction.
	wsConn.SetReadLimit(cfg.EffectiveReadLimit())
	optimized := newOptimizedWebSocketConn(stream, cfg, mode)
	return &webSocketStreamConn{Conn: optimized, remoteAddr: remoteAddr, virtualHost: virtualHost}
}

func (c *webSocketStreamConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return c.remoteAddr
	}
	return c.Conn.RemoteAddr()
}

func (c *webSocketStreamConn) WebSocketVirtualHost() string { return c.virtualHost }

type optimizedWebSocketConn struct {
	net.Conn
	cfg config.WebSocketListenerConfig

	queue chan []byte
	done  chan struct{}

	closeOnce sync.Once
	writeGate sync.RWMutex
	closing   bool
	errMu     sync.Mutex
	fatalErr  error

	pendingMu    sync.Mutex
	pendingBytes int
	spaceChanged chan struct{}

	targetFrame atomic.Int64
	lastActive  atomic.Int64
	mode        string
}

func newOptimizedWebSocketConn(conn net.Conn, cfg config.WebSocketListenerConfig, mode string) *optimizedWebSocketConn {
	c := &optimizedWebSocketConn{
		Conn:         conn,
		cfg:          cfg,
		queue:        make(chan []byte, 1024),
		done:         make(chan struct{}),
		spaceChanged: make(chan struct{}),
		mode:         mode,
	}
	c.targetFrame.Store(int64(cfg.EffectiveFramePayloadSize()))
	c.touch()
	go c.writeLoop()
	go c.idleLoop()
	return c
}

func (c *optimizedWebSocketConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.touch()
		wsmcMetrics.read(c.mode, n)
	}
	return n, err
}

func (c *optimizedWebSocketConn) Write(p []byte) (int, error) {
	c.writeGate.RLock()
	defer c.writeGate.RUnlock()
	if c.closing {
		return 0, net.ErrClosed
	}
	if len(p) == 0 {
		return 0, c.currentError()
	}
	maxReservation := c.cfg.EffectiveMaxPendingBytes()
	written := 0
	for len(p) > 0 {
		n := min(len(p), maxReservation)
		if err := c.reserve(n); err != nil {
			return written, err
		}
		copyBuf := make([]byte, n)
		copy(copyBuf, p[:n])
		select {
		case c.queue <- copyBuf:
			written += n
			p = p[n:]
		case <-c.done:
			c.release(n)
			return written, c.currentError()
		}
	}
	return written, nil
}

func (c *optimizedWebSocketConn) Close() error {
	c.writeGate.Lock()
	if c.closing {
		c.writeGate.Unlock()
		return nil
	}
	c.closing = true
	c.writeGate.Unlock()

	// Write is intentionally buffered to coalesce small Minecraft writes. Preserve
	// net.Conn's close semantics by giving the single writer time to flush them.
	timer := time.NewTimer(c.cfg.EffectiveBackpressureTimeout())
	defer timer.Stop()
	for {
		c.pendingMu.Lock()
		pending := c.pendingBytes
		changed := c.spaceChanged
		c.pendingMu.Unlock()
		if pending == 0 {
			break
		}
		select {
		case <-changed:
		case <-timer.C:
			wsmcMetrics.event(c.mode, "close_flush_timeout")
			c.fail(errors.New("websocket close flush timeout"))
			return c.currentError()
		case <-c.done:
			return c.currentError()
		}
	}
	c.fail(net.ErrClosed)
	return nil
}

func (c *optimizedWebSocketConn) reserve(size int) error {
	timeout := c.cfg.EffectiveBackpressureTimeout()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		c.pendingMu.Lock()
		if c.pendingBytes+size <= c.cfg.EffectiveMaxPendingBytes() {
			c.pendingBytes += size
			c.pendingMu.Unlock()
			return nil
		}
		changed := c.spaceChanged
		c.pendingMu.Unlock()
		select {
		case <-changed:
		case <-timer.C:
			wsmcMetrics.event(c.mode, "backpressure_timeout")
			c.fail(errors.New("websocket backpressure timeout"))
			return c.currentError()
		case <-c.done:
			return c.currentError()
		}
	}
}

func (c *optimizedWebSocketConn) release(size int) {
	c.pendingMu.Lock()
	c.pendingBytes -= size
	if c.pendingBytes < 0 {
		c.pendingBytes = 0
	}
	close(c.spaceChanged)
	c.spaceChanged = make(chan struct{})
	c.pendingMu.Unlock()
}

func (c *optimizedWebSocketConn) writeLoop() {
	var carry []byte
	for {
		var first []byte
		if carry != nil {
			first, carry = carry, nil
		} else {
			select {
			case first = <-c.queue:
			case <-c.done:
				return
			}
		}

		batch := [][]byte{first}
		batchBytes := len(first)
		window := c.cfg.EffectiveCoalesceWindow()
		var timer *time.Timer
		var timerC <-chan time.Time
		if window > 0 {
			timer = time.NewTimer(window)
			timerC = timer.C
		}
	collect:
		for batchBytes < c.cfg.EffectiveCoalesceLimit() {
			select {
			case next := <-c.queue:
				if batchBytes+len(next) > c.cfg.EffectiveCoalesceLimit() {
					carry = next
					break collect
				}
				batch = append(batch, next)
				batchBytes += len(next)
			case <-timerC:
				break collect
			case <-c.done:
				if timer != nil {
					timer.Stop()
				}
				return
			default:
				if window <= 0 {
					break collect
				}
				// Wait for either another small write or the coalescing window.
				select {
				case next := <-c.queue:
					if batchBytes+len(next) > c.cfg.EffectiveCoalesceLimit() {
						carry = next
						break collect
					}
					batch = append(batch, next)
					batchBytes += len(next)
				case <-timerC:
					break collect
				case <-c.done:
					if timer != nil {
						timer.Stop()
					}
					return
				}
			}
		}
		if timer != nil && !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		payload := make([]byte, 0, batchBytes)
		for _, part := range batch {
			payload = append(payload, part...)
		}
		started := time.Now()
		frames, err := c.writeFrames(payload)
		elapsed := time.Since(started)
		c.release(batchBytes)
		if err != nil {
			wsmcMetrics.event(c.mode, "write_failure")
			c.fail(err)
			return
		}
		c.touch()
		c.adjustFrameSize(elapsed)
		c.pendingMu.Lock()
		pending := c.pendingBytes
		c.pendingMu.Unlock()
		wsmcMetrics.write(c.mode, frames, batchBytes, pending, int(c.targetFrame.Load()), elapsed)
	}
}

func (c *optimizedWebSocketConn) writeFrames(payload []byte) (int, error) {
	frameSize := int(c.targetFrame.Load())
	frames := 0
	for len(payload) > 0 {
		n := min(len(payload), frameSize)
		written, err := c.Conn.Write(payload[:n])
		if err != nil {
			return frames, err
		}
		if written != n {
			return frames, io.ErrShortWrite
		}
		frames++
		payload = payload[n:]
	}
	return frames, nil
}

func (c *optimizedWebSocketConn) adjustFrameSize(writeDuration time.Duration) {
	if !c.cfg.EffectiveAdaptiveFraming() {
		return
	}
	c.pendingMu.Lock()
	pending := c.pendingBytes
	c.pendingMu.Unlock()
	current := int(c.targetFrame.Load())
	minFrame := c.cfg.EffectiveMinFramePayloadSize()
	maxFrame := c.cfg.EffectiveFramePayloadSize()
	if writeDuration > 20*time.Millisecond || pending > c.cfg.EffectiveMaxPendingBytes()/2 {
		current = max(minFrame, current/2)
	} else if writeDuration < 2*time.Millisecond && pending < c.cfg.EffectiveMaxPendingBytes()/4 {
		current = min(maxFrame, current*2)
	}
	c.targetFrame.Store(int64(current))
}

func (c *optimizedWebSocketConn) idleLoop() {
	idle := c.cfg.EffectiveIdleTimeout()
	interval := max(time.Second, idle/3)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			last := time.Unix(0, c.lastActive.Load())
			if now.Sub(last) >= idle {
				wsmcMetrics.event(c.mode, "idle_timeout")
				c.fail(errors.New("websocket idle timeout"))
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *optimizedWebSocketConn) touch() { c.lastActive.Store(time.Now().UnixNano()) }

func (c *optimizedWebSocketConn) fail(err error) {
	c.closeOnce.Do(func() {
		c.errMu.Lock()
		c.fatalErr = err
		c.errMu.Unlock()
		close(c.done)
		_ = c.Conn.Close()
		c.pendingMu.Lock()
		close(c.spaceChanged)
		c.spaceChanged = make(chan struct{})
		c.pendingMu.Unlock()
	})
}

func (c *optimizedWebSocketConn) currentError() error {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	if c.fatalErr != nil {
		return c.fatalErr
	}
	return net.ErrClosed
}
