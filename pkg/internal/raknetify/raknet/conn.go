package raknet

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"go.minekube.com/gate/pkg/internal/raknetify/raknet/internal/message"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// currentProtocol is the RakNet protocol version used by Raknetify's
	// netty-raknet transport. Raknetify accepts 9 and 10 and defaults clients to
	// 9, while Bedrock's current RakNet protocol is 11.
	currentProtocol byte = 9

	maxMTUSize    = 1400
	maxWindowSize = 2048

	// fragmentTimeout is the expiry for incomplete unreliable fragments (3s).
	// Matches netty-raknet FrameJoiner default.
	fragmentTimeout = time.Second * 3

	// reliableFragmentTimeout caps reliable fragment lifetime at 120s.
	// Unlike netty-raknet which cleans up via channel close, Gate needs this
	// because connections can be long-lived passthrough pipes.
	reliableFragmentTimeout = time.Second * 120

	// maxPendingFragmentBytes caps total pending fragment data at 4 MiB.
	// Matches netty-raknet FrameJoiner maxPendingFragmentBytes.
	maxPendingFragmentBytes = 4 * 1024 * 1024

	// maxFragmentComponents caps the split count of a fragmented packet.
	// Matches netty-raknet FrameJoiner MAX_FRAGMENT_COMPONENTS.
	maxFragmentComponents = 16384

	// maxPendingBuilders caps the number of incomplete fragmented packet builders.
	// Matches netty-raknet FrameJoiner DEFAULT_MAX_PENDING_BUILDERS.
	maxPendingBuilders = 256

	// defaultPingInterval is the default ping interval used when RTT hasn't been measured yet.
	defaultPingInterval = time.Millisecond * 200

	// minPingInterval is the minimum adaptive ping interval.
	minPingInterval = time.Millisecond * 50

	// maxPingInterval is the maximum adaptive ping interval.
	maxPingInterval = time.Millisecond * 500

	// tickInterval is the interval at which the connection ticks.
	tickInterval = time.Millisecond * 50

	// maxMissedPongs is the number of missed pongs before considering the connection dead.
	maxMissedPongs = 10

	// raknetifyRetryDelay is a small constant delay added to the retry timeout.
	raknetifyRetryDelay = time.Millisecond * 10

	// ackFlushDelay is the delay before a deferred ACK flush fires.
	// Matches netty-raknet ACK_FLUSH_DELAY_NANOS (2ms + 0.5ms grace).
	ackFlushDelay = 2500 * time.Microsecond
)

// Conn represents a connection to a specific client. It is not a real
// connection, as UDP is connectionless, but rather a connection emulated using
// RakNet. Methods may be called on Conn from multiple goroutines
// simultaneously.
type Conn struct {
	// rtt is the last measured round-trip time between both ends of the
	// connection. The rtt is measured in nanoseconds.
	rtt atomic.Int64

	closing atomic.Int64

	conn   net.PacketConn
	addr   net.Addr
	limits bool

	once              sync.Once
	closed, connected chan struct{}
	close             func()

	mu  sync.Mutex
	buf *bytes.Buffer

	ackBuf, nackBuf *bytes.Buffer

	pk *packet

	seq, messageIndex uint24
	orderIndex        [256]uint24
	sequenceIndex     [256]uint24
	splitID           uint32

	// mtuSize is the MTU size of the connection. Packets longer than this size
	// must be split into fragments for them to arrive at the client without
	// losing bytes.
	mtuSize uint16

	// splits is a map of slices indexed by split IDs. The length of each of the
	// slices is equal to the split count, and packets are positioned in that
	// slice indexed by the split index.
	splits map[uint16][][]byte

	// win is an ordered queue used to track which datagrams were received and
	// which datagrams were missing, so that we can send NACKs to request
	// missing datagrams.
	win *datagramWindow

	ackMu sync.Mutex
	// ackSlice is a slice containing sequence numbers of datagrams that were
	// received over the last second. When ticked, all of these packets are sent
	// in an ACK and the slice is cleared.
	ackSlice []uint24

	// packetQueues are ordered queues containing packets indexed by their order
	// index, scoped per RakNet order channel.
	packetQueues [256]*packetQueue
	// packets is a channel containing content of packets that were fully
	// processed. Calling Conn.Read() consumes a value from this channel.
	packets chan *Frame

	// retransmission is a queue filled with packets that were sent with a given
	// datagram sequence number.
	retransmission *resendMap

	// readDeadline is a channel that receives a time.Time after a specific
	// time. It is used to listen for timeouts in Read after calling
	// SetReadDeadline.
	readDeadline <-chan time.Time

	lastActivity atomic.Pointer[time.Time]

	// hasRTT is set true when the first real RTT sample arrives via an ACK.
	// Used to distinguish the unreliabled 50ms fallback in rtt() from actual measurements.
	hasRTT atomic.Bool

	// lastPongAt stores the time of the last pong OR ACK/NACK received.
	// ACK/NACK packets are treated as liveness signals so connections receiving
	// only ACK/NACK traffic (no pongs) stay alive.
	lastPongAt atomic.Pointer[time.Time]

	// hasPong is set true when the first pong is received. Used to double the
	// initial dead-detection grace period before any pong arrives.
	hasPong atomic.Bool

	// splitReliabilities tracks the reliability of each pending split by its
	// split ID. Used by cleanupExpiredFragments to apply different timeouts
	// for reliable vs unreliable fragments.
	splitReliabilities map[uint16]byte

	// splitTimes tracks the creation time of each pending split by its split ID.
	splitTimes map[uint16]time.Time

	// pendingFragmentBytes is the total bytes of all pending fragment data across
	// all incomplete splits. Capped at maxPendingFragmentBytes.
	pendingFragmentBytes int

	// firstActivityAt records when the connection was created.
	firstActivityAt time.Time

	// ackFlushTimer holds a timer for deferred ACK flushing.
	// When non-nil, a deferred ACK flush is scheduled.
	ackFlushTimer *time.Timer

	// receiveQueue is an optional buffered channel for queuing inbound datagrams
	// for ordered processing. When nil, receive() is called directly.
	receiveQueue chan []byte
}

