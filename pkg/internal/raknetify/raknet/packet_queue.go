package raknet

import "time"

// packetQueue is an ordered queue for reliable ordered packets.
type packetQueue struct {
	lowest  uint24
	highest uint24
	queue   map[uint24]*Frame

	// gapStart tracks when the first gap was detected for unreliable frames.
	// Reset to zero when the gap is resolved. Only used for unreliable ordering.
	gapStart time.Time
}

// newPacketQueue returns a new initialised ordered queue.
func newPacketQueue() *packetQueue {
	return &packetQueue{queue: make(map[uint24]*Frame)}
}

// put puts a value at the index passed. If the index was already occupied
// once, false is returned.
func (queue *packetQueue) put(index uint24, packet *Frame) bool {
	if index < queue.lowest {
		return false
	}
	if _, ok := queue.queue[index]; ok {
		return false
	}
	if index >= queue.highest {
		queue.highest = index + 1
	}
	queue.queue[index] = packet
	return true
}

// putUnreliable puts a value at the index passed, tracking gap timeout for
// unreliable frames. Returns (inserted, isFirstInGap).
// inserted=false means the index was already occupied or below lowest.
// isFirstInGap=true means this insert created a new gap.
func (queue *packetQueue) putUnreliable(index uint24, packet *Frame) (inserted bool, isFirstInGap bool) {
	if !queue.put(index, packet) {
		return false, false
	}
	if index > queue.lowest && queue.gapStart.IsZero() {
		queue.gapStart = time.Now()
		return true, true
	}
	return true, false
}

// gapSince returns the duration since the gap was first detected.
func (queue *packetQueue) gapSince() time.Duration {
	if queue.gapStart.IsZero() {
		return 0
	}
	return time.Since(queue.gapStart)
}

// flushGap skips over missing packets and delivers all queued data past the
// gap. Handles consecutive gaps by scanning forward. Returns the frames that
// should be delivered.
func (queue *packetQueue) flushGap() (packets []*Frame) {
	nextIdx := int(queue.lowest) + 1
	for {
		p, ok := queue.queue[uint24(nextIdx)]
		if ok {
			delete(queue.queue, uint24(nextIdx))
			packets = append(packets, p)
			queue.lowest = uint24(nextIdx)
			nextIdx++
			for {
				p2, ok2 := queue.queue[uint24(nextIdx)]
				if !ok2 {
					queue.lowest = uint24(nextIdx)
					break
				}
				delete(queue.queue, uint24(nextIdx))
				packets = append(packets, p2)
				queue.lowest = uint24(nextIdx)
				nextIdx++
			}
			break
		}
		queue.lowest = uint24(nextIdx)
		nextIdx++
		if len(queue.queue) == 0 {
			break
		}
	}
	queue.gapStart = time.Time{}
	return
}

// fetch attempts to take out as many values from the ordered queue as
// possible. Upon encountering an index that has no value yet, the function
// returns all values that it did find and takes them out.
func (queue *packetQueue) fetch() (packets []*Frame) {
	index := queue.lowest
	for index < queue.highest {
		packet, ok := queue.queue[index]
		if !ok {
			break
		}
		delete(queue.queue, index)
		packets = append(packets, packet)
		index++
	}
	queue.lowest = index
	// If the gap was resolved by fetch (all consecutive), clear gap tracking.
	if queue.lowest >= queue.highest {
		queue.gapStart = time.Time{}
	}
	return
}

// WindowSize returns the size of the window held by the packet queue.
func (queue *packetQueue) WindowSize() uint24 {
	return queue.highest - queue.lowest
}
