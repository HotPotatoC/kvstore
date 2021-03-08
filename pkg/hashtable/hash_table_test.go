package hashtable_test

import (
	"fmt"
	"testing"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

func populate(n int) *hashtable.HashTable {
	ht := hashtable.New()
	for i := 0; i < n; i++ {
		ht.Set(fmt.Sprintf("k%d", i+1), fmt.Sprintf("v%d", i+1))
	}
	return ht
}

func TestPopulate_100(t *testing.T) {
	ht := populate(100)
	if ht.Size() != 100 {
		t.Errorf("Failed TestPopulate100 -> Expected Size: %d | Got: %d", 100, ht.Size())
	}
}

func TestPopulate_1000(t *testing.T) {
	ht := populate(1000)
	if ht.Size() != 1000 {
		t.Errorf("Failed TestPopulate100 -> Expected Size: %d | Got: %d", 1000, ht.Size())
	}
}

func TestPopulate_10000(t *testing.T) {
	ht := populate(10000)
	if ht.Size() != 10000 {
		t.Errorf("Failed TestPopulate10000 -> Expected Size: %d | Got: %d", 10000, ht.Size())
	}
}

func TestPopulate_100000(t *testing.T) {
	ht := populate(100000)
	if ht.Size() != 100000 {
		t.Errorf("Failed TestPopulate100000 -> Expected Size: %d | Got: %d", 100000, ht.Size())
	}
}

func TestPut(t *testing.T) {
	ht := populate(4)
	if ht.Size() != 4 {
		t.Errorf("Failed TestPut -> Expected Size: %d | Got: %d", 4, ht.Size())
	}

	ht.Set("my-key", "value")
	if ht.Size() != 5 {
		t.Errorf("Failed TestPut -> Expected Size: %d | Got: %d", 5, ht.Size())
	}
}

func TestRemove(t *testing.T) {
	ht := populate(4)

	ht.Remove("k1")

	if ht.Size() != 3 {
		t.Errorf("Failed TestRemove -> Expected Size: %d | Got: %d", 3, ht.Size())
	}
}

func TestGet(t *testing.T) {
	ht := populate(5)

	value := ht.Get("k2")
	expected := "v2"
	if value != expected {
		t.Errorf("Failed TestGet -> Expected value: %s | Got: %s", expected, value)
	}
}

func TestIter(t *testing.T) {
	ht := populate(5)

	for bucket := range ht.Iter() {
		t.Log(bucket.Key)
		t.Log(bucket.Value)
		if bucket.Next != nil {
			bucket = bucket.Next
			t.Log(bucket.Key)
			t.Log(bucket.Value)
		}
	}
}

func BenchmarkSet(b *testing.B) {
	ht := hashtable.New()
	for i := 0; i < b.N; i++ {
		ht.Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}
}

func BenchmarkDel(b *testing.B) {
	ht := populate(b.N)
	for i := 0; i < b.N; i++ {
		ht.Remove(fmt.Sprintf("k%d", i))
	}
}

func BenchmarkGet(b *testing.B) {
	ht := populate(b.N)
	for i := 0; i < b.N; i++ {
		ht.Get(fmt.Sprintf("k%d", i))
	}
}

func BenchmarkIter(b *testing.B) {
	ht := populate(b.N)
	for bucket := range ht.Iter() {
		if !ht.Exist(bucket.Key) {
			b.Errorf("Iter benchmark failed, key does not exists")
		}
	}
}