// newConn constructs a new connection specifically dedicated to the address
// passed.
func newConn(conn net.PacketConn, addr net.Addr, mtuSize uint16) *Conn {
	return newConnWithLimits(conn, addr, mtuSize, true)
}

// newConnWithLimits returns a Conn for the net.Addr passed with a specific mtu
// size. The limits bool passed specifies if the connection should limit the
// bounds of things such as the size of packets. This is generally recommended
// for connections coming from a client.
func newConnWithLimits(conn net.PacketConn, addr net.Addr, mtuSize uint16, limits bool) *Conn {
	if mtuSize < 500 || mtuSize > 1500 {
		mtuSize = maxMTUSize
	}
	c := &Conn{
		addr:           addr,
		conn:           conn,
		limits:         limits,
		mtuSize:        mtuSize,
		pk:             new(packet),
		closed:         make(chan struct{}),
		connected:      make(chan struct{}),
		packets:        make(chan *Frame, 512),
		splits:         make(map[uint16][][]byte),
		win:            newDatagramWindow(),
		retransmission: newRecoveryQueue(),
		buf:            bytes.NewBuffer(make([]byte, 0, mtuSize)),
		ackBuf:         bytes.NewBuffer(make([]byte, 0, 256)),
		nackBuf:        bytes.NewBuffer(make([]byte, 0, 256)),
	}
	t := time.Now()
	c.lastActivity.Store(&t)
	c.lastPongAt.Store(&t)
	c.splitReliabilities = make(map[uint16]byte)
	c.splitTimes = make(map[uint16]time.Time)
	c.firstActivityAt = t
	go c.startTicking()
	return c
}

// startTicking makes the connection start ticking, sending ACKs and pings to
// the other end where necessary and checking if the connection should be timed
// out.
func (conn *Conn) startTicking() {
	var (
		ticker    = time.NewTicker(tickInterval)
		i         int64
		acksLeft  int
		lastPing  time.Time
	)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			i++
			conn.flushACKs()

			// Adaptive ping: send pings at a rate based on measured RTT.
			rtt := time.Duration(conn.rtt.Load())
			if !conn.hasRTT.Load() {
				rtt = defaultPingInterval
			}
			interval := maxDuration(minDuration(rtt, maxPingInterval), minPingInterval)
			if time.Since(lastPing) >= interval {
				conn.sendPing()
				lastPing = t
			}
			if i%3 == 0 {
				conn.checkResend(t)
			}
			// Fragment cleanup every 10 ticks (500ms).
			if i%10 == 0 {
				conn.cleanupExpiredFragments()
			}
			if i%5 == 0 {
				// Dead connection detection: check both lastPongAt and lastActivity.
				// Floor effective interval to defaultPingInterval (200ms) so dead
				// detection is never more aggressive than 200ms * maxMissed, matching
				// netty-raknet PingProducer DEFAULT_INTERVAL_MILLIS floor.
				effectiveInterval := maxDuration(time.Duration(conn.rtt.Load()), defaultPingInterval)
				multiplier := time.Duration(maxMissedPongs)
				if !conn.hasPong.Load() {
					// No pong received yet: double the grace period (2x maxMissedPongs),
					// matching netty-raknet firstPingNanos path.
					multiplier *= 2
				}
				missedPongDeadline := multiplier * effectiveInterval
				lastPong := conn.lastPongAt.Load()
				lastActivity := conn.lastActivity.Load()
				now := time.Now()
				pongExpired := lastPong != nil && now.Sub(*lastPong) > missedPongDeadline
				activityExpired := lastActivity != nil && now.Sub(*lastActivity) > missedPongDeadline
				if pongExpired && activityExpired {
					_ = conn.Close()
				}
			}
			if unix := conn.closing.Load(); unix != 0 {
				before := acksLeft
				conn.mu.Lock()
				acksLeft = len(conn.retransmission.unacknowledged)
				conn.mu.Unlock()

				if before != 0 && acksLeft == 0 {
					_ = conn.Close()
				}

				since := time.Since(time.Unix(unix, 0))
				if (acksLeft == 0 && since > time.Second) || since > time.Second*8 {
					conn.closeImmediately()
				}
			}
		case <-conn.closed:
			return
		}
	}
}

// flushACKs flushes all pending datagram acknowledgements.
func (conn *Conn) flushACKs() {
	conn.ackMu.Lock()
	defer conn.ackMu.Unlock()

	if conn.ackFlushTimer != nil {
		conn.ackFlushTimer.Stop()
		conn.ackFlushTimer = nil
	}

	if len(conn.ackSlice) > 0 {
		// Write an ACK packet to the connection containing all datagram
		// sequence numbers that we received since the last tick.
		if err := conn.sendACK(conn.ackSlice...); err != nil {
			return
		}
		conn.ackSlice = conn.ackSlice[:0]
	}
}

