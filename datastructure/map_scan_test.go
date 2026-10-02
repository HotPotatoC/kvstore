package datastructure

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func scanAll(t *testing.T, m *Map, pattern string, count, output int) map[string]bool {
	t.Helper()
	seen := make(map[string]bool)
	var cursor uint64
	for calls := 0; ; calls++ {
		if calls > 10000 {
			t.Fatal("scan did not terminate")
		}
		next, keys, err := m.Scan(cursor, pattern, count, output)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range keys {
			seen[key] = true
		}
		if next == 0 {
			return seen
		}
		cursor = next
	}
}

func TestScanFullIterationAndFilters(t *testing.T) {
	var m Map
	for i := 0; i < 2000; i++ {
		m.Store(NewItem(fmt.Sprintf("key:%04d", i), "value", 0))
	}
	for _, key := range []string{"", "path/to/key", "binary\x00\xff", "literal*"} {
		m.Store(NewItem(key, "v", 0))
	}
	expired := NewItem("expired", "v", time.Hour)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	m.Store(expired)
	all := scanAll(t, &m, "*", 7, 0)
	if len(all) != 2004 || all["expired"] {
		t.Fatalf("returned %d keys, expired=%t", len(all), all["expired"])
	}
	for _, key := range []string{"", "path/to/key", "binary\x00\xff", "literal*"} {
		if !all[key] {
			t.Fatalf("missing %q", key)
		}
	}
	if got := scanAll(t, &m, "path*", 11, 0); len(got) != 1 || !got["path/to/key"] {
		t.Fatalf("slash match: %v", got)
	}
	if got := scanAll(t, &m, "literal\\*", 10, 0); len(got) != 1 {
		t.Fatalf("escaped match: %v", got)
	}
	if got := scanAll(t, &m, "absent*", 1, 0); len(got) != 0 {
		t.Fatalf("sparse matches: %v", got)
	}
}

func fillOneShard(m *Map, count int) {
	for i, added := 0, 0; added < count; i++ {
		key := fmt.Sprintf("collision:%d", i)
		if shardIndex(key) == 0 {
			m.Store(NewItem(key, "v", 0))
			added++
		}
	}
}

func TestScanCountAndOutputBudgets(t *testing.T) {
	var m Map
	fillOneShard(&m, MaxScanCount+1)
	next, keys, err := m.Scan(0, "missing", 1, 0)
	if err != nil || len(keys) != 0 || next != 256 {
		t.Fatalf("nonmatching work was not bounded: %d %v %v", next, keys, err)
	}
	next, keys, err = m.Scan(0, "*", 1<<30, 0)
	if err != nil || len(keys) > MaxScanCount || next == 0 {
		t.Fatalf("large count: next=%d keys=%d err=%v", next, len(keys), err)
	}
	seen := scanAll(t, &m, "*", 100, 128)
	if len(seen) != MaxScanCount+1 {
		t.Fatalf("output-limited scan lost keys: %d", len(seen))
	}
	if _, _, err := m.Scan(0, "*", 10, 65); err == nil {
		t.Fatal("oversized single key accepted")
	}
	if next, _, err := m.Scan(^uint64(0), "*", 10, 0); err != nil || next != 0 {
		t.Fatalf("large cursor: %d %v", next, err)
	}
}

func TestScanSurvivesGrowthDeletionAndSlotReuse(t *testing.T) {
	var m Map
	for i := 0; i < 1000; i++ {
		m.Store(NewItem(fmt.Sprintf("stable:%d", i), "v", 0))
	}
	seen := make(map[string]bool)
	var cursor uint64
	for page := 0; ; page++ {
		next, keys, err := m.Scan(cursor, "*", 13, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range keys {
			seen[key] = true
		}
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("churn:%d", page*20+i)
			m.Store(NewItem(key, "value", 0))
			m.Delete(key)
		}
		if next == 0 {
			break
		}
		if page > 1000 {
			t.Fatal("scan did not terminate under bounded churn")
		}
		cursor = next
	}
	for i := 0; i < 1000; i++ {
		if !seen[fmt.Sprintf("stable:%d", i)] {
			t.Fatalf("missed continuously present key %d", i)
		}
	}
	var slots int
	for i := range m.shards {
		slots += len(m.shards[i].slots)
	}
	if slots > 1000+numShards {
		t.Fatalf("deleted slots were not reused: %d", slots)
	}
}

func TestScanConcurrentMutations(t *testing.T) {
	var m Map
	for i := 0; i < 100; i++ {
		m.Store(NewItem(fmt.Sprintf("stable:%d", i), "v", 0))
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("changing:%d", i%30)
			m.Store(NewItem(key, "v", 0))
			m.Expire(key, time.Hour)
			m.Delete(key)
		}
	}()
	seen := scanAll(t, &m, "stable:*", 1, 0)
	wg.Wait()
	if len(seen) != 100 {
		t.Fatalf("stable key count=%d", len(seen))
	}
}

func TestScanSlotIndexDoesNotMutateSharedItems(t *testing.T) {
	var a, b Map
	item := NewItem("shared", "value", time.Hour)
	a.Store(item)
	b.Store(item)
	a.Delete("shared")
	if !scanAll(t, &b, "*", 10, 0)["shared"] {
		t.Fatal("sharing an item corrupted another map")
	}
	b.Delete("shared")
}

func TestScanMatchWorkIsBounded(t *testing.T) {
	var m Map
	m.Store(NewItem(strings.Repeat("a", 20000), "v", 0))
	if _, _, err := m.Scan(0, "*"+strings.Repeat("a", 100)+"b", 10, 0); err == nil {
		t.Fatal("pathological glob exceeded work budget")
	}
}

func BenchmarkScanSparse(b *testing.B) {
	for _, n := range []int{10000, 100000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			var m Map
			for i := 0; i < n; i++ {
				m.Store(NewItem(fmt.Sprintf("key:%d", i), "value", 0))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = m.Scan(0, "absent:*", 10, 1<<20)
			}
		})
	}
}
