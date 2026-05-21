package raknet

import (
	"encoding/binary"
	"testing"
)

func TestReceivePacketOrdersByChannel(t *testing.T) {
	conn := &Conn{packets: make(chan *Frame, 4)}

	packets := []*packet{
		{reliability: reliabilityReliableOrdered, orderChannel: 0, orderIndex: 0, content: []byte{0xfd, 0x00}},
		{reliability: reliabilityReliableOrdered, orderChannel: 1, orderIndex: 0, content: []byte{0xfd, 0x01}},
		{reliability: reliabilityReliableOrdered, orderChannel: 0, orderIndex: 1, content: []byte{0xfd, 0x02}},
	}

	for _, pk := range packets {
		if err := conn.receivePacket(pk); err != nil {
			t.Fatal(err)
		}
	}

	for i, want := range [][]byte{{0xfd, 0x00}, {0xfd, 0x01}, {0xfd, 0x02}} {
		select {
		case frame := <-conn.packets:
			if string(frame.Payload) != string(want) {
				t.Fatalf("frame %d: got %v, want %v", i, frame.Payload, want)
			}
		default:
			t.Fatalf("frame %d was not delivered", i)
		}
	}
}

func TestRaknetifySyncFrameUsesNextOutgoingSequenceID(t *testing.T) {
	conn := &Conn{seq: 42}

	frame := conn.RaknetifySyncFrame()
	payload := frame.Payload
	gotSeq := binary.BigEndian.Uint32(payload[len(payload)-4:])
	if gotSeq != 42 {
		t.Fatalf("sync sequence id = %d, want 42", gotSeq)
	}
	if frame.Reliability != ReliabilityReliable {
		t.Fatalf("sync reliability = %d, want %d", frame.Reliability, ReliabilityReliable)
	}
}