// scheduleDeferredACKFlush schedules a one-shot timer to flush pending ACKs
// after ackFlushDelay. This ensures low-traffic connections don't have to wait
// for the full 50ms tick before the peer receives ACKs.
func (conn *Conn) scheduleDeferredACKFlush() {
	if conn.ackFlushTimer != nil {
		conn.ackFlushTimer.Stop()
	}
	conn.ackFlushTimer = time.AfterFunc(ackFlushDelay, func() {
		conn.flushACKs()
	})
}

// checkResend checks if the connection needs to resend any packets. It sends
// an ACK for packets it has received and sends any packets that have been
// pending for too long.
func (conn *Conn) checkResend(now time.Time) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	rtt := conn.retransmission.rtt()
	sd := conn.retransmission.rttStdDev()
	baseTimeout := rtt + 2*sd + raknetifyRetryDelay
	rttNanos := int64(rtt)
	conn.rtt.Store(rttNanos)

	resend := make([]uint24, 0)
	for seq, record := range conn.retransmission.unacknowledged {
		// Exponential backoff: 1x, 2x, 4x, 8x (capped at 8x).
		multiplier := time.Duration(1 << min(record.retryCount, 3))
		timeout := baseTimeout * multiplier
		if now.Sub(record.timestamp) > timeout {
			resend = append(resend, seq)
		}
	}
	_ = conn.resend(resend)
}

// Write writes a buffer b over the RakNet connection. The amount of bytes
// written n is always equal to the length of the bytes written if writing was
// successful. If not, an error is returned and n is 0. Write may be called
// simultaneously from multiple goroutines, but will write one by one.
func (conn *Conn) Write(b []byte) (n int, err error) {
	return conn.WriteFrame(&Frame{Payload: b, Reliability: ReliabilityReliableOrdered})
}

// WriteFrame writes a RakNet user frame over the connection. It preserves the
// frame's reliability and order channel, which Raknetify uses for
// multi-channel traffic.
func (conn *Conn) WriteFrame(frame *Frame) (n int, err error) {
	select {
	case <-conn.closed:
		return 0, conn.wrap(net.ErrClosed, "write")
	default:
		conn.mu.Lock()
		defer conn.mu.Unlock()
		n, err := conn.writeFrame(frame)
		return n, conn.wrap(err, "write")
	}
}

// RaknetifySyncFrame returns a Raknetify synchronization frame describing the
// current outgoing order/sequence state of this connection. Raknetify uses this
// frame when it resets protocol state during joins, respawns, and
// reconfiguration. A proxy that terminates RakNet on both sides must generate
// this from the destination connection instead of forwarding the peer's sync
// payload verbatim.
func (conn *Conn) RaknetifySyncFrame() *Frame {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	ignored := map[byte]struct{}{1: {}}
	channels := make([]byte, 0, 7)
	for ch := byte(0); ch < 8; ch++ {
		if _, ok := ignored[ch]; ok {
			continue
		}
		channels = append(channels, ch)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1+1+len(channels)*5+4))
	buf.WriteByte(0xfc)
	buf.WriteByte(byte(len(channels)))
	for _, ch := range channels {
		buf.WriteByte(ch)
		_ = binary.Write(buf, binary.BigEndian, int32(conn.orderIndex[ch]))
	}
	_ = binary.Write(buf, binary.BigEndian, int32(conn.seq))

	return &Frame{
		Payload:     buf.Bytes(),
		Reliability: ReliabilityReliable,
	}
}

func (conn *Conn) writeFrame(frame *Frame) (n int, err error) {
	if frame == nil {
		return 0, nil
	}
	reliability := byte(frame.Reliability)
	switch reliability {
	case reliabilityUnreliable,
		reliabilityUnreliableSequenced,
		reliabilityReliable,
		reliabilityReliableOrdered,
		reliabilityReliableSequenced:
	default:
		reliability = reliabilityReliableOrdered
	}
	orderChannel := frame.OrderChannel
	fragments := conn.split(frame.Payload)
	orderIndex := conn.orderIndex[orderChannel]
	if packetNeedsOrderIndex(reliability) {
		conn.orderIndex[orderChannel]++
	}
	sequenceIndex := conn.sequenceIndex[orderChannel]
	if packetNeedsSequenceIndex(reliability) {
		conn.sequenceIndex[orderChannel]++
	}

	splitID := uint16(conn.splitID)
	split := len(fragments) > 1
	if split {
		conn.splitID++
	}
	for splitIndex, content := range fragments {
		sequenceNumber := conn.seq
		conn.seq++
		messageIndex := conn.messageIndex
		conn.messageIndex++

		conn.buf.WriteByte(bitFlagDatagram | bitFlagNeedsBAndAS)
		writeUint24(conn.buf, sequenceNumber)
		pk := packetPool.Get().(*packet)
		if cap(pk.content) < len(content) {
			pk.content = make([]byte, len(content))
		}
		// We set the actual slice size to the same size as the content. It
		// might be bigger than the previous size, in which case it will grow,
		// which is fine as the underlying array will always be big enough.
		pk.content = pk.content[:len(content)]
		copy(pk.content, content)

		pk.reliability = reliability
		pk.orderIndex = orderIndex
		pk.orderChannel = orderChannel
		pk.sequenceIndex = sequenceIndex
		pk.messageIndex = messageIndex

		pk.split = split
		if split {
			// If there were more than one fragment, the pk was split, so we
			// need to make sure we set the appropriate fields.
			pk.splitCount = uint32(len(fragments))
			pk.splitIndex = uint32(splitIndex)
			pk.splitID = splitID
		}
		pk.write(conn.buf)
		// We then send the pk to the connection.
		if _, err := conn.conn.WriteTo(conn.buf.Bytes(), conn.addr); err != nil {
			return 0, net.ErrClosed
		}

		// We reset the buffer so that we can re-use it for each fragment
		// created when splitting the packet.
		conn.buf.Reset()

		// Finally we add the pk to the recovery queue.
		conn.retransmission.add(sequenceNumber, pk)
		n += len(content)
	}
	return
}

