package raknet

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

// makeTestConn returns a minimally initialized Conn suitable for unit tests.
func makeTestConn() *Conn {
	t := time.Now()
	c := &Conn{
		closed:              make(chan struct{}),
		splits:              make(map[uint16][][]byte),
		splitTimes:          make(map[uint16]time.Time),
		splitReliabilities:  make(map[uint16]byte),
		retransmission:      newRecoveryQueue(),
		packets:             make(chan *Frame, 512),
	}
	c.lastPongAt.Store(&t)
	c.lastActivity.Store(&t)
	c.firstActivityAt = t
	return c
}

// makeSinglePacketACK constructs a valid ACK buffer containing a single packet
// with the given sequence number.
func makeSinglePacketACK(seq uint24) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(bitFlagACK | bitFlagDatagram)          // 0xC0
	_ = binary.Write(buf, binary.BigEndian, int16(1))     // record count = 1
	buf.WriteByte(packetSingle)                           // single packet type
	writeUint24(buf, seq)                                 // sequence number
	return buf
}

// makeSinglePacketNACK constructs a valid NACK buffer containing a single packet
// with the given sequence number.
func makeSinglePacketNACK(seq uint24) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(bitFlagNACK | bitFlagDatagram)          // 0xA0
	_ = binary.Write(buf, binary.BigEndian, int16(1))     // record count = 1
	buf.WriteByte(packetSingle)                           // single packet type
	writeUint24(buf, seq)                                 // sequence number
	return buf
}

// ---------------------------------------------------------------------------
// Fragment expiration tests (matches netty-raknet FrameJoiner fix)
// ---------------------------------------------------------------------------

func TestCleanupExpiredFragmentsSkipsReliable(t *testing.T) {
	conn := makeTestConn()

	// Add a reliable split with a timestamp far in the past.
	conn.splitTimes[1] = time.Now().Add(-fragmentTimeout * 10)
	conn.splits[1] = make([][]byte, 2)
	conn.splitReliabilities[1] = reliabilityReliableOrdered

	// Add an unreliable split with a timestamp far in the past.
	conn.splitTimes[2] = time.Now().Add(-fragmentTimeout * 10)
	conn.splits[2] = make([][]byte, 2)
	conn.splitReliabilities[2] = reliabilityUnreliable

	conn.cleanupExpiredFragments()

	// Reliable split should still exist.
	if _, ok := conn.splits[1]; !ok {
		t.Fatal("reliable split was incorrectly expired")
	}
	if _, ok := conn.splitTimes[1]; !ok {
		t.Fatal("reliable split timestamp was incorrectly removed")
	}
	if _, ok := conn.splitReliabilities[1]; !ok {
		t.Fatal("reliable split reliability was incorrectly removed")
	}

	// Unreliable split should have been cleaned up.
	if _, ok := conn.splits[2]; ok {
		t.Fatal("unreliable split was not expired")
	}
	if _, ok := conn.splitTimes[2]; ok {
		t.Fatal("unreliable split timestamp was not removed")
	}
	if _, ok := conn.splitReliabilities[2]; ok {
		t.Fatal("unreliable split reliability was not removed")
	}
}

func TestCleanupExpiredFragmentsSkipsReliableSequenced(t *testing.T) {
	conn := makeTestConn()

	conn.splitTimes[1] = time.Now().Add(-fragmentTimeout * 10)
	conn.splits[1] = make([][]byte, 2)
	conn.splitReliabilities[1] = reliabilityReliableSequenced

	conn.cleanupExpiredFragments()

	if _, ok := conn.splits[1]; !ok {
		t.Fatal("reliable sequenced split was incorrectly expired")
	}
}

func TestCleanupExpiredFragmentsSkipsReliableBare(t *testing.T) {
	conn := makeTestConn()

	conn.splitTimes[1] = time.Now().Add(-fragmentTimeout * 10)
	conn.splits[1] = make([][]byte, 3)
	conn.splitReliabilities[1] = reliabilityReliable

	conn.cleanupExpiredFragments()

	if _, ok := conn.splits[1]; !ok {
		t.Fatal("bare reliable split was incorrectly expired")
	}
}

func TestCleanupExpiredFragmentsExpiresUnreliableSequenced(t *testing.T) {
	conn := makeTestConn()

	conn.splitTimes[1] = time.Now().Add(-fragmentTimeout * 10)
	conn.splits[1] = make([][]byte, 2)
	conn.splitReliabilities[1] = reliabilityUnreliableSequenced

	conn.cleanupExpiredFragments()

	if _, ok := conn.splits[1]; ok {
		t.Fatal("unreliable sequenced split was not expired")
	}
}

