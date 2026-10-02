package disk

import (
	"path/filepath"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/vmihailenco/msgpack/v5"
)

func TestReadLegacyItem(t *testing.T) {
	for _, value := range []any{"value", []byte("value")} {
		db, err := OpenKVSDB(filepath.Join(t.TempDir(), "dump.kvsdb"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { db.Close() })
		// The old Item.Data interface could encode either strings or bytes.
		legacy := map[string]any{"Key": "key", "Data": value, "Flag": datastructure.ItemFlagExpireNX}
		if err := msgpack.NewEncoder(db.file).Encode(legacy); err != nil {
			t.Fatal(err)
		}
		if _, err := db.file.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		data, err := db.Read()
		if err != nil {
			t.Fatal(err)
		}
		if item, ok := data.Get("key"); !ok || item.Data != "value" {
			t.Fatalf("legacy value = %v, found = %t", item, ok)
		}
		data.Store(datastructure.NewItem("new", "new value", 0))
		if err := db.Write(data); err != nil {
			t.Fatal(err)
		}
		if _, err := db.file.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		loaded, err := db.Read()
		if err != nil {
			t.Fatal(err)
		}
		if item, ok := loaded.Get("new"); !ok || item.Data != "new value" || loaded.Len() != 2 {
			t.Fatalf("round trip value = %v, found = %t, count = %d", item, ok, loaded.Len())
		}
	}
}