// Read reads from the connection into the byte slice passed. If successful,
// the amount of bytes read n is returned, and the error returned will be nil.
// Read blocks until a packet is received over the connection, or until the
// session is closed or the read times out, in which case an error is returned.
func (conn *Conn) Read(b []byte) (n int, err error) {
	select {
	case frame := <-conn.packets:
		if len(b) < len(frame.Payload) {
			err = conn.wrap(errBufferTooSmall, "read")
		}
		return copy(b, frame.Payload), err
	case <-conn.closed:
		return 0, conn.wrap(net.ErrClosed, "read")
	case <-conn.readDeadline:
		return 0, conn.wrap(context.DeadlineExceeded, "read")
	}
}

// ReadPacket attempts to read the next packet as a byte slice. ReadPacket
// blocks until a packet is received over the connection, or until the session
// is closed or the read times out, in which case an error is returned.
func (conn *Conn) ReadPacket() (b []byte, err error) {
	frame, err := conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	return frame.Payload, nil
}

// ReadFrame attempts to read the next RakNet user frame. ReadFrame blocks
// until a frame is received over the connection, or until the session is closed
// or the read times out, in which case an error is returned.
func (conn *Conn) ReadFrame() (frame *Frame, err error) {
	select {
	case frame := <-conn.packets:
		return frame, nil
	case <-conn.closed:
		return nil, conn.wrap(net.ErrClosed, "read")
	case <-conn.readDeadline:
		return nil, conn.wrap(context.DeadlineExceeded, "read")
	}
}

// Close closes the connection. All blocking Read or Write actions are
// cancelled and will return an error, as soon as the closing of the connection
// is acknowledged by the client.
func (conn *Conn) Close() error {
	conn.closing.CompareAndSwap(0, time.Now().Unix())
	return nil
}

// closeImmediately sends a Disconnect notification to the other end of the
// connection and closes the underlying UDP connection immediately.
func (conn *Conn) closeImmediately() {
	conn.once.Do(func() {
		_, _ = conn.Write([]byte{message.IDDisconnectNotification})
		close(conn.closed)
		if conn.close != nil {
			conn.close()
			conn.close = nil
		}
	})
}

// RemoteAddr returns the remote address of the connection, meaning the address
// this connection leads to.
func (conn *Conn) RemoteAddr() net.Addr {
	return conn.addr
}

// LocalAddr returns the local address of the connection, which is always the
// same as the listener's.
func (conn *Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

// SetReadDeadline sets the read deadline of the connection. An error is
// returned only if the time passed is before time.Now(). Calling
// SetReadDeadline means the next Read call that exceeds the deadline will fail
// and return an error. Setting the read deadline to the default value of
// time.Time removes the deadline.
func (conn *Conn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		conn.readDeadline = make(chan time.Time)
		return nil
	}
	if t.Before(time.Now()) {
		panic(fmt.Errorf("read deadline cannot be before now"))
	}
	conn.readDeadline = time.After(time.Until(t))
	return nil
}

// SetWriteDeadline has no behaviour. It is merely there to satisfy the
// net.Conn interface.
func (conn *Conn) SetWriteDeadline(time.Time) error {
	return nil
}

// SetDeadline sets the deadline of the connection for both Read and Write.
// SetDeadline is equivalent to calling both SetReadDeadline and
// SetWriteDeadline.
func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.SetReadDeadline(t)
}

// Latency returns a rolling average of rtt between the sending and the
// receiving end of the connection. The rtt returned is updated continuously
// and is half the average round trip time (RTT).
func (conn *Conn) Latency() time.Duration {
	return time.Duration(conn.rtt.Load() / 2)
}

// sendPing pings the connection, updating the rtt of the Conn if successful.
func (conn *Conn) sendPing() {
	b := bytes.NewBuffer(nil)
	(&message.ConnectedPing{ClientTimestamp: timestamp()}).Write(b)
	_, _ = conn.Write(b.Bytes())
}

// packetPool is a sync.Pool used to pool packets that encapsulate their
// content.
var packetPool = sync.Pool{
	New: func() interface{} {
		return &packet{reliability: reliabilityReliableOrdered}
	},
}

const (
	// Datagram header +
	// Datagram sequence number +
	// Packet header +
	// Packet content length +
	// Packet message index +
	// Packet order index +
	// Packet order channel
	packetAdditionalSize = 1 + 3 + 1 + 2 + 3 + 3 + 1
	// Packet split count +
	// Packet split ID +
	// Packet split index
	splitAdditionalSize = 4 + 2 + 4
)

