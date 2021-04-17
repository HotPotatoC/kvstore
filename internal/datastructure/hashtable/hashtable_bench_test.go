package hashtable_test

import (
	"fmt"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/datastructure/hashtable"
)

func BenchmarkSet(b *testing.B) {
	b.StopTimer()
	ht := hashtable.New()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ht.Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}
}

func BenchmarkDel(b *testing.B) {
	b.StopTimer()
	ht := populate(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ht.Remove(fmt.Sprintf("k%d", i))
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	ht := populate(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ht.Get(fmt.Sprintf("k%d", i))
	}
}

func BenchmarkIter(b *testing.B) {
	b.StopTimer()
	ht := populate(b.N)
	b.StartTimer()
	for bucket := range ht.Iter() {
		if !ht.Exist(bucket.Key) {
			b.Errorf("Iter benchmark failed, key does not exists")
		}
	}
}
