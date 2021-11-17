package protocol_test

import (
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func BenchmarkWriter_MakeCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeCommand("PING")
	}
}

func BenchmarkWriter_MakeCommand_With_Arguments(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeCommand("SET", "key", "value")
	}
}

func BenchmarkWriter_MakeSimpleString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeSimpleString("PONG")
	}
}

func BenchmarkWriter_MakeError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeError("ERR unknown command 'foobar'")
	}
}

func BenchmarkWriter_MakeInteger_123(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeInteger(123)
	}
}

func BenchmarkWriter_MakeInteger_123456(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeInteger(123456)
	}
}

func BenchmarkWriter_MakeInteger_123456789(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeInteger(123456789)
	}
}

func BenchmarkWriter_MakeBool_True(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeBool(true)
	}
}

func BenchmarkWriter_MakeBool_False(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeBool(false)
	}
}

func BenchmarkWriter_MakeBulkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protocol.MakeBulkString("PONG")
	}
}
