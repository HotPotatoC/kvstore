package protocol_test

import (
	"strings"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func BenchmarkWriter_MakeCommand(b *testing.B) {
	benchmarks := []struct {
		name string
		args []string
	}{
		{"Simple_PING", []string{"PING"}},
		{"Simple_SET", []string{"SET", "key", "value"}},
		{"Complex_MSET", []string{"MSET", "key1", "value1", "key2", "value2", "key3", "value3"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				protocol.MakeCommand(bm.args...)
			}
		})
	}
}

func BenchmarkWriter_MakeSimpleString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		protocol.MakeSimpleString("PONG")
	}
}

func BenchmarkWriter_MakeError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		protocol.MakeError("ERR unknown command 'foobar'")
	}
}

func BenchmarkWriter_MakeInteger(b *testing.B) {
	benchmarks := []struct {
		name  string
		value int64
	}{
		{"Small", 123},
		{"Medium", 123456},
		{"Large", 123456789123456},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				protocol.MakeInteger(bm.value)
			}
		})
	}
}

func BenchmarkWriter_MakeBool(b *testing.B) {
	b.Run("True", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			protocol.MakeBool(true)
		}
	})
	b.Run("False", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			protocol.MakeBool(false)
		}
	})
}

func BenchmarkWriter_MakeBulkString(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_32B", 32},
		{"Medium_4KB", 4 * 1024},
		{"Large_256KB", 256 * 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			payload := strings.Repeat("a", bm.size)

			b.ReportAllocs()
			b.SetBytes(int64(len(payload)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				protocol.MakeBulkString(payload)
			}
		})
	}
}

func BenchmarkWriter_MakeNull(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		protocol.MakeNull()
	}
}
