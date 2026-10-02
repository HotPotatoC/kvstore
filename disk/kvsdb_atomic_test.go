package disk

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/vmihailenco/msgpack/v5"
)

func TestSnapshotRepeatedWriteReadAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dump.kvsdb")
	db, err := OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, value := range []string{"first snapshot", "second", "x"} {
		var data datastructure.Map
		data.Store(datastructure.NewItem("key", value, 0))
		if err := db.Write(&data); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			loaded, err := db.Read()
			if err != nil {
				t.Fatal(err)
			}
			defer loaded.Close()
			if item, ok := loaded.Get("key"); !ok || item.Data != value || loaded.Len() != 1 {
				t.Fatalf("loaded %v, found %t, count %d", item, ok, loaded.Len())
			}
		}
		if err := db.Clear(); err != nil {
			t.Fatal(err)
		}
		loaded, err := db.Read()
		if loaded != nil {
			defer loaded.Close()
		}
		if err != nil || loaded.Len() != 0 {
			t.Fatalf("Read after Clear: data=%v, err=%v", loaded, err)
		}
		contents, err := os.ReadFile(path)
		if err != nil || len(contents) != 0 {
			t.Fatalf("Clear did not replace snapshot: bytes=%d, err=%v", len(contents), err)
		}
	}
}

func TestSnapshotRenameFailurePreservesPreviousFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dump.kvsdb")
	db, err := OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var data datastructure.Map
	data.Store(datastructure.NewItem("key", "original", 0))
	if err := db.Write(&data); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(dir, "previous.kvsdb")
	if err := os.Rename(path, backup); err != nil {
		t.Fatal(err)
	}
	// A directory at the destination deterministically rejects the rename,
	// including when tests run with permission to bypass file mode checks.
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	data.Store(datastructure.NewItem("key", "replacement", 0))
	for _, operation := range []func() error{func() error { return db.Write(&data) }, db.Clear} {
		if err := operation(); err == nil {
			t.Fatal("snapshot replacement unexpectedly succeeded")
		}
		loaded, err := db.Read()
		if err != nil {
			t.Fatal(err)
		}
		defer loaded.Close()
		if item, ok := loaded.Get("key"); !ok || item.Data != "original" {
			t.Fatalf("failed replacement changed open snapshot: %v, found %t", item, ok)
		}
		after, err := os.ReadFile(backup)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("failed replacement changed previous file: err=%v", err)
		}
		leftovers, err := filepath.Glob(filepath.Join(dir, ".dump.kvsdb-*"))
		if err != nil || len(leftovers) != 0 {
			t.Fatalf("temporary snapshots left behind: %v, err=%v", leftovers, err)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(backup, path); err != nil {
		t.Fatal(err)
	}
	if err := db.Write(&data); err != nil {
		t.Fatalf("write after failure: %v", err)
	}
	loaded, err := db.Read()
	if err != nil {
		t.Fatal(err)
	}
	defer loaded.Close()
	if item, ok := loaded.Get("key"); !ok || item.Data != "replacement" {
		t.Fatalf("write after failure returned %v, found %t", item, ok)
	}
}

func TestReadWithLimitPreservesSnapshotOnFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dump.kvsdb")
	db, err := OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var data datastructure.Map
	data.Store(datastructure.NewItem("key", "value", 0))
	if err := db.Write(&data); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded, err := db.ReadWithLimit(1); err == nil || loaded != nil {
		t.Fatalf("limited restore returned data=%v, err=%v", loaded, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("limited restore changed snapshot: err=%v", err)
	}
	loaded, err := db.Read()
	if loaded != nil {
		defer loaded.Close()
	}
	if err != nil || loaded.Len() != 1 {
		t.Fatalf("unlimited restore after failure: data=%v, err=%v", loaded, err)
	}
}

func TestReadWithLimitSkipsExpiredItems(t *testing.T) {
	db, err := OpenKVSDB(filepath.Join(t.TempDir(), "dump.kvsdb"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	expired := datastructure.NewItem("expired", "value too large for limit", time.Second)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	if err := msgpack.NewEncoder(db.file).Encode(expired); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.ReadWithLimit(1)
	if loaded != nil {
		defer loaded.Close()
	}
	if err != nil || loaded.Len() != 0 {
		t.Fatalf("expired restore: data=%v, err=%v", loaded, err)
	}
}

func TestClosedSnapshotOperationsReturnErrors(t *testing.T) {
	db, err := OpenKVSDB(filepath.Join(t.TempDir(), "dump.kvsdb"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Read(); err == nil {
		t.Fatal("Read after Close succeeded")
	}
	if err := db.Clear(); err == nil {
		t.Fatal("Clear after Close succeeded")
	}
	if err := db.Close(); err == nil {
		t.Fatal("second Close succeeded")
	}
}
