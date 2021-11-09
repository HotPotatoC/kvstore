package protocol_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func BenchmarkReader_SimpleString(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString("+OK\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.NewReader(&buf).ReadObject()
	}
}

func BenchmarkReader_Error(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString("-ERR\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.NewReader(&buf).ReadObject()
	}
}

func BenchmarkReader_Integer(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(":123456789\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.NewReader(&buf).ReadObject()
	}
}

func BenchmarkReader_IntegerNegative(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(":-123456789\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.NewReader(&buf).ReadObject()
	}
}

func BenchmarkReader_BulkString(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString("$5\r\nhello\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.NewReader(&buf).ReadObject()
	}
}

func BenchmarkReader_Array(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString("*2\r\n+OK\r\n$5\r\nhello\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protocol.NewReader(&buf).ReadObject()
	}
}
