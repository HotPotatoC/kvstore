package datastructure

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStoreBytesOwnsInputAndPreservesReaders(t *testing.T) {
	var m Map
	wantKey := strings.Repeat("key", 40)
	key, value := []byte(wantKey), bytes.Repeat([]byte{0, 255, '\r', '\n'}, 32)
	wantValue := string(value)
	if !m.StoreBytes(key, value, time.Minute, false, false) {
		t.Fatal("insert failed")
	}
	old, _ := m.Get(wantKey)
	if !m.StoreBytes(key, []byte("replacement"), 0, false, true) {
		t.Fatal("overwrite failed")
	}
	clear(key)
	clear(value)
	got, ok := m.Get(wantKey)
	if !ok || got.Data != "replacement" || got.Key != wantKey || !got.HasFlag(ItemFlagExpireNX) {
		t.Fatalf("stored item changed: %+v", got)
	}
	if old.Data != wantValue || old.Key != wantKey || !old.HasFlag(ItemFlagExpireXX) {
		t.Fatal("overwrite or buffer reuse mutated previous reader snapshot")
	}
	if m.Len() != 1 {
		t.Fatalf("len=%d", m.Len())
	}
}

func TestStoreBytesConditions(t *testing.T) {
	var m Map
	key := []byte("key")
	if m.StoreBytes(key, []byte("missing"), 0, false, true) {
		t.Fatal("XX inserted missing key")
	}
	if !m.StoreBytes(key, []byte("first"), 0, true, false) || m.StoreBytes(key, []byte("second"), 0, true, false) {
		t.Fatal("NX condition violated")
	}
	expired := NewItem("key", "expired", time.Second)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	m.Store(expired)
	if m.StoreBytes(key, []byte("xx"), 0, false, true) || m.Len() != 0 {
		t.Fatal("XX accepted expired key or failed expiry cleanup")
	}
	if !m.StoreBytes(key, []byte("nx"), 0, true, false) || m.Len() != 1 {
		t.Fatal("NX failed after expiry")
	}
}

func TestStoreBytesConcurrentNX(t *testing.T) {
	var m Map
	var successes atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if m.StoreBytes([]byte("key"), []byte("value"), 0, true, false) {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if successes.Load() != 1 || m.Len() != 1 {
		t.Fatalf("winners=%d len=%d", successes.Load(), m.Len())
	}
}
