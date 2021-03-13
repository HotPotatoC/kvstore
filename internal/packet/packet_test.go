package packet_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/packet"
)

func TestPackageEncode(t *testing.T) {
	packet := packet.NewPacket(command.GET, []byte("key"))

	buffer, err := packet.Encode()
	if err != nil {
		t.Errorf("Failed TestPackageEncode -> Expected: nil | Got: %v ", err)
	}

	expected := []byte{37, 255, 129, 3, 1, 1, 6, 80, 97, 99, 107, 101, 116, 1, 255, 130, 0, 1, 2, 1, 3, 67, 109, 100, 1, 4, 0, 1, 4, 65, 114, 103, 115, 1, 10, 0, 0, 0, 10, 255, 130, 1, 2, 1, 3, 107, 101, 121, 0}
	if !bytes.Equal(buffer.Bytes(), expected) {
		t.Errorf("Failed TestPackageEncode -> Expected: %s | Got: %s ", expected, buffer.Bytes())
	}
}

func TestPackageDecode(t *testing.T) {
	buf := []byte{37, 255, 129, 3, 1, 1, 6, 80, 97, 99, 107, 101, 116, 1, 255, 130, 0, 1, 2, 1, 3, 67, 109, 100, 1, 4, 0, 1, 4, 65, 114, 103, 115, 1, 10, 0, 0, 0, 10, 255, 130, 1, 2, 1, 3, 107, 101, 121, 0}

	var packet packet.Packet

	err := packet.Decode(bytes.NewBuffer(buf))
	if err != nil {
		t.Errorf("Failed TestPackageDecode -> Expected: nil | Got: %v ", err)
	}

	if packet.Cmd != command.GET {
		t.Errorf("Failed TestPackageDecode -> Expected: %s | Got: %s ", command.GET, packet.Cmd)
	}

	if !bytes.Equal(packet.Args, []byte("key")) {
		t.Errorf("Failed TestPackageDecode -> Expected: %s | Got: %s ", []byte("key"), packet.Args)
	}
}