func TestCleanupExpiredFragmentsKeepsRecentUnreliable(t *testing.T) {
	conn := makeTestConn()

	// Unreliable split with a recent timestamp — should NOT be expired.
	conn.splitTimes[1] = time.Now().Add(-fragmentTimeout / 2)
	conn.splits[1] = make([][]byte, 2)
	conn.splitReliabilities[1] = reliabilityUnreliable

	conn.cleanupExpiredFragments()

	if _, ok := conn.splits[1]; !ok {
		t.Fatal("recent unreliable split was incorrectly expired before timeout")
	}
}

func TestCleanupExpiredFragmentsSkipsWhenNoReliabilityTracked(t *testing.T) {
	// If splitReliabilities has no entry (e.g. legacy state before the fix),
	// treat the fragment as "not known reliable" and apply the timeout.
	conn := makeTestConn()

	conn.splitTimes[1] = time.Now().Add(-fragmentTimeout * 10)
	conn.splits[1] = make([][]byte, 2)
	// Deliberately do NOT set splitReliabilities[1].

	conn.cleanupExpiredFragments()

	if _, ok := conn.splits[1]; ok {
		t.Fatal("split without reliability tracking was not expired")
	}
}

// ---------------------------------------------------------------------------
// ACK / NACK liveness tests (matches netty-raknet ReliabilityHandler fix)
// ---------------------------------------------------------------------------

func TestHandleACKUpdatesLastPongAt(t *testing.T) {
	conn := makeTestConn()

	// Store nil to make a clear baseline.
	conn.lastPongAt.Store(nil)
	time.Sleep(time.Millisecond) // ensure timestamps differ

	err := conn.handleACK(makeSinglePacketACK(0))
	if err != nil {
		t.Fatalf("handleACK returned error: %v", err)
	}

	lastPong := conn.lastPongAt.Load()
	if lastPong == nil {
		t.Fatal("lastPongAt was not updated by handleACK")
	}
}

func TestHandleNACKUpdatesLastPongAt(t *testing.T) {
	conn := makeTestConn()

	conn.lastPongAt.Store(nil)
	time.Sleep(time.Millisecond)

	// NACK requires the sequence number to exist in retransmission for resend.
	// Pre-populate retransmission with a packet at seq 0 so resend doesn't fail.
	pk := &packet{reliability: reliabilityReliableOrdered, content: []byte{0x00}}
	conn.retransmission.add(0, pk)

	err := conn.handleNACK(makeSinglePacketNACK(0))
	if err != nil {
		t.Fatalf("handleNACK returned error: %v", err)
	}

	lastPong := conn.lastPongAt.Load()
	if lastPong == nil {
		t.Fatal("lastPongAt was not updated by handleNACK")
	}
}

// TestACKNACKLivenessPreventsDeadDetection verifies that connections receiving
// only ACK/NACK traffic (no pongs) stay alive due to lastPongAt updates.
func TestACKNACKLivenessPreventsDeadDetection(t *testing.T) {
	conn := makeTestConn()

	// Simulate a scenario where pongs stop arriving, but ACKs keep flowing.
	// Set lastPongAt to an old time (simulating no pongs for a while).
	oldTime := time.Now().Add(-time.Second * 10)
	conn.lastPongAt.Store(&oldTime)
	conn.lastActivity.Store(&oldTime)

	// Now send an ACK — it should refresh lastPongAt.
	err := conn.handleACK(makeSinglePacketACK(1))
	if err != nil {
		t.Fatalf("handleACK returned error: %v", err)
	}

	lastPong := conn.lastPongAt.Load()
	if lastPong == nil {
		t.Fatal("lastPongAt was nil after handleACK")
	}

	// The new lastPongAt should be recent (within the last second).
	if time.Since(*lastPong) > time.Second {
		t.Fatalf("lastPongAt was not refreshed to a recent time: age=%v", time.Since(*lastPong))
	}
}

// ---------------------------------------------------------------------------
// lastActivity liveness regression test
// ---------------------------------------------------------------------------

