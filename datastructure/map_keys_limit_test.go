package datastructure

import (
	"sort"
	"testing"
	"time"
)

func TestKeysWithPatternLimit(t *testing.T) {
	var m Map
	m.Store(NewItem("key", "v", 0))
	for _, tc := range []struct {
		limit    int
		exceeded bool
	}{{0, false}, {-1, false}, {67, false}, {66, true}, {1, true}} {
		keys, exceeded := m.KeysWithPatternLimit("*", tc.limit)
		if exceeded != tc.exceeded || (!exceeded && (len(keys) != 1 || keys[0] != "key")) || (exceeded && len(keys) != 0) {
			t.Fatalf("limit=%d keys=%v exceeded=%v", tc.limit, keys, exceeded)
		}
	}
	// These writes also prove an early return released the shard read lock.
	m.Store(NewItem("path/key", "v", 0))
	expired := NewItem("expired", "v", time.Second)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	m.Store(expired)
	keys, exceeded := m.KeysWithPatternLimit("*", 107)
	sort.Strings(keys)
	if exceeded || len(keys) != 2 || keys[0] != "key" || keys[1] != "path/key" {
		t.Fatalf("all keys=%v exceeded=%v", keys, exceeded)
	}
	if _, exceeded := m.KeysWithPatternLimit("*", 106); !exceeded {
		t.Fatal("accepted result above limit")
	}
	keys, exceeded = m.KeysWithPatternLimit("path/*", 72)
	if exceeded || len(keys) != 1 || keys[0] != "path/key" {
		t.Fatalf("pattern keys=%v exceeded=%v", keys, exceeded)
	}
	for _, pattern := range []string{"none*", "[", "expired"} {
		keys, exceeded := m.KeysWithPatternLimit(pattern, 1)
		if exceeded || len(keys) != 0 {
			t.Fatalf("pattern=%q keys=%v exceeded=%v", pattern, keys, exceeded)
		}
	}
	// The original pattern API still uses filepath semantics even for '*'.
	keys = m.KeysWithPattern("*")
	if len(keys) != 1 || keys[0] != "key" {
		t.Fatalf("legacy pattern semantics changed: %v", keys)
	}
}
