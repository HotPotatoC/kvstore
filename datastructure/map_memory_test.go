package datastructure

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryBudgetRejectsGrowthWithoutMutation(t *testing.T) {
	var m Map
	limit := itemBytes("key", "value")
	m.SetMaxMemory(limit)
	if ok, err := m.StoreBytesLimited([]byte("key"), []byte("value"), time.Hour, false, false); !ok || err != nil {
		t.Fatalf("initial write: %v %v", ok, err)
	}
	before, _ := m.Get("key")
	if ok, err := m.StoreBytesLimited([]byte("key"), []byte("larger"), 0, false, false); ok || !errors.Is(err, ErrMaxMemory) {
		t.Fatalf("growth: %v %v", ok, err)
	}
	if after, _ := m.Get("key"); after != before || m.UsedMemory() != limit {
		t.Fatal("rejected write changed data, TTL, or accounting")
	}
	if ok, err := m.StoreBytesLimited([]byte("key"), []byte("larger"), 0, true, false); ok || err != nil {
		t.Fatalf("failed NX should not reserve memory: %v %v", ok, err)
	}
	if ok, err := m.StoreLimited(NewItem("other", "x", 0)); ok || !errors.Is(err, ErrMaxMemory) {
		t.Fatalf("new key admitted over limit: %v %v", ok, err)
	}
	if ok, err := m.StoreBytesLimited([]byte("key"), []byte("v"), 0, false, true); !ok || err != nil {
		t.Fatalf("shrinking write: %v %v", ok, err)
	}
	if m.UsedMemory() != itemBytes("key", "v") {
		t.Fatal("shrink did not release memory")
	}
	if m.Delete("key") != 1 || m.UsedMemory() != 0 {
		t.Fatal("delete did not release memory")
	}
}

func TestMemoryBudgetConcurrentAdmission(t *testing.T) {
	var m Map
	const capacity = 20
	m.SetMaxMemory(capacity * itemBytes("k000", "value"))
	var accepted atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ok, err := m.StoreBytesLimited([]byte(fmt.Sprintf("k%03d", i)), []byte("value"), 0, false, false)
			if ok {
				accepted.Add(1)
			} else if !errors.Is(err, ErrMaxMemory) {
				t.Errorf("write: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if accepted.Load() != capacity || m.Len() != capacity || m.UsedMemory() != m.maxMemory.Load() {
		t.Fatalf("accepted=%d keys=%d memory=%d", accepted.Load(), m.Len(), m.UsedMemory())
	}
	m.Clear()
	if m.UsedMemory() != 0 {
		t.Fatal("clear did not release memory")
	}
}

func TestExpiredKeysReleaseMemoryAndCannotBeRevived(t *testing.T) {
	for _, op := range []string{"get", "exists", "expire", "delete", "replace"} {
		t.Run(op, func(t *testing.T) {
			var m Map
			i := NewItem("key", "value", time.Hour)
			i.ExpiresAt = time.Now().Add(-time.Second)
			m.Store(i)
			m.SetMaxMemory(itemBytes("key", "value"))
			switch op {
			case "get":
				if _, ok := m.Get("key"); ok {
					t.Fatal("expired key returned")
				}
			case "exists":
				if m.Exists("key") {
					t.Fatal("expired key exists")
				}
			case "expire":
				if m.Expire("key", time.Hour) != 0 {
					t.Fatal("expired key revived")
				}
			case "delete":
				if m.Delete("key") != 0 {
					t.Fatal("expired key counted")
				}
			case "replace":
				if ok, err := m.StoreBytesLimited([]byte("key"), []byte("value"), 0, true, false); !ok || err != nil {
					t.Fatalf("replacement: %v %v", ok, err)
				}
				m.Expire("key", 0)
			}
			if m.Len() != 0 || m.UsedMemory() != 0 {
				t.Fatalf("keys=%d memory=%d", m.Len(), m.UsedMemory())
			}
		})
	}
}

func TestJanitorReleasesMemory(t *testing.T) {
	m := NewMap()
	defer m.Close()
	i := NewItem("key", "value", time.Hour)
	i.ExpiresAt = time.Now().Add(-time.Second)
	m.Store(i)
	deadline := time.Now().Add(3 * time.Second)
	for m.UsedMemory() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if m.UsedMemory() != 0 || m.Len() != 0 {
		t.Fatal("janitor did not release memory")
	}
}

func TestMemoryAccountingConcurrentClearAndWrites(t *testing.T) {
	var m Map
	m.SetMaxMemory(10000)
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("%d:%d", worker, i%20)
				m.StoreBytes([]byte(key), []byte("value"), 0, false, false)
				if i%3 == 0 {
					m.Delete(key)
				}
				if i%50 == 0 {
					m.Clear()
				}
			}
		}(worker)
	}
	wg.Wait()
	var want int64
	for _, item := range m.List() {
		want += itemBytes(item.Key, item.Data)
	}
	for i := range m.shards {
		for _, slot := range m.shards[i].slots {
			if slot.item == nil {
				want += scanSlotBytes
			}
		}
	}
	if m.UsedMemory() != want || want > 10000 {
		t.Fatalf("memory=%d actual=%d", m.UsedMemory(), want)
	}
	m.Clear()
	if m.UsedMemory() != 0 {
		t.Fatal("final clear leaked budget")
	}
}

func TestMemoryBudgetChargesRetainedScanSlots(t *testing.T) {
	var m Map
	fillOneShard(&m, 100)
	keys := m.Keys()
	for _, key := range keys[1:] {
		m.Delete(key)
	}
	item, _ := m.Get(keys[0])
	want := itemBytes(item.Key, item.Data) + 99*scanSlotBytes
	if m.UsedMemory() != want {
		t.Fatalf("retained slot accounting=%d want%d", m.UsedMemory(), want)
	}
	// Reusing an existing slot only needs the item's remaining memory.
	key := keys[1]
	m.SetMaxMemory(want + itemBytes(key, "v") - scanSlotBytes)
	if ok, err := m.StoreLimited(NewItem(key, "v", 0)); !ok || err != nil {
		t.Fatalf("slot reuse: %v %v", ok, err)
	}
	if _, err := m.StoreLimited(NewItem(keys[2], "v", 0)); err == nil {
		t.Fatal("retained slots bypassed memory cap")
	}
	m.Delete(key)
	m.Delete(keys[0])
	if m.UsedMemory() != 0 || len(m.shards[0].slots) != 0 {
		t.Fatal("empty shard retained scan index")
	}
}
