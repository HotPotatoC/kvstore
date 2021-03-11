package packet

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/HotPotatoC/kvstore/internal/command"
)

// Packet represents the tcp payload with the command operation and it's arguments
type Packet struct {
	Cmd  command.Op
	Args []byte
}

// NewPacket creates a new packet
func NewPacket(cmd command.Op, args []byte) *Packet {
	return &Packet{
		Cmd:  cmd,
		Args: args,
	}
}

// Encode encodes the packet
func (p *Packet) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	err := gob.NewEncoder(buf).Encode(p)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Decode decodes the packet
func (p *Packet) Decode(buffer *bytes.Buffer) error {
	err := gob.NewDecoder(buffer).Decode(p)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
