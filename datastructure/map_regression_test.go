package datastructure

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMapOverwriteKeepsSize(t *testing.T) {
	var m Map
	for i := 0; i < 100; i++ {
		m.Store(NewItem("key", fmt.Sprint(i), 0))
	}
	if m.Len() != 1 {
		t.Fatalf("Len = %d, want 1", m.Len())
	}
	if item, ok := m.Get("key"); !ok || item.Data != "99" {
		t.Fatalf("Get = %v, %v, want latest value", item, ok)
	}
	if deleted := m.Delete("key"); deleted != 1 || m.Len() != 0 {
		t.Fatalf("Delete = %d, Len = %d, want 1 and 0", deleted, m.Len())
	}
}

func TestMapDeleteLiteralBeforePattern(t *testing.T) {
	for _, key := range []string{"a[", "a*", "a?", `a\b`} {
		t.Run(key, func(t *testing.T) {
			var m Map
			m.Store(NewItem(key, "literal", 0))
			m.Store(NewItem("abc", "other", 0))
			if deleted := m.Delete(key); deleted != 1 {
				t.Fatalf("Delete = %d, want 1", deleted)
			}
			if !m.Exists("abc") || m.Exists(key) || m.Len() != 1 {
				t.Fatal("literal deletion changed another key or left its key behind")
			}
		})
	}
}

func TestMapConcurrentExpiredGetAndReplacement(t *testing.T) {
	var m Map
	for i := 0; i < 1000; i++ {
		expired := NewItem("key", "expired", time.Hour)
		expired.ExpiresAt = time.Now().Add(-time.Second)
		m.Store(expired)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for j := 0; j < 4; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				m.Get("key")
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			m.Store(NewItem("key", "replacement", 0))
		}()
		close(start)
		wg.Wait()
		if item, ok := m.Get("key"); !ok || item.Data != "replacement" {
			t.Fatalf("iteration %d: replacement removed", i)
		}
		if m.Len() != 1 {
			t.Fatalf("iteration %d: Len = %d, want 1", i, m.Len())
		}
	}
}

func TestMapExpirePreservesReturnedItems(t *testing.T) {
	var m Map
	m.Store(NewItem("key", "value", 0))
	item, _ := m.Get("key")
	listed := m.List()["key"]
	expiresAt := item.ExpiresAt
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			m.Expire("key", time.Hour)
		}
	}()
	for i := 0; i < 1000; i++ {
		if !item.HasFlag(ItemFlagExpireNX) || !listed.ExpiresAt.Equal(expiresAt) {
			t.Error("Expire changed a previously returned item")
			break
		}
	}
	wg.Wait()
	if !item.HasFlag(ItemFlagExpireNX) || !listed.ExpiresAt.Equal(expiresAt) {
		t.Fatal("Expire changed a previously returned item")
	}
	current, _ := m.Get("key")
	if !current.HasFlag(ItemFlagExpireXX) || current.HasFlag(ItemFlagExpireNX) {
		t.Fatal("stored item did not acquire expiry")
	}
}

func TestMapConcurrentDeleteCounts(t *testing.T) {
	var m Map
	const count = 1000
	for i := 0; i < count; i++ {
		m.Store(NewItem(fmt.Sprint(i), "value", 0))
	}
	var deleted atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if i%2 == 0 {
				deleted.Add(m.Clear())
			} else {
				for j := 0; j < count; j++ {
					deleted.Add(m.Delete(fmt.Sprint(j)))
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if deleted.Load() != count || m.Len() != 0 {
		t.Fatalf("deleted = %d, Len = %d, want %d and 0", deleted.Load(), m.Len(), count)
	}
}

func TestMapTTLTrackingBounded(t *testing.T) {
	var m Map
	key := "key"
	s := &m.shards[shardIndex(key)]
	for i := 0; i < 1000; i++ {
		m.Store(NewItem(key, "value", time.Hour))
		m.Expire(key, time.Hour)
	}
	if s.ttlCount != 1 {
		t.Fatalf("TTL keys = %d, want 1", s.ttlCount)
	}
	m.Store(NewItem(key, "value", 0))
	if s.ttlCount != 0 {
		t.Fatal("permanent replacement retained TTL tracking")
	}
	for _, remove := range []struct {
		name string
		run  func()
	}{
		{"Delete", func() { m.Delete(key) }},
		{"Clear", func() { m.Clear() }},
		{"ExpiredGet", func() { m.Expire(key, -time.Second); m.Get(key) }},
	} {
		t.Run(remove.name, func(t *testing.T) {
			m.Store(NewItem(key, "value", time.Hour))
			remove.run()
			if s.ttlCount != 0 || m.Len() != 0 {
				t.Fatalf("TTL keys = %d, Len = %d, want 0", s.ttlCount, m.Len())
			}
		})
	}
}
