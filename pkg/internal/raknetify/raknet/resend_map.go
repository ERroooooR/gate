package raknet

import (
	"math"
	"time"
)

// resendMap is a map of packets, used to recover datagrams if the other end of
// the connection ended up not having them.
type resendMap struct {
	unacknowledged map[uint24]resendRecord
	delays         map[time.Time]time.Duration
	cachedRTT      time.Duration
	cachedRTTAt    time.Time
	cachedStdDev   time.Duration
	cachedMinRTT   time.Duration
	// minRTTResetAt tracks when the minRTT window started, for 10-second sliding window
	minRTTResetAt time.Time
}

// resendRecord represents a single packet with a timestamp from when it was
// initially sent. It may be either acknowledged or NACKed by the other end.
type resendRecord struct {
	pk        *packet
	timestamp time.Time
	// retryCount tracks how many times this record has been retransmitted.
	// Used for exponential backoff in the retry timeout calculation.
	retryCount int
}

// newRecoveryQueue returns a new initialised recovery queue.
func newRecoveryQueue() *resendMap {
	return &resendMap{
		delays:         make(map[time.Time]time.Duration),
		unacknowledged: make(map[uint24]resendRecord),
	}
}

// add puts a packet at the index passed and records the current time.
func (m *resendMap) add(index uint24, pk *packet) {
	m.unacknowledged[index] = resendRecord{pk: pk, timestamp: time.Now()}
}

// addWithRetryCount puts a packet at the index passed with an existing retryCount,
// preserving it across retransmissions for exponential backoff.
func (m *resendMap) addWithRetryCount(index uint24, pk *packet, retryCount int) {
	m.unacknowledged[index] = resendRecord{pk: pk, timestamp: time.Now(), retryCount: retryCount}
}

// acknowledge marks a packet with the index passed as acknowledged. The packet
// is removed from the resendMap and returned if found.
func (m *resendMap) acknowledge(index uint24) (*packet, bool) {
	return m.remove(index, 1)
}

// retransmit looks up a packet with an index from the resendMap so that it may
// be resent. Increments the retry count on the record and returns it alongside
// the packet so the caller can preserve it when re-adding.
func (m *resendMap) retransmit(index uint24) (pk *packet, retryCount int, ok bool) {
	record, found := m.unacknowledged[index]
	if !found {
		return nil, 0, false
	}
	// Increment retry count for exponential backoff BEFORE removing.
	// The caller must use addWithRetryCount to preserve this across re-adds.
	record.retryCount++
	newCount := record.retryCount
	delete(m.unacknowledged, index)

	now := time.Now()
	m.delays[now] = now.Sub(record.timestamp)
	return record.pk, newCount, true
}

// remove deletes an index from the resendMap and adds the time since the
// packet was originally sent multiplied by mul to the delays slice.
func (m *resendMap) remove(index uint24, mul int) (*packet, bool) {
	record, ok := m.unacknowledged[index]
	if !ok {
		return nil, false
	}
	delete(m.unacknowledged, index)

	now := time.Now()
	m.delays[now] = now.Sub(record.timestamp) * time.Duration(mul)
	return record.pk, true
}

// rtt returns the average round trip time between the putting of the value
// into the recovery queue and the taking out of it again. It is measured over
// the last averageDuration worth of delay records.
func (m *resendMap) rtt() time.Duration {
	const averageDuration = time.Second * 5
	const cacheDuration = time.Millisecond * 250
	now := time.Now()
	if m.cachedRTT != 0 && now.Sub(m.cachedRTTAt) < cacheDuration {
		return m.cachedRTT
	}

	var (
		total, records time.Duration
	)
	for t, rtt := range m.delays {
		if now.Sub(t) > averageDuration {
			delete(m.delays, t)
			continue
		}
		total += rtt
		records++
	}
	if records == 0 {
		// No records yet, generally should not happen. Just return a reasonable amount of time.
		m.cachedRTT = time.Millisecond * 50
	} else {
		m.cachedRTT = total / records
	}
	m.cachedRTTAt = now
	return m.cachedRTT
}

// rttStdDev returns the standard deviation of RTT samples over the last 5 seconds.
// Returns 0 if there are fewer than 2 samples.
func (m *resendMap) rttStdDev() time.Duration {
	const averageDuration = time.Second * 5
	const cacheDuration = time.Millisecond * 250
	now := time.Now()
	if m.cachedStdDev != 0 && now.Sub(m.cachedRTTAt) < cacheDuration {
		return m.cachedStdDev
	}

	mean := m.rtt()
	var (
		totalVariance float64
		records       int
	)
	for t, rtt := range m.delays {
		if now.Sub(t) > averageDuration {
			continue
		}
		diff := float64(rtt - mean)
		totalVariance += diff * diff
		records++
	}
	if records < 2 {
		m.cachedStdDev = 0
	} else {
		m.cachedStdDev = time.Duration(math.Sqrt(totalVariance / float64(records)))
	}
	return m.cachedStdDev
}

// minRTT returns the minimum RTT observed over a 10-second sliding window.
// This matches the netty-raknet DefaultConfig.getMinRTTNanos() behavior.
func (m *resendMap) minRTT() time.Duration {
	const minRTTWindow = time.Second * 10
	const cacheDuration = time.Millisecond * 250
	now := time.Now()
	if m.cachedMinRTT != 0 && now.Sub(m.cachedRTTAt) < cacheDuration {
		return m.cachedMinRTT
	}

	// Reset minRTT window every 10 seconds
	if now.Sub(m.minRTTResetAt) > minRTTWindow {
		m.minRTTResetAt = now
		m.cachedMinRTT = 0
	}

	var (
		minRTT  time.Duration
		records int
	)
	for t, rtt := range m.delays {
		if now.Sub(t) > minRTTWindow {
			continue
		}
		if minRTT == 0 || rtt < minRTT {
			minRTT = rtt
		}
		records++
	}
	if records == 0 && m.cachedMinRTT == 0 {
		m.cachedMinRTT = m.rtt() // fallback to average RTT
	} else if minRTT != 0 {
		m.cachedMinRTT = minRTT
	}
	return m.cachedMinRTT
}

// retryCount returns the maximum retry count among all unacknowledged records.
func (m *resendMap) maxRetryCount() int {
	maxCount := 0
	for _, record := range m.unacknowledged {
		if record.retryCount > maxCount {
			maxCount = record.retryCount
		}
	}
	return maxCount
}
