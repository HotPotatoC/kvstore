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

func TestSet(t *testing.T) {
	ht := populate(4)
	if ht.Size() != 4 {
		t.Errorf("Failed TestSet -> Expected Size: %d | Got: %d", 4, ht.Size())
	}

	ht.Set("my-key", "value")
	if ht.Size() != 5 {
		t.Errorf("Failed TestSet -> Expected Size: %d | Got: %d", 5, ht.Size())
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

func TestFlush_100(t *testing.T) {
	ht := populate(100)
	if ht.Size() != 100 {
		t.Errorf("Failed TestFlush_100 -> Expected Size: %d | Got: %d", 100, ht.Size())
	}

	ht.Flush()
	if ht.Size() != 0 {
		t.Errorf("Failed TestFlush_100 -> Expected Size: %d | Got: %d", 0, ht.Size())
	}
}

func TestFlush_1000(t *testing.T) {
	ht := populate(1000)
	if ht.Size() != 1000 {
		t.Errorf("Failed TestFlush_1000 -> Expected Size: %d | Got: %d", 1000, ht.Size())
	}

	ht.Flush()
	if ht.Size() != 0 {
		t.Errorf("Failed TestFlush_1000 -> Expected Size: %d | Got: %d", 0, ht.Size())
	}
}

func TestFlush_10000(t *testing.T) {
	ht := populate(10000)
	if ht.Size() != 10000 {
		t.Errorf("Failed TestFlush_10000 -> Expected Size: %d | Got: %d", 10000, ht.Size())
	}

	ht.Flush()
	if ht.Size() != 0 {
		t.Errorf("Failed TestFlush_10000 -> Expected Size: %d | Got: %d", 0, ht.Size())
	}
}

func TestFlush_100000(t *testing.T) {
	ht := populate(100000)
	if ht.Size() != 100000 {
		t.Errorf("Failed TestFlush_100000 -> Expected Size: %d | Got: %d", 100000, ht.Size())
	}

	ht.Flush()
	if ht.Size() != 0 {
		t.Errorf("Failed TestFlush_100000 -> Expected Size: %d | Got: %d", 0, ht.Size())
	}
}

func TestIter(t *testing.T) {
	ht := populate(5)

	kv := make([]*hashtable.Entry, 0)
	for entry := range ht.Iter() {
		kv = append(kv, entry)
	}

	if len(kv) != 5 {
		t.Errorf("Failed TestIter -> Expected size: %d | Got: %d", 5, len(kv))
	}
}
