package protocol

import (
	"bytes"
	"math"
	"testing"
)

func TestWriteResponses(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write func(*bytes.Buffer)
		want  string
	}{
		{"simple", func(b *bytes.Buffer) { WriteSimpleString(b, "OK") }, "+OK\r\n"},
		{"error", func(b *bytes.Buffer) { WriteError(b, "ERR failed") }, "-ERR failed\r\n"},
		{"zero", func(b *bytes.Buffer) { WriteInteger(b, 0) }, ":0\r\n"},
		{"min int", func(b *bytes.Buffer) { WriteInteger(b, math.MinInt64) }, ":-9223372036854775808\r\n"},
		{"max int", func(b *bytes.Buffer) { WriteInteger(b, math.MaxInt64) }, ":9223372036854775807\r\n"},
		{"empty bulk", func(b *bytes.Buffer) { WriteBulkString(b, "") }, "$0\r\n\r\n"},
		{"binary bulk", func(b *bytes.Buffer) { WriteBulkString(b, "\x00\r\né") }, "$5\r\n\x00\r\né\r\n"},
		{"null", WriteNull, "$-1\r\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			b.WriteString("prefix")
			tc.write(&b)
			if b.String() != "prefix"+tc.want {
				t.Fatalf("response = %q, want %q", b.String(), "prefix"+tc.want)
			}
		})
	}
}