// split splits a content buffer in smaller buffers so that they do not exceed
// the MTU size that the connection holds.
func (conn *Conn) split(b []byte) [][]byte {
	maxSize := int(conn.mtuSize-packetAdditionalSize) - 28
	contentLength := len(b)
	if contentLength > maxSize {
		// If the content size is bigger than the maximum size here, it means
		// the packet will get split. This means that the packet will get even
		// bigger because a split packet uses 4 + 2 + 4 more bytes.
		maxSize -= splitAdditionalSize
	}
	fragmentCount := contentLength / maxSize
	if contentLength%maxSize != 0 {
		// If the content length can't be divided by maxSize perfectly, we need
		// to reserve another fragment for the last bit of the packet.
		fragmentCount++
	}
	fragments := make([][]byte, fragmentCount)

	buf := bytes.NewBuffer(b)
	for i := 0; i < fragmentCount; i++ {
		// Take a piece out of the content with the size of maxSize.
		fragments[i] = buf.Next(maxSize)
	}
	return fragments
}

// receive receives a packet from the connection, handling it as appropriate.
// If not successful, an error is returned.
func (conn *Conn) receive(b *bytes.Buffer) error {
	headerFlags, err := b.ReadByte()
	if err != nil {
		return fmt.Errorf("error reading datagram header flags: %v", err)
	}
	if headerFlags&bitFlagDatagram == 0 {
		// Ignore packets that do not have the datagram bitflag.
		return nil
	}
	t := time.Now()
	conn.lastActivity.Store(&t)
	switch {
	case headerFlags&bitFlagACK != 0:
		return conn.handleACK(b)
	case headerFlags&bitFlagNACK != 0:
		return conn.handleNACK(b)
	default:
		return conn.receiveDatagram(b)
	}
}

// receiveDatagram handles the receiving of a datagram found in buffer b. If
// successful, all packets inside the datagram are handled. if not, an error is
// returned.
func (conn *Conn) receiveDatagram(b *bytes.Buffer) error {
	seq, err := readUint24(b)
	if err != nil {
		return fmt.Errorf("error reading datagram sequence number: %v", err)
	}
	conn.ackMu.Lock()
	wasEmpty := len(conn.ackSlice) == 0
	// Add this sequence number to the received datagrams, so that it is
	// included in an ACK.
	conn.ackSlice = append(conn.ackSlice, seq)
	if wasEmpty && len(conn.ackSlice) > 0 {
		conn.scheduleDeferredACKFlush()
	}
	conn.ackMu.Unlock()

	if !conn.win.new(seq) {
		// Datagram was already received, this might happen if a packet took a long time to arrive, and we already sent
		// a NACK for it. This is expected to happen sometimes under normal circumstances, so no reason to return an
		// error.
		return nil
	}
	conn.win.add(seq)
	if conn.win.shift() == 0 {
		// Datagram window couldn't be shifted up, so we're still missing
		// packets.
		rtt := time.Duration(conn.rtt.Load())
		if missing := conn.win.missing(rtt + rtt/2); len(missing) > 0 {
			if err = conn.sendNACK(missing); err != nil {
				return fmt.Errorf("error sending NACK to request datagrams: %v", err)
			}
		}
	}
	if conn.win.size() > maxWindowSize && conn.limits {
		return fmt.Errorf("datagram receive queue window size is too big (%v-%v)", conn.win.lowest, conn.win.highest)
	}
	return conn.handleDatagram(b)
}

// handleDatagram handles the contents of a datagram encoded in a bytes.Buffer.
func (conn *Conn) handleDatagram(b *bytes.Buffer) error {
	for b.Len() > 0 {
		if err := conn.pk.read(b); err != nil {
			return fmt.Errorf("error decoding datagram packet: %v", err)
		}
		handle := conn.receivePacket
		if conn.pk.split {
			handle = conn.receiveSplitPacket
		}
		if err := handle(conn.pk); err != nil {
			return fmt.Errorf("error handling packet in datagram: %v", err)
		}
	}
	return nil
}

// receivePacket handles the receiving of a packet. It puts the packet in the
// queue and takes out all packets that were obtainable after that, and handles
// them. Unreliable ordered/sequenced packets use gap timeout to unblock
// head-of-line when missing datagrams are not retransmitted.
func (conn *Conn) receivePacket(packet *packet) error {
	needsOrder := packetNeedsOrderIndex(packet.reliability)
	if !needsOrder {
		return conn.handleFrame(packet.frame())
	}
	queue := conn.packetQueues[packet.orderChannel]
	if queue == nil {
		queue = newPacketQueue()
		conn.packetQueues[packet.orderChannel] = queue
	}

	isReliable := packetReliable(packet.reliability)
	if isReliable {
		if !queue.put(packet.orderIndex, packet.frame()) {
			return nil
		}
	} else {
		// Unreliable ordered/sequenced: track gap timeout to prevent indefinite
		// head-of-line blocking. Matches netty-raknet FrameOrderIn gap timeout.
		inserted, _ := queue.putUnreliable(packet.orderIndex, packet.frame())
		if !inserted {
			return nil
		}
		// Gate: use 2x RTT as gap timeout (one retransmission cycle).
		rtt := time.Duration(conn.rtt.Load())
		gapTimeout := maxDuration(rtt*2, 100*time.Millisecond)
		if queue.gapSince() > gapTimeout && queue.WindowSize() > 0 {
			if gapFrames := queue.flushGap(); len(gapFrames) > 0 {
				for _, frame := range gapFrames {
					if err := conn.handleFrame(frame); err != nil {
						return fmt.Errorf("error handling flushed gap frame: %v", err)
					}
				}
			}
		}
	}
	if queue.WindowSize() > maxWindowSize && conn.limits {
		return fmt.Errorf("packet queue window size is too big on channel %v (%v-%v)", packet.orderChannel, queue.lowest, queue.highest)
	}
	for _, frame := range queue.fetch() {
		if err := conn.handleFrame(frame); err != nil {
			return fmt.Errorf("error handling packet: %v", err)
		}
	}
	return nil
}

