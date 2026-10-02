package disk

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func TestSnapshotStatsSuccessAndFailure(t *testing.T) {
	db, err := OpenKVSDB(filepath.Join(t.TempDir(), "snapshot"))
	if err != nil {
		t.Fatal(err)
	}
	data := datastructure.NewMap()
	t.Cleanup(data.Close)
	if stats := db.SnapshotStats(); stats.LastStatus != "never" || stats.LastSuccess != 0 || stats.Failures != 0 || stats.InProgress {
		t.Fatalf("initial stats=%+v", stats)
	}
	before := time.Now().Unix()
	if err := db.Write(data); err != nil {
		t.Fatal(err)
	}
	if stats := db.SnapshotStats(); stats.LastStatus != "ok" || stats.LastSuccess < before || stats.Failures != 0 || stats.InProgress {
		t.Fatalf("successful stats=%+v", stats)
	}
	if err := db.Clear(); err != nil {
		t.Fatal(err)
	}
	lastSuccess := db.SnapshotStats().LastSuccess
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.Write(data); err == nil {
		t.Fatal("closed snapshot unexpectedly saved")
	}
	if stats := db.SnapshotStats(); stats.LastStatus != "err" || stats.LastSuccess != lastSuccess || stats.Failures != 1 || stats.InProgress {
		t.Fatalf("failed stats=%+v", stats)
	}
	if err := db.Clear(); err == nil || db.SnapshotStats().Failures != 2 {
		t.Fatal("failed Clear not counted")
	}
}

func TestSnapshotStatsConcurrentAndWithoutDiskLock(t *testing.T) {
	db, err := OpenKVSDB(filepath.Join(t.TempDir(), "snapshot"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	data := datastructure.NewMap()
	t.Cleanup(data.Close)
	db.mu.Lock()
	done := make(chan struct{})
	go func() { db.SnapshotStats(); close(done) }()
	select {
	case <-done:
		db.mu.Unlock()
	case <-time.After(time.Second):
		db.mu.Unlock()
		t.Fatal("snapshot stats waits for disk lock")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			if err := db.Write(data); err != nil {
				t.Error(err)
			}
		}
	}()
	for i := 0; i < 1000; i++ {
		db.SnapshotStats()
	}
	wg.Wait()
}
