package raknet

// Reliability describes the RakNet reliability mode of a user frame.
type Reliability byte

const (
	ReliabilityUnreliable Reliability = Reliability(reliabilityUnreliable)
	ReliabilityReliable   Reliability = Reliability(reliabilityReliable)

	// ReliabilityReliableOrdered is the default mode Raknetify uses for game
	// packets unless it deliberately marks a frame as unordered/unreliable.
	ReliabilityReliableOrdered Reliability = Reliability(reliabilityReliableOrdered)
)

// Frame is a RakNet user frame after the offline/session RakNet protocol has
// been handled by Conn.
type Frame struct {
	Payload      []byte
	Reliability  Reliability
	OrderChannel byte
}
