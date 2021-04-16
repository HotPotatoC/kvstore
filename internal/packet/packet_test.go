package packet_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/packet"
)

func TestPacketEncode(t *testing.T) {
	packet := packet.NewPacket(command.GET, []byte("key"))

	buffer, err := packet.Encode()
	if err != nil {
		t.Errorf("Failed TestPacketEncode -> Expected: nil | Got: %v ", err)
	}

	expected := []byte{0, 2, 0, 0, 0, 3, 107, 101, 121}
	if !bytes.Equal(buffer.Bytes(), expected) {
		t.Errorf("Failed TestPacketEncode -> Expected: %s | Got: %s ", expected, buffer.Bytes())
	}
}

func TestPacketDecode(t *testing.T) {
	buf := []byte{0, 2, 0, 0, 0, 3, 107, 101, 121}

	var packet packet.Packet

	err := packet.Decode(bytes.NewBuffer(buf))
	if err != nil {
		t.Errorf("Failed TestPacketDecode -> Expected: nil | Got: %v ", err)
	}

	if packet.Cmd != command.GET {
		t.Errorf("Failed TestPacketDecode -> Expected: %s | Got: %s ", command.GET, packet.Cmd)
	}

	if !bytes.Equal(packet.Args, []byte("key")) {
		t.Errorf("Failed TestPacketDecode -> Expected: %s | Got: %s ", []byte("key"), packet.Args)
	}
}
