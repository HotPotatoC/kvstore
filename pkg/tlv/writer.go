package tlv

import (
	"encoding/binary"
	"io"
)

// Writer is for encoding messages
// containts the io writer interface and the configured codec
type Writer struct {
	writer io.Writer
	codec  *Codec
}

// NewWriter creates a new TLV message encoding scheme writer
func NewWriter(writer io.Writer, codec *Codec) *Writer {
	return &Writer{
		writer: writer,
		codec:  codec,
	}
}

// Write encodes records to TLV format into the provided io.Writer
func (w *Writer) Write(record *Record) error {
	if err := writeUint(w.writer, w.codec.TypeBytes, record.Type); err != nil {
		return err
	}

	ulen := uint(len(record.Payload))
	if err := writeUint(w.writer, w.codec.LenBytes, ulen); err != nil {
		return err
	}

	_, err := w.writer.Write(record.Payload)
	return err
}

func writeUint(w io.Writer, n ByteSize, i uint) error {
	var num interface{}
	switch n {
	case OneByte:
		num = uint8(i)
	case TwoBytes:
		num = uint16(i)
	case FourBytes:
		num = uint32(i)
	case EightBytes:
		num = uint64(i)
	}

	return binary.Write(w, binary.BigEndian, num)
}
