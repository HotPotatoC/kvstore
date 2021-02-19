package hashtable_test

import (
	"fmt"
	"testing"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

func populate(n int) *hashtable.HashTable {
	ht := hashtable.NewHashTable()
	for i := 0; i < n; i++ {
		ht.Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}
	return ht
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

func BenchmarkSet(b *testing.B) {
	ht := hashtable.NewHashTable()
	for i := 0; i < b.N; i++ {
		ht.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}
