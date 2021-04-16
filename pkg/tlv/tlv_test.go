package tlv_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/pkg/tlv"
)

func TestReadWriteTLV(t *testing.T) {
	var buf bytes.Buffer
	w := tlv.NewWriter(&buf, tlv.DefaultTLVCodec)

	err := w.Write(tlv.NewRecord([]byte("hello, world!"), 0x8))
	if err != nil {
		t.Error(err)
	}

	br := bytes.NewReader(buf.Bytes())
	r := tlv.NewReader(br, tlv.DefaultTLVCodec)
	next, err := r.Read()
	if err != nil {
		t.Error(err)
	}

	if next.Type != 0x8 {
		t.Errorf("Failed TesReadWriteTLV -> Expected: %d | Got: %d", 0x8, next.Type)
	}

	if !bytes.Equal(next.Payload, []byte("hello, world!")) {
		t.Errorf("Failed TesReadWriteTLV -> Expected: %d | Got: %d", []byte("hello, world!"), next.Payload)
	}
}
