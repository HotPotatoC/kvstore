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
	hmap.Store(datastructure.NewItem("hello", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hallo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hbllo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hxllo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hllo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("heeeello", []byte("value"), 0))

	benchmarks := []struct {
		name    string
		pattern string
	}{
		{"All", "*"},
		{"AllMixed", "h*llo"},
		{"OneChar", "h?llo"},
		{"Range", "h[a-e]llo"},
		{"Mixed", "?[a-e]*"},
	}
	b.ResetTimer()
	for _, v := range benchmarks {
		b.Run(v.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				hmap.Delete(v.pattern)
			}
		})
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

func Benchmark_KeysWithPattern(b *testing.B) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("hello", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hallo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hbllo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hxllo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("hllo", []byte("value"), 0))
	hmap.Store(datastructure.NewItem("heeeello", []byte("value"), 0))

	benchmarks := []struct {
		name    string
		pattern string
	}{
		{"All", "*"},
		{"AllMixed", "h*llo"},
		{"OneChar", "h?llo"},
		{"Range", "h[a-e]llo"},
		{"Mixed", "?[a-e]*"},
	}
	b.ResetTimer()
	for _, v := range benchmarks {
		b.Run(v.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				hmap.KeysWithPattern(v.pattern)
			}
		})
	}
}
