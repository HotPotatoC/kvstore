package packet

import (
	"bytes"
	"io"

	"github.com/HotPotatoC/kvstore/command"
	"github.com/HotPotatoC/kvstore/pkg/tlv"
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

	tlvWriter := tlv.NewWriter(buf, tlv.DefaultTLVCodec)
	record := tlv.NewRecord(p.Args, uint(p.Cmd))

	err := tlvWriter.Write(record)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Decode decodes the packet
func (p *Packet) Decode(buffer *bytes.Buffer) error {
	record, err := tlv.NewReader(buffer, tlv.DefaultTLVCodec).Read()
	if err != nil && err != io.EOF {
		return err
	}

	p.Args = record.Payload
	p.Cmd = command.Op(record.Type)

	return nil
}