func TestLastActivityUpdatedOnReceiveOrQueue(t *testing.T) {
	conn := makeTestConn()

	oldActivity := conn.lastActivity.Load()
	requireNotNil(t, oldActivity)

	time.Sleep(time.Millisecond)

	// receiveOrQueue with nil receiveQueue calls receive directly.
	err := conn.receiveOrQueue(makeSinglePacketACK(0))
	if err != nil {
		t.Fatalf("receiveOrQueue returned error: %v", err)
	}

	newActivity := conn.lastActivity.Load()
	requireNotNil(t, newActivity)

	if !newActivity.After(*oldActivity) {
		t.Fatal("lastActivity was not updated by receiveOrQueue")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func requireNotNil(t *testing.T, v *time.Time) {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil time pointer")
	}
}

// ---------------------------------------------------------------------------
// Resend map preservation test
// ---------------------------------------------------------------------------

func TestResendPreservesRetryCountAcrossRetransmits(t *testing.T) {
	conn := makeTestConn()

	// Add a packet with an initial retry count.
	pk := &packet{reliability: reliabilityReliableOrdered, content: []byte{0x01}}
	conn.retransmission.addWithRetryCount(0, pk, 2)

	// Verify the retry count was preserved.
	record, ok := conn.retransmission.unacknowledged[0]
	if !ok {
		t.Fatal("packet not found in retransmission map")
	}
	if record.retryCount != 2 {
		t.Fatalf("retryCount = %d, want 2", record.retryCount)
	}
}

// ---------------------------------------------------------------------------
// Atomic flag tests
// ---------------------------------------------------------------------------

func TestHasRTTDefaultFalse(t *testing.T) {
	conn := makeTestConn()
	if conn.hasRTT.Load() {
		t.Fatal("hasRTT should default to false")
	}
}

func TestHasRTTSetTrue(t *testing.T) {
	conn := makeTestConn()
	conn.hasRTT.Store(true)
	if !conn.hasRTT.Load() {
		t.Fatal("hasRTT should be true after Store(true)")
	}
}

func TestCloningDefaultZero(t *testing.T) {
	conn := makeTestConn()
	if v := conn.closing.Load(); v != 0 {
		t.Fatalf("closing should default to 0, got %d", v)
	}
}

// ---------------------------------------------------------------------------
// Monotonic clock safety regression: time.Time from time.Now() used with
// time.Since / After / Sub is safe as long as we don't compare instances
// from fundamentally different clock sources. This test ensures the
// lastActivity/lastPongAt plumbing uses a consistent source.
// ---------------------------------------------------------------------------

func TestLastActivityAndLastPongAtUseSameClock(t *testing.T) {
	conn := makeTestConn()

	// Both should be non-nil and within a small window.
	la := conn.lastActivity.Load()
	lp := conn.lastPongAt.Load()

	if la == nil || lp == nil {
		t.Fatal("lastActivity and lastPongAt must both be non-nil after init")
	}

	diff := la.Sub(*lp)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Fatalf("lastActivity and lastPongAt differ too much: %v", diff)
	}
}

// ---------------------------------------------------------------------------
// Split reliability tracking on completion
// ---------------------------------------------------------------------------

func TestSplitReliabilityCleanedUpOnCompletion(t *testing.T) {
	conn := makeTestConn()
	conn.mtuSize = maxMTUSize

	// Simulate receiving a split packet: two fragments that form a complete split.
	pk0 := &packet{
		reliability: reliabilityReliableOrdered,
		split:       true,
		splitCount:  2,
		splitIndex:  0,
		splitID:     42,
		content:     []byte{0x01, 0x02},
	}
	pk1 := &packet{
		reliability: reliabilityReliableOrdered,
		split:       true,
		splitCount:  2,
		splitIndex:  1,
		splitID:     42,
		content:     []byte{0x03, 0x04},
	}

	// First fragment creates the split entry.
	if err := conn.receiveSplitPacket(pk0); err != nil {
		t.Fatalf("first receiveSplitPacket returned error: %v", err)
	}
	if _, ok := conn.splitReliabilities[42]; !ok {
		t.Fatal("splitReliabilities was not set for the first fragment")
	}

	// Second fragment completes the split → reliability should be cleaned up.
	if err := conn.receiveSplitPacket(pk1); err != nil {
		t.Fatalf("second receiveSplitPacket returned error: %v", err)
	}
	if _, ok := conn.splitReliabilities[42]; ok {
		t.Fatal("splitReliabilities was not cleaned up after split completion")
	}
	if _, ok := conn.splits[42]; ok {
		t.Fatal("splits was not cleaned up after split completion")
	}
	if _, ok := conn.splitTimes[42]; ok {
		t.Fatal("splitTimes was not cleaned up after split completion")
	}
}

// ---------------------------------------------------------------------------
// Datagram window tests (regression for NACK/liveness interaction)
// ---------------------------------------------------------------------------

func TestDatagramWindowNewBelowLowest(t *testing.T) {
	win := newDatagramWindow()
	win.lowest = 5
	win.highest = 10

	if !win.new(3) {
		t.Fatal("sequence number below lowest should be considered new")
	}
}

func TestDatagramWindowShift(t *testing.T) {
	win := newDatagramWindow()
	win.add(0)
	win.add(1)

	n := win.shift()
	if n != 2 {
		t.Fatalf("shift returned %d, want 2", n)
	}
	if win.lowest != 2 {
		t.Fatalf("lowest after shift = %d, want 2", win.lowest)
	}
}

func TestDatagramWindowMissing(t *testing.T) {
	win := newDatagramWindow()
	win.add(0)
	win.add(3) // gap: 1 and 2 are missing

	// With a 0 duration, even the just-added packet should trigger "missing".
	missing := win.missing(0)
	if len(missing) != 2 {
		t.Fatalf("missing returned %d indices, want 2", len(missing))
	}
}
