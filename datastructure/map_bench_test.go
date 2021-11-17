package datastructure_test

import (
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func Benchmark_Set(b *testing.B) {
	hmap := datastructure.NewMap()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmap.Store(datastructure.NewItem("key", []byte("value"), 0))
	}
}

func Benchmark_Get(b *testing.B) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmap.Get("key")
	}
}

func Benchmark_Delete(b *testing.B) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmap.Delete("key")
	}
}

func Benchmark_List(b *testing.B) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmap.List()
	}
}

func Benchmark_Keys(b *testing.B) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmap.Keys()
	}
}

func Benchmark_Values(b *testing.B) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmap.Values()
	}
}
