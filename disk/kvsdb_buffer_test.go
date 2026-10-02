package disk

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func TestBufferedSnapshotRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dump.kvsdb")
	db, err := OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	var data datastructure.Map
	value := strings.Repeat("x\x00\r\n", 32*1024)
	data.Store(datastructure.NewItem("large", value, 0))
	data.Store(datastructure.NewItem("tail", "last bytes", 0))
	if err := db.Write(&data); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	restored, err := db.Read()
	if err != nil {
		t.Fatal(err)
	}
	if restored.Len() != 2 {
		t.Fatalf("count=%d, want 2", restored.Len())
	}
	for key, want := range map[string]string{"large": value, "tail": "last bytes"} {
		item, ok := restored.Get(key)
		if !ok || item.Data != want {
			t.Fatalf("bad restored value for %s", key)
		}
	}
}

func TestSnapshotAfterCloseReturnsError(t *testing.T) {
	db, err := OpenKVSDB(filepath.Join(t.TempDir(), "dump.kvsdb"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	var data datastructure.Map
	if err := db.Write(&data); err == nil {
		t.Fatal("Write after Close succeeded")
	}
}

func TestConcurrentSnapshotAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dump.kvsdb")
	db, err := OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	var data datastructure.Map
	data.Store(datastructure.NewItem("key", "value", 0))
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if err := db.Write(&data); err != nil {
					t.Error(err)
				}
				if err := db.Clear(); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	if err := db.Write(&data); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	restored, err := db.Read()
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := restored.Get("key"); !ok || item.Data != "value" {
		t.Fatal("final snapshot corrupted")
	}
}
