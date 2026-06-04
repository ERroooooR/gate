package raknet

// Reliability describes the RakNet reliability mode of a user frame.
type Reliability byte

const (
	ReliabilityUnreliable          Reliability = Reliability(reliabilityUnreliable)
	ReliabilityUnreliableSequenced Reliability = Reliability(reliabilityUnreliableSequenced)
	ReliabilityReliable            Reliability = Reliability(reliabilityReliable)
	ReliabilityReliableOrdered     Reliability = Reliability(reliabilityReliableOrdered)
	ReliabilityReliableSequenced   Reliability = Reliability(reliabilityReliableSequenced)
)

// IsReliable returns true if the frame is a reliable type.
func (r Reliability) IsReliable() bool {
	return byte(r) == reliabilityReliable ||
		byte(r) == reliabilityReliableOrdered ||
		byte(r) == reliabilityReliableSequenced
}

// Frame is a RakNet user frame after the offline/session RakNet protocol has
// been handled by Conn.
type Frame struct {
	Payload      []byte
	Reliability  Reliability
	OrderChannel byte
}
