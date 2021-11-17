package datastructure_test

import (
	"testing"

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
