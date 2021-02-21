package packet

import (
	"bytes"
	"encoding/gob"
	"io"

	"github.com/HotPotatoC/kvstore/internal/command"
)

type Packet struct {
	Cmd  command.CommandOp
	Args []byte
}

func NewPacket(cmd command.CommandOp, args []byte) *Packet {
	return &Packet{
		Cmd:  cmd,
		Args: args,
	}
}

func (p *Packet) Encode() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	err := gob.NewEncoder(buf).Encode(p)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (p *Packet) Decode(buffer *bytes.Buffer) error {
	err := gob.NewDecoder(buffer).Decode(p)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