// handleFrame handles a frame serialised in byte slice b. If not successful,
// an error is returned. If the packet was not handled by RakNet, it is sent to
// the packet channel.
func (conn *Conn) handleFrame(frame *Frame) error {
	buffer := bytes.NewBuffer(frame.Payload)
	id, err := buffer.ReadByte()
	if err != nil {
		return fmt.Errorf("error reading packet ID: %v", err)
	}

	switch id {
	case message.IDConnectionRequest:
		return conn.handleConnectionRequest(buffer)
	case message.IDConnectionRequestAccepted:
		return conn.handleConnectionRequestAccepted(buffer)
	case message.IDNewIncomingConnection:
		select {
		case <-conn.connected:
		default:
			close(conn.connected)
		}
	case message.IDConnectedPing:
		return conn.handleConnectedPing(buffer)
	case message.IDConnectedPong:
		return conn.handleConnectedPong(buffer)
	case message.IDDisconnectNotification:
		conn.closeImmediately()
	case message.IDDetectLostConnections:
		// Let the other end know the connection is still alive.
		conn.sendPing()
	default:
		// Insert the packet contents the packet queue could release in the
		// channel so that Conn.Read() can get a hold of them, but always first
		// try to escape if the connection was closed.
		select {
		case <-conn.closed:
		case conn.packets <- frame:
		}
	}
	return nil
}

// handleConnectedPing handles a connected ping packet inside of buffer b. An
// error is returned if the packet was invalid.
func (conn *Conn) handleConnectedPing(b *bytes.Buffer) error {
	packet := &message.ConnectedPing{}
	if err := packet.Read(b); err != nil {
		return fmt.Errorf("error reading connected ping: %v", err)
	}
	b.Reset()

	// Respond with a connected pong that has the ping timestamp found in the
	// connected ping, and our own timestamp for the pong timestamp.
	(&message.ConnectedPong{ClientTimestamp: packet.ClientTimestamp, ServerTimestamp: timestamp()}).Write(b)
	_, err := conn.Write(b.Bytes())
	return err
}

// handleConnectedPong handles a connected pong packet inside of buffer b. An
// error is returned if the packet was invalid.
func (conn *Conn) handleConnectedPong(b *bytes.Buffer) error {
	packet := &message.ConnectedPong{}
	if err := packet.Read(b); err != nil {
		return fmt.Errorf("error reading connected pong: %v", err)
	}
	if packet.ClientTimestamp > timestamp() {
		return fmt.Errorf("error measuring rtt: ping timestamp is in the future")
	}
	t := time.Now()
	conn.lastPongAt.Store(&t)
	conn.hasPong.Store(true)
	return nil
}

// handleConnectionRequest handles a connection request packet inside of buffer
// b. An error is returned if the packet was invalid.
func (conn *Conn) handleConnectionRequest(b *bytes.Buffer) error {
	packet := &message.ConnectionRequest{}
	if err := packet.Read(b); err != nil {
		return fmt.Errorf("error reading connection request: %v", err)
	}
	b.Reset()
	(&message.ConnectionRequestAccepted{ClientAddress: *conn.addr.(*net.UDPAddr), RequestTimestamp: packet.RequestTimestamp, AcceptedTimestamp: timestamp()}).Write(b)
	_, err := conn.Write(b.Bytes())
	return err
}

// handleConnectionRequestAccepted handles a serialised connection request
// accepted packet in b, and returns an error if not successful.
func (conn *Conn) handleConnectionRequestAccepted(b *bytes.Buffer) error {
	packet := &message.ConnectionRequestAccepted{}
	_ = packet.Read(b)
	b.Reset()

	(&message.NewIncomingConnection{ServerAddress: *conn.addr.(*net.UDPAddr), RequestTimestamp: packet.RequestTimestamp, AcceptedTimestamp: packet.AcceptedTimestamp, SystemAddresses: packet.SystemAddresses}).Write(b)
	_, err := conn.Write(b.Bytes())

	select {
	case <-conn.connected:
	default:
		close(conn.connected)
	}
	return err
}

