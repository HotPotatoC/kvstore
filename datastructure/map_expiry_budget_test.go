package datastructure

import (
	"fmt"
	"testing"
	"time"
)

func TestExpiryCycleStrictBudgetAndFairness(t *testing.T) {
	var m Map
	fillOneShard(&m, 100)
	for _, item := range m.List() {
		m.Expire(item.Key, time.Hour)
	}
	// Put already-expired items after the permanent TTL entries in each list.
	for i := 0; i < 100; i++ {
		item := NewItem(fmt.Sprintf("expired:%d", i), "v", time.Hour)
		item.ExpiresAt = time.Now().Add(-time.Second)
		m.Store(item)
	}
	before := m.UsedMemory()
	for i := 0; i < 1000 && m.Len() > 100; i++ {
		if checked := m.expireCycle(7, time.Time{}); checked > 7 {
			t.Fatalf("checked %d, budget7", checked)
		}
	}
	if m.Len() != 100 || m.UsedMemory() >= before {
		t.Fatalf("cleanup starved: keys=%d memory=%d", m.Len(), m.UsedMemory())
	}
	if checked := m.expireCycle(1000, time.Now().Add(-time.Second)); checked != 0 {
		t.Fatalf("expired time budget checked %d", checked)
	}
}

func TestTTLIndexTracksReplacementsExpiryAndClear(t *testing.T) {
	var m Map
	fillOneShard(&m, 20)
	keys := m.Keys()
	for cycle := 0; cycle < 20; cycle++ {
		for _, key := range keys {
			m.Expire(key, time.Hour)
			m.Store(NewItem(key, "value", time.Hour))
		}
		if m.shards[0].ttlCount != 20 {
			t.Fatal("TTL replacements duplicated index entries")
		}
		for i, key := range keys {
			if i%2 == 0 {
				m.Store(NewItem(key, "value", 0))
			} else {
				m.Expire(key, 0)
			}
		}
		if m.shards[0].ttlCount != 0 {
			t.Fatal("TTL index retained removed expiries")
		}
		for _, key := range keys {
			m.Store(NewItem(key, "value", 0))
		}
	}
	m.Clear()
	if m.shards[0].ttlCount != 0 || len(m.shards[0].slots) != 0 || m.UsedMemory() != 0 {
		t.Fatal("clear retained index or memory")
	}
}
