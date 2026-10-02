package datastructure

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStoreIfAbsentConcurrent(t *testing.T) {
	var m Map
	var successes atomic.Int64
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if m.StoreIfAbsent(NewItem("key", "value", 0)) {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if successes.Load() != 1 || m.Len() != 1 {
		t.Fatalf("successes=%d count=%d, want 1 each", successes.Load(), m.Len())
	}
}

func TestConditionalStore(t *testing.T) {
	var m Map
	if m.StoreIfPresent(NewItem("key", "missing", 0)) {
		t.Fatal("XX created absent key")
	}
	if !m.StoreIfAbsent(NewItem("key", "first", 0)) {
		t.Fatal("NX rejected absent key")
	}
	if m.StoreIfAbsent(NewItem("key", "ignored", 0)) {
		t.Fatal("NX replaced existing key")
	}
	if !m.StoreIfPresent(NewItem("key", "second", 0)) {
		t.Fatal("XX rejected existing key")
	}
	if item, ok := m.Get("key"); !ok || item.Data != "second" || m.Len() != 1 {
		t.Fatalf("item=%v found=%v count=%d", item, ok, m.Len())
	}
}

func TestConditionalStoreExpiredKey(t *testing.T) {
	for _, onlyAbsent := range []bool{true, false} {
		var m Map
		expired := NewItem("key", "expired", time.Hour)
		expired.ExpiresAt = time.Now().Add(-time.Second)
		m.Store(expired)
		if onlyAbsent {
			if !m.StoreIfAbsent(NewItem("key", "fresh", 0)) {
				t.Fatal("NX rejected expired key")
			}
			if item, ok := m.Get("key"); !ok || item.Data != "fresh" || m.Len() != 1 {
				t.Fatal("NX did not replace expired key")
			}
		} else {
			if m.StoreIfPresent(NewItem("key", "fresh", 0)) {
				t.Fatal("XX replaced expired key")
			}
			if m.Len() != 0 {
				t.Fatalf("expired item count=%d, want 0", m.Len())
			}
		}
	}
}