// receiveSplitPacket handles a passed split packet. If it is the last split
// packet of its sequence, it will continue handling the full packet as it
// otherwise would. An error is returned if the packet was not valid.
func (conn *Conn) receiveSplitPacket(p *packet) error {
	// Validate split parameters to prevent overflow/malformed DoS.
	// Matches netty-raknet FrameJoiner.decode validation.
	if p.splitCount <= 0 || p.splitCount > maxFragmentComponents {
		return fmt.Errorf("invalid split count %v (max %v)", p.splitCount, maxFragmentComponents)
	}
	if p.splitIndex >= p.splitCount {
		return fmt.Errorf("split index %v out of range [0, %v)", p.splitIndex, p.splitCount)
	}
	if uint64(p.splitCount)*uint64(maxMTUSize) > maxPendingFragmentBytes {
		return fmt.Errorf("fragmented frame total size exceeds maximum")
	}

	// All split state (splits, splitTimes, splitReliabilities, pendingFragmentBytes)
	// is protected by conn.mu. cleanupExpiredFragments runs on the ticker goroutine
	// and holds conn.mu while iterating and releasing splits. We must hold the
	// same lock to avoid concurrent map read/write crashes.
	conn.mu.Lock()
	unlocked := false
	defer func() {
		if !unlocked {
			conn.mu.Unlock()
		}
	}()

	m, ok := conn.splits[p.splitID]
	if !ok {
		// Enforce pending builder count limit only for new splits.
		if conn.limits && len(conn.splits) >= maxPendingBuilders {
			return fmt.Errorf("pending fragment builders exceeded: %d >= %d", len(conn.splits), maxPendingBuilders)
		}
		capped := p.splitCount
		if capped > maxFragmentComponents {
			capped = maxFragmentComponents
		}
		m = make([][]byte, capped)
		conn.splits[p.splitID] = m
		conn.splitTimes[p.splitID] = time.Now()
		conn.splitReliabilities[p.splitID] = p.reliability
	}
	if p.splitIndex > uint32(len(m)-1) {
		return fmt.Errorf("error handing split packet: split index %v is out of range (0 - %v)", p.splitIndex, len(m)-1)
	}
	// Enforce pending fragment byte limit.
	if conn.limits && conn.pendingFragmentBytes+len(p.content) > maxPendingFragmentBytes {
		return fmt.Errorf("pending fragment bytes exceeded: %d + %d > %d", conn.pendingFragmentBytes, len(p.content), maxPendingFragmentBytes)
	}
	if m[p.splitIndex] != nil {
		// Duplicate fragment; don't double-count.
		return nil
	}
	m[p.splitIndex] = p.content
	conn.pendingFragmentBytes += len(p.content)

	// Refresh the creation time on each fragment arrival.
	// Matches netty-raknet FrameJoiner Builder.add behavior.
	conn.splitTimes[p.splitID] = time.Now()

	size := 0
	for _, fragment := range m {
		if len(fragment) == 0 {
			return nil
		}
		size += len(fragment)
	}
	content := make([]byte, 0, size)
	for _, fragment := range m {
		content = append(content, fragment...)
	}
	// releaseSplitLocked must be called with conn.mu held.
	conn.releaseSplitLocked(p.splitID)

	// Release conn.mu before calling receivePacket to avoid holding the lock
	// across the downstream handleFrame  packets channel send.
	conn.mu.Unlock()
	unlocked = true

	p.content = content
	return conn.receivePacket(p)
}

// cleanupExpiredFragments removes incomplete fragments that have exceeded their
// timeout. Reliable fragments use reliableFragmentTimeout; unreliable fragments
// use fragmentTimeout. On reliable fragment timeout, the connection is closed
// to avoid permanent data loss (fragments already ACKed won't be retransmitted).
func (conn *Conn) cleanupExpiredFragments() {
	conn.mu.Lock()
	var doClose bool
	now := time.Now()
	for splitID, t := range conn.splitTimes {
		reliability, tracked := conn.splitReliabilities[splitID]
		timeout := fragmentTimeout
		if tracked && packetReliable(reliability) {
			timeout = reliableFragmentTimeout
		}
		if now.Sub(t) > timeout {
			if tracked && packetReliable(reliability) {
				conn.releaseSplitLocked(splitID)
				doClose = true
				break
			}
			conn.releaseSplitLocked(splitID)
		}
	}
	conn.mu.Unlock()
	if doClose {
		_ = conn.Close()
	}
}

// releaseSplitLocked releases all state associated with a split ID.
// Must be called with conn.mu held.
func (conn *Conn) releaseSplitLocked(splitID uint16) {
	if fragments, ok := conn.splits[splitID]; ok {
		for _, frag := range fragments {
			conn.pendingFragmentBytes -= len(frag)
		}
	}
	delete(conn.splits, splitID)
	delete(conn.splitTimes, splitID)
	delete(conn.splitReliabilities, splitID)
}

// receiveOrQueue receives a datagram or queues it for ordered processing if a
// receiveQueue is configured. This is the main entry point for inbound datagrams
// from Gate's ordered processing pipeline.
func (conn *Conn) receiveOrQueue(b *bytes.Buffer) error {
	if conn.receiveQueue == nil {
		return conn.receive(b)
	}
	now := time.Now()
	conn.lastActivity.Store(&now)
	datagram := copyQueuedDatagram(b.Bytes())
	select {
	case conn.receiveQueue <- datagram:
		return nil
	case <-conn.closed:
		releaseQueuedDatagram(datagram)
		return nil
	default:
		releaseQueuedDatagram(datagram)
		conn.closeImmediately()
		return fmt.Errorf("connection receive queue is full")
	}
}

// startQueuedReceiver starts a background goroutine that reads from
// receiveQueue and calls receive(), used by Gate for ordered message processing.
func (conn *Conn) startQueuedReceiver() {
	if conn.receiveQueue == nil {
		conn.receiveQueue = make(chan []byte, 512)
	}
	go func() {
		for {
			select {
			case b := <-conn.receiveQueue:
				if b == nil {
					return
				}
				buf := bytes.NewBuffer(b)
				_ = conn.receive(buf)
				releaseQueuedDatagram(b)
			case <-conn.closed:
				return
			}
		}
	}()
}

