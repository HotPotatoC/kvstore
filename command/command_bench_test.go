package command_test

import (
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/command"
)

func Benchmark_WrapArgsFromQuotes_HelloWorld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		args := [][]byte{{byte('"')}, []byte("Hello"), {' '}, []byte("World"), {'"'}}
		command.WrapArgsFromQuotes(args)
	}
}

func Benchmark_WrapArgsFromQuotes_JSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		args := [][]byte{{byte('"')}, []byte("{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}"), {'"'}}
		command.WrapArgsFromQuotes(args)
	}
}
