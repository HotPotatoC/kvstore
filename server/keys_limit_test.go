package server

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func TestKeysCommandLimit(t *testing.T) {
	var db datastructure.Map
	db.Store(datastructure.NewItem("key", "value", 0))
	c := &client.Client{DB: &db, Argc: 1, Argv: [][]byte{[]byte("*")}, OutputLimit: 66}
	var res bytes.Buffer
	keysCommand(c, &res)
	if res.String() != "-ERR response exceeds output limit\r\n" {
		t.Fatalf("response=%q", res.String())
	}
	c.OutputLimit = 67
	res.Reset()
	keysCommand(c, &res)
	if res.String() != "*1\r\n$3\r\nkey\r\n" {
		t.Fatalf("response=%q", res.String())
	}
}

func BenchmarkKeysLimit(b *testing.B) {
	var db datastructure.Map
	for i := 0; i < 100000; i++ {
		db.Store(datastructure.NewItem(fmt.Sprintf("key:%012d", i), "value", 0))
	}
	c := &client.Client{DB: &db, Argc: 1, Argv: [][]byte{[]byte("*")}, OutputLimit: 128}
	var res bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res.Reset()
		keysCommand(c, &res)
		if res.String() != "-ERR response exceeds output limit\r\n" {
			b.Fatal("expected limit error")
		}
	}
}
