package protocol_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func BenchmarkWriter_WriteCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		protocol.NewWriter(&buf).WriteCommand("PING")
	}
}

func BenchmarkWriter_WriteCommand_With_Arguments(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		protocol.NewWriter(&buf).WriteCommand("SET", "key", "value")
	}
}

func BenchmarkWriter_WriteSimpleString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		protocol.NewWriter(&buf).WriteSimpleString("PONG")
	}
}

func BenchmarkWriter_WriteError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		protocol.NewWriter(&buf).WriteError("ERR unknown command 'foobar'")
	}
}

func BenchmarkWriter_WriteInteger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		protocol.NewWriter(&buf).WriteInteger(123)
	}
}

func BenchmarkWriter_WriteBulkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		protocol.NewWriter(&buf).WriteBulkString("PONG")
	}
}
