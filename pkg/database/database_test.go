package database_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/database"
	"github.com/HotPotatoC/kvstore/pkg/datastructure/hashtable"
)

func populate(n int) database.Store {
	db := database.New()
	for i := 0; i < n; i++ {
		db.Set(fmt.Sprintf("k%d", i+1), fmt.Sprintf("v%d", i+1))
	}
	return db
}

func TestPopulate_100(t *testing.T) {
	db := populate(100)
	if db.Size() != 100 {
		t.Errorf("Failed TestPopulate100 -> Expected Size: %d | Got: %d", 100, db.Size())
	}
}

func TestPopulate_1000(t *testing.T) {
	db := populate(1000)
	if db.Size() != 1000 {
		t.Errorf("Failed TestPopulate100 -> Expected Size: %d | Got: %d", 1000, db.Size())
	}
}

func TestPopulate_10000(t *testing.T) {
	db := populate(10000)
	if db.Size() != 10000 {
		t.Errorf("Failed TestPopulate10000 -> Expected Size: %d | Got: %d", 10000, db.Size())
	}
}

func TestPopulate_100000(t *testing.T) {
	db := populate(100000)
	if db.Size() != 100000 {
		t.Errorf("Failed TestPopulate100000 -> Expected Size: %d | Got: %d", 100000, db.Size())
	}
}

func TestSet(t *testing.T) {
	db := populate(4)
	if db.Size() != 4 {
		t.Errorf("Failed TestSet -> Expected Size: %d | Got: %d", 4, db.Size())
	}

	db.Set("my-key", "value")
	if db.Size() != 5 {
		t.Errorf("Failed TestSet -> Expected Size: %d | Got: %d", 5, db.Size())
	}
}

func TestSetEX(t *testing.T) {
	db := populate(4)
	if db.Size() != 4 {
		t.Errorf("Failed TestSetEX -> Expected Size: %d | Got: %d", 4, db.Size())
	}

	db.SetEX("my-key", "value", 5)
	if db.Size() != 5 {
		t.Errorf("Failed TestSetEX -> Expected Size: %d | Got: %d", 5, db.Size())
	}
	time.Sleep(2 * time.Second)
	if !db.Exist("my-key") {
		t.Error("Failed TestSetEX -> Expected Key to exists | Got empty")
	}

	time.Sleep(4 * time.Second)
	if db.Exist("my-key") {
		t.Error("Failed TestSetEX -> Expected Key to be expired | Got a key")
	}
}

func TestRemove(t *testing.T) {
	db := populate(4)

	db.Remove("k1")

	if db.Size() != 3 {
		t.Errorf("Failed TestRemove -> Expected Size: %d | Got: %d", 3, db.Size())
	}
}

func TestGet(t *testing.T) {
	db := populate(5)

	value := db.Get("k2")
	expected := "v2"
	if value != expected {
		t.Errorf("Failed TestGet -> Expected value: %s | Got: %s", expected, value)
	}
}

func TestFlush_100(t *testing.T) {
	db := populate(100)
	if db.Size() != 100 {
		t.Errorf("Failed TestFlush_100 -> Expected Size: %d | Got: %d", 100, db.Size())
	}

	db.Flush()
	if db.Size() != 0 {
		t.Errorf("Failed TestFlush_100 -> Expected Size: %d | Got: %d", 0, db.Size())
	}
}

func TestFlush_1000(t *testing.T) {
	db := populate(1000)
	if db.Size() != 1000 {
		t.Errorf("Failed TestFlush_1000 -> Expected Size: %d | Got: %d", 1000, db.Size())
	}

	db.Flush()
	if db.Size() != 0 {
		t.Errorf("Failed TestFlush_1000 -> Expected Size: %d | Got: %d", 0, db.Size())
	}
}

func TestFlush_10000(t *testing.T) {
	db := populate(10000)
	if db.Size() != 10000 {
		t.Errorf("Failed TestFlush_10000 -> Expected Size: %d | Got: %d", 10000, db.Size())
	}

	db.Flush()
	if db.Size() != 0 {
		t.Errorf("Failed TestFlush_10000 -> Expected Size: %d | Got: %d", 0, db.Size())
	}
}

func TestFlush_100000(t *testing.T) {
	db := populate(100000)
	if db.Size() != 100000 {
		t.Errorf("Failed TestFlush_100000 -> Expected Size: %d | Got: %d", 100000, db.Size())
	}

	db.Flush()
	if db.Size() != 0 {
		t.Errorf("Failed TestFlush_100000 -> Expected Size: %d | Got: %d", 0, db.Size())
	}
}

func TestIter(t *testing.T) {
	db := populate(5)

	kv := make([]*hashtable.Entry, 0)
	for entry := range db.Iter() {
		kv = append(kv, entry)
	}

	if len(kv) != 5 {
		t.Errorf("Failed TestIter -> Expected size: %d | Got: %d", 5, len(kv))
	}
}
