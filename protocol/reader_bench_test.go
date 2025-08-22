package protocol_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func BenchmarkReader_SimpleString(b *testing.B) {
	benchmarks := []struct {
		name    string
		payload []byte
	}{
		{"OK_FastPath", []byte("+OK\r\n")},
		{"Generic", []byte("+Some other simple string\r\n")},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := bytes.NewReader(bm.payload)
			r := protocol.NewReader(data)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ReadObject()
				data.Seek(0, io.SeekStart)
			}
		})
	}
}

func BenchmarkReader_Error(b *testing.B) {
	benchmarks := []struct {
		name    string
		payload []byte
	}{
		{"Short", []byte("-ERR\r\n")},
		{"Long_With_Spaces", []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := bytes.NewReader(bm.payload)
			r := protocol.NewReader(data)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ReadObject()
				data.Seek(0, io.SeekStart)
			}
		})
	}
}

func BenchmarkReader_Integer(b *testing.B) {
	data := bytes.NewReader([]byte(":123456789\r\n"))
	r := protocol.NewReader(data)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.ReadObject()
		data.Seek(0, io.SeekStart)
	}
}

func BenchmarkReader_IntegerNegative(b *testing.B) {
	data := bytes.NewReader([]byte(":-123456789\r\n"))
	r := protocol.NewReader(data)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.ReadObject()
		data.Seek(0, io.SeekStart)
	}
}

func generateBulkString(size int) []byte {
	value := strings.Repeat("a", size)
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", size, value))
}

func BenchmarkReader_BulkString(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_32B", 32},
		{"Medium_4KB", 4 * 1024},
		{"Large_1MB", 1 * 1024 * 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			payload := generateBulkString(bm.size)
			data := bytes.NewReader(payload)
			r := protocol.NewReader(data)
			b.SetBytes(int64(len(payload))) // Reports throughput (MB/s)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ReadObject()
				data.Seek(0, io.SeekStart)
			}
		})
	}
}

func BenchmarkReader_Array(b *testing.B) {
	benchmarks := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "FlatCommand_SET",
			payload: []byte("*3\r\n$3\r\nSET\r\n$10\r\nmy-key-123\r\n$10\r\nmy-value-a\r\n"),
		},
		{
			name:    "MixedTypes",
			payload: []byte("*5\r\n:1\r\n:2\r\n:3\r\n$5\r\nhello\r\n-Error message\r\n"),
		},
		{
			name:    "LargeFlat_100",
			payload: []byte("*100\r\n" + strings.Repeat("$1\r\na\r\n", 100)),
		},
		{
			name:    "Nested",
			payload: []byte("*2\r\n*2\r\n$1\r\na\r\n$1\r\nb\r\n*2\r\n$1\r\nc\r\n$1\r\nd\r\n"),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := bytes.NewReader(bm.payload)
			r := protocol.NewReader(data)
			b.SetBytes(int64(len(bm.payload)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ReadObject()
				data.Seek(0, io.SeekStart)
			}
		})
	}
}
