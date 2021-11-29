package common_test

import (
	"strconv"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/common"
)

func Benchmark_ByteToInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = common.ByteToInt([]byte("123456789"))
	}
}

func Benchmark_ByteToIntStrconv(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = strconv.Atoi("123456789")
	}
}
