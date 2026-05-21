package raknet

import "testing"

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
