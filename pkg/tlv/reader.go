package tlv

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Reader decodes records from TLV format messages
type Reader struct {
	reader io.Reader
	codec  *Codec
}

// NewReader creates a new reader for decoding
func NewReader(reader io.Reader, codec *Codec) *Reader {
	return &Reader{
		reader: reader,
		codec:  codec,
	}
}

// Read decodes a record from the provided io.Reader
func (r *Reader) Read() (*Record, error) {
	typeBytes := make([]byte, r.codec.TypeBytes)
	_, err := r.reader.Read(typeBytes)
	if err != nil {
		return nil, err
	}

	recordType := readUint(typeBytes, r.codec.TypeBytes)

	payloadLenBytes := make([]byte, r.codec.LenBytes)
	_, err = r.reader.Read(payloadLenBytes)
	if err != nil && err != io.EOF {
		return nil, err
	}

	recordPayloadLen := readUint(payloadLenBytes, r.codec.LenBytes)

	if err == io.EOF && recordPayloadLen != 0 {
		return nil, err
	}

	recordValue := make([]byte, recordPayloadLen)
	_, err = r.reader.Read(recordValue)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &Record{
		Type:    recordType,
		Payload: recordValue,
	}, nil
}

func readUint(b []byte, n ByteSize) uint {
	reader := bytes.NewReader(b)
	switch n {
	case OneByte:
		var i uint8
		binary.Read(reader, binary.BigEndian, &i)
		return uint(i)
	case TwoBytes:
		var i uint16
		binary.Read(reader, binary.BigEndian, &i)
		return uint(i)
	case FourBytes:
		var i uint32
		binary.Read(reader, binary.BigEndian, &i)
		return uint(i)
	case EightBytes:
		var i uint64
		binary.Read(reader, binary.BigEndian, &i)
		return uint(i)
	default:
		return 0
	}
}