// sendACK sends an acknowledgement packet containing the packet sequence
// numbers passed. If not successful, an error is returned.
func (conn *Conn) sendACK(packets ...uint24) error {
	defer conn.ackBuf.Reset()
	return conn.sendAcknowledgement(packets, bitFlagACK, conn.ackBuf)
}

// sendNACK sends an acknowledgement packet containing the packet sequence
// numbers passed. If not successful, an error is returned.
func (conn *Conn) sendNACK(packets []uint24) error {
	defer conn.nackBuf.Reset()
	return conn.sendAcknowledgement(packets, bitFlagNACK, conn.nackBuf)
}

// sendAcknowledgement sends an acknowledgement packet with the packets passed,
// potentially sending multiple if too many packets are passed. The bitflag is
// added to the header byte.
func (conn *Conn) sendAcknowledgement(packets []uint24, bitflag byte, buf *bytes.Buffer) error {
	ack := &acknowledgement{packets: packets}

	for len(ack.packets) != 0 {
		buf.WriteByte(bitflag | bitFlagDatagram)
		n, err := ack.write(buf, conn.mtuSize)
		if err != nil {
			panic(fmt.Sprintf("error encoding ACK packet: %v", err))
		}
		// We managed to write n packets in the ACK with this MTU size, write
		// the next of the packets in a new ACK.
		ack.packets = ack.packets[n:]
		if _, err := conn.conn.WriteTo(buf.Bytes(), conn.addr); err != nil {
			return fmt.Errorf("error sending ACK packet: %v", err)
		}
		buf.Reset()
	}
	return nil
}

// handleACK handles an acknowledgement packet from the other end of the
// connection. These mean that a datagram was successfully received by the
// other end.
func (conn *Conn) handleACK(b *bytes.Buffer) error {
	t := time.Now()
	conn.lastPongAt.Store(&t)

	conn.mu.Lock()
	defer conn.mu.Unlock()

	ack := &acknowledgement{}
	if err := ack.read(b); err != nil {
		return fmt.Errorf("error reading ACK: %v", err)
	}
	for _, sequenceNumber := range ack.packets {
		// Take out all stored packets from the recovery queue.
		p, ok := conn.retransmission.acknowledge(sequenceNumber)
		if ok {
			conn.hasRTT.Store(true)
			// Clear the packet and return it to the pool so that it may be
			// re-used.
			p.content = nil
			packetPool.Put(p)
		}
	}
	return nil
}

// handleNACK handles a negative acknowledgment packet from the other end of
// the connection. These mean that a datagram was found missing.
func (conn *Conn) handleNACK(b *bytes.Buffer) error {
	t := time.Now()
	conn.lastPongAt.Store(&t)

	conn.mu.Lock()
	defer conn.mu.Unlock()

	nack := &acknowledgement{}
	if err := nack.read(b); err != nil {
		return fmt.Errorf("error reading NACK: %v", err)
	}
	return conn.resend(nack.packets)
}

// resend sends all datagrams currently in the recovery queue with the sequence
// numbers passed.
func (conn *Conn) resend(sequenceNumbers []uint24) (err error) {
	for _, sequenceNumber := range sequenceNumbers {
		pk, retryCount, ok := conn.retransmission.retransmit(sequenceNumber)
		if !ok {
			// We could not resend this datagram. Maybe it was already resent
			// before at the request of the client. This is generally expected
			// so we just continue.
			continue
		}

		// We first write a new datagram header using a new send sequence number
		// that we find.
		if err := conn.buf.WriteByte(bitFlagDatagram | bitFlagNeedsBAndAS); err != nil {
			return fmt.Errorf("error writing recovered datagram header: %v", err)
		}
		newSeqNum := conn.seq
		conn.seq++
		writeUint24(conn.buf, newSeqNum)
		pk.write(conn.buf)

		// We then send the pk to the connection.
		if _, err := conn.conn.WriteTo(conn.buf.Bytes(), conn.addr); err != nil {
			return fmt.Errorf("error sending pk to addr %v: %v", conn.addr, err)
		}
		// We then re-add the pk to the recovery queue in case the new one gets
		// lost too, in which case we need to resend it again.
		conn.retransmission.addWithRetryCount(newSeqNum, pk, retryCount)
		conn.buf.Reset()
	}
	return nil
}

// requestConnection requests the connection from the server, provided this
// connection operates as a client. An error occurs if the request was not
// successful.
func (conn *Conn) requestConnection(id int64) error {
	b := bytes.NewBuffer(nil)
	(&message.ConnectionRequest{ClientGUID: id, RequestTimestamp: timestamp()}).Write(b)
	_, err := conn.Write(b.Bytes())
	return err
}

// queuedDatagramPool is a sync.Pool for reusing datagram byte slices in the
// receive queue.
var queuedDatagramPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, maxMTUSize)
		return &b
	},
}

// copyQueuedDatagram copies a datagram's bytes into a pooled buffer for queuing.
func copyQueuedDatagram(b []byte) []byte {
	dstp := queuedDatagramPool.Get().(*[]byte)
	dst := *dstp
	if cap(dst) < len(b) {
		dst = make([]byte, len(b))
	} else {
		dst = dst[:len(b)]
	}
	copy(dst, b)
	return dst
}

// releaseQueuedDatagram returns a queued datagram buffer to the pool.
func releaseQueuedDatagram(b []byte) {
	queuedDatagramPool.Put(&b)
}

// maxDuration returns the larger of a and b.
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// minDuration returns the smaller of a and b.
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
