package datastructure_test

import (
	"sort"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func Test_SetGet(t *testing.T) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))

	if v, ok := hmap.Get("key"); !ok {
		t.Errorf("Get failed")
	} else if string(v.Data.([]byte)) != "value" {
		t.Errorf("Get failed")
	}
}

func Test_Delete(t *testing.T) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))

	if n := hmap.Delete("key"); n != 1 {
		t.Errorf("Delete failed")
	}
}

var testData = []struct {
	key   string
	value []byte
}{
	{"hello", []byte("value")},
	{"hallo", []byte("value")},
	{"hbllo", []byte("value")},
	{"hllo", []byte("value")},
	{"hxllo", []byte("value")},
	{"heeeeello", []byte("value")},
}

func fillMap(hmap *datastructure.Map) {
	for _, v := range testData {
		hmap.Store(datastructure.NewItem(v.key, v.value, 0))
	}

	sort.Slice(testData, func(i, j int) bool {
		return testData[i].key < testData[j].key
	})
}

func Test_DeletePattern(t *testing.T) {
	hmap := datastructure.NewMap()
	fillMap(hmap)

	if n := hmap.Delete("*"); n != 6 {
		t.Errorf("Delete * failed: expected 6, got %d", n)
	}

	fillMap(hmap)

	if n := hmap.Delete("h[a-e]llo"); n != 3 {
		// Should delete hello, hallo and hbllo, with hxllo, hllo and heeeeello not deleted
		t.Errorf("Delete h[a-e]llo failed expected 3, got %d", n)
	}

	if n := hmap.Delete("h*llo"); n != 3 {
		// Should delete hxllo, hllo, and heeeeello
		t.Errorf("Delete h?llo failed expected 1, got %d", n)
	}
}

func Test_Len(t *testing.T) {
	hmap := datastructure.NewMap()
	fillMap(hmap)

	if hmap.Len() != 6 {
		t.Errorf("Len failed: expected 6, got %d", hmap.Len())
	}
}

func Test_List(t *testing.T) {
	hmap := datastructure.NewMap()
	fillMap(hmap)

	list := hmap.List()
	if len(list) != 6 {
		t.Errorf("List len failed: expected 6, got %d", len(list))
	}

	for _, v := range testData {
		if _, ok := list[v.key]; !ok {
			t.Errorf("List failed")
		}
	}
}

func Test_Exists(t *testing.T) {
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))

	if !hmap.Exists("key") {
		t.Errorf("Exists failed")
	}

	if hmap.Exists("key2") {
		t.Errorf("Exists failed")
	}
}

func Test_Keys(t *testing.T) {
	hmap := datastructure.NewMap()
	fillMap(hmap)

	keys := hmap.Keys()
	if len(keys) != 6 {
		t.Errorf("Keys len failed: expected 6, got %d", len(keys))
	}

	sort.Strings(keys)
	for i, v := range testData {
		if keys[i] != v.key {
			t.Errorf("Keys failed")
		}
	}
}

func Test_KeysWithPattern(t *testing.T) {
	hmap := datastructure.NewMap()
	fillMap(hmap)

	tc := []struct {
		pattern string
		keys    []string
	}{
		{"*", []string{"hello", "hallo", "hbllo", "hllo", "hxllo", "heeeeello"}},
		{"h[a-e]llo", []string{"hello", "hallo", "hbllo"}},
		{"h?llo", []string{"hello", "hallo", "hbllo", "hxllo"}},
		{"?[a-e]*", []string{"hello", "hallo", "hbllo", "heeeeello"}},
	}

	for _, tt := range tc {
		t.Run(tt.pattern, func(t *testing.T) {
			keys := hmap.KeysWithPattern(tt.pattern)
			if len(keys) != len(tt.keys) {
				t.Errorf("KeysWithPattern len failed: expected %d, got %d", len(tt.keys), len(keys))
			}
			sort.Strings(tt.keys)
			sort.Strings(keys)
			for i, v := range tt.keys {
				if keys[i] != v {
					t.Errorf("KeysWithPattern failed: expected %s, got %s", v, keys[i])
				}
			}
		})
	}
}

func Test_Clear(t *testing.T) {
	hmap := datastructure.NewMap()
	fillMap(hmap)

	if n := hmap.Clear(); n != 6 {
		t.Errorf("Clear failed: expected 6, got %d", n)
	}

	if hmap.Len() != 0 {
		t.Errorf("Clear failed: expected 0, got %d", hmap.Len())
	}
}

func Test_SetExpireItem(t *testing.T) {
	t.Parallel()
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 0))

	n := hmap.Expire("key", 1*time.Second)
	if n == 0 {
		t.Errorf("Expire failed")
	}

	v, ok := hmap.Get("key")
	if !ok {
		t.Errorf("Get failed")
	}

	if !v.HasFlag(datastructure.ItemFlagExpireXX) {
		t.Errorf("Expire failed")
	}

	time.Sleep(1 * time.Second)
	if _, ok := hmap.Get("key"); ok {
		t.Errorf("Expire failed")
	}
}

func Test_GetExpired(t *testing.T) {
	t.Parallel()
	hmap := datastructure.NewMap()
	hmap.Store(datastructure.NewItem("key", []byte("value"), 1*time.Second))

	if v, ok := hmap.Get("key"); !ok {
		t.Errorf("Get failed")
	} else if string(v.Data.([]byte)) != "value" {
		t.Errorf("Get failed")
	}

	time.Sleep(1 * time.Second)
	if _, ok := hmap.Get("key"); ok {
		t.Errorf("Item should have expired")
	}
}
