package server

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func testScan(db *datastructure.Map, outputLimit int, args ...string) string {
	c := &client.Client{DB: db, OutputLimit: outputLimit, Argc: len(args)}
	for _, arg := range args {
		c.Argv = append(c.Argv, []byte(arg))
	}
	var response bytes.Buffer
	scanCommand(c, &response)
	return response.String()
}

func readScanPage(t *testing.T, reader *protocol.Reader) (string, []string) {
	t.Helper()
	object, err := reader.ReadObject()
	if err != nil {
		t.Fatal(err)
	}
	page, ok := object.([]any)
	if !ok || len(page) != 2 {
		t.Fatalf("SCAN response is not a two-element array: %#v", object)
	}
	cursor, ok := page[0].([]byte)
	if !ok {
		t.Fatalf("SCAN cursor is not a bulk string: %#v", page[0])
	}
	if _, err := strconv.ParseUint(string(cursor), 10, 64); err != nil {
		t.Fatalf("SCAN cursor=%q: %v", cursor, err)
	}
	array, ok := page[1].([]any)
	if !ok {
		t.Fatalf("SCAN keys are not an array: %#v", page[1])
	}
	keys := make([]string, len(array))
	for i, value := range array {
		key, ok := value.([]byte)
		if !ok {
			t.Fatalf("SCAN key is not a bulk string: %#v", value)
		}
		keys[i] = string(key)
	}
	return string(cursor), keys
}

func TestSCANRejectsInvalidArgumentsWithoutMutation(t *testing.T) {
	db := semanticMap(t)
	previous := datastructure.NewItem("key", "value", time.Hour)
	db.Store(previous)
	for _, args := range [][]string{
		nil, {""}, {"-1"}, {"+1"}, {" 0"}, {"0x1"}, {"1.0"}, {"18446744073709551616"},
		{"0", "MATCH"}, {"0", "COUNT"}, {"0", "unknown", "value"},
		{"0", "TYPE", "string"}, {"0", "COUNT", "0"}, {"0", "COUNT", "-1"},
		{"0", "COUNT", ""}, {"0", "COUNT", "1.5"}, {"0", "COUNT", "+1"},
		{"0", "COUNT", "01"}, {"0", "COUNT", "-0"},
		{"0", "COUNT", "9223372036854775808"}, {"0", "MATCH", "*", "extra"},
	} {
		t.Run(strings.Join(args, "/"), func(t *testing.T) {
			if response := testScan(db, 0, args...); !strings.HasPrefix(response, "-ERR ") {
				t.Fatalf("SCAN %v response=%q", args, response)
			}
			if item, found := db.Get("key"); !found || item != previous {
				t.Fatal("invalid SCAN changed the item")
			}
		})
	}
}

func TestSCANEmptyAndLargeCount(t *testing.T) {
	db := semanticMap(t)
	for _, cursor := range []string{"0", "000", "18446744073709551615"} {
		if response := testScan(db, 0, cursor); response != "*2\r\n$1\r\n0\r\n*0\r\n" {
			t.Fatalf("empty SCAN %s response=%q", cursor, response)
		}
	}
	for i := 0; i < 1100; i++ {
		db.Store(datastructure.NewItem(fmt.Sprintf("key:%04d", i), "value", 0))
	}
	response := testScan(db, 0, "0", "cOuNt", "9223372036854775807")
	next, keys := readScanPage(t, protocol.NewReader(strings.NewReader(response)))
	if next == "0" || len(keys) == 0 || len(keys) > datastructure.MaxScanCount {
		t.Fatalf("large count next=%s keys=%d", next, len(keys))
	}
}

func TestSCANPagesAndBytePatterns(t *testing.T) {
	db := semanticMap(t)
	matching := map[string]bool{"a/b": true, "a\x00b": true, "a\xffb": true}
	for key := range matching {
		db.Store(datastructure.NewItem(key, "value", 0))
	}
	for _, key := range []string{"ignored", "a-long-b", "a*b"} {
		db.Store(datastructure.NewItem(key, "value", 0))
	}
	matching["a*b"] = true
	cursor := "0"
	seen := make(map[string]bool)
	for pages := 0; ; pages++ {
		if pages > 100 {
			t.Fatal("SCAN did not terminate")
		}
		response := testScan(db, 0, cursor, "mAtCh", "a?b", "CoUnT", "1")
		next, keys := readScanPage(t, protocol.NewReader(strings.NewReader(response)))
		if len(keys) > 1 {
			t.Fatal("COUNT 1 inspected more than one key")
		}
		for _, key := range keys {
			if !matching[key] || seen[key] {
				t.Fatalf("unexpected or duplicate key=%q", key)
			}
			seen[key] = true
		}
		if next == "0" {
			break
		}
		cursor = next
	}
	if len(seen) != len(matching) {
		t.Fatalf("scan returned %v want %v", seen, matching)
	}
	cursor = "0"
	emptyContinuation := false
	for pages := 0; ; pages++ {
		if pages > 100 {
			t.Fatal("filtered SCAN did not terminate")
		}
		next, keys := readScanPage(t, protocol.NewReader(strings.NewReader(testScan(db, 0, cursor, "MATCH", "no-match", "COUNT", "1"))))
		if len(keys) != 0 {
			t.Fatalf("nonmatching SCAN returned %v", keys)
		}
		if next == "0" {
			break
		}
		emptyContinuation = true
		cursor = next
	}
	if !emptyContinuation {
		t.Fatal("nonmatching pages did not retain a continuation cursor")
	}
}

func TestSCANOutputAndMatchWorkLimits(t *testing.T) {
	db := semanticMap(t)
	db.Store(datastructure.NewItem(strings.Repeat("a", 200), "value", 0))
	if response := testScan(db, 100, "0"); response != "-ERR response exceeds output limit\r\n" {
		t.Fatalf("output-limited SCAN=%q", response)
	}
	db.Clear()
	db.Store(datastructure.NewItem("a", "value", 0))
	pattern := "[" + strings.Repeat("a", 1<<20) + "]"
	if response := testScan(db, 0, "0", "MATCH", pattern); response != "-ERR scan match work limit exceeded\r\n" {
		t.Fatalf("work-limited SCAN=%q", response)
	}
}

func TestSCANTrafficWorkerAndPipelineOrder(t *testing.T) {
	s := newTestServer(t)
	s.DB.Store(datastructure.NewItem("key", "value", 0))
	releases := make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	conn := openTestConn(s)
	conn.input = append(protocol.MakeCommand("sCaN", "0"), protocol.MakeCommand("PING")...)
	traffic(t, s, conn)
	if conn.output.Len() != 0 || !conn.ctx.(*connectionState).running {
		t.Fatal("SCAN did not offload to a worker")
	}
	other := openTestConn(s)
	other.input = protocol.MakeCommand("PING")
	traffic(t, s, other)
	if other.output.String() != "+PONG\r\n" {
		t.Fatal("SCAN blocked another client")
	}
	for _, release := range releases {
		close(release)
	}
	waitWake(t, conn)
	traffic(t, s, conn)
	reader := protocol.NewReader(bytes.NewReader(conn.output.Bytes()))
	cursor, keys := readScanPage(t, reader)
	if cursor != "0" || len(keys) != 1 || keys[0] != "key" {
		t.Fatalf("SCAN cursor=%s keys=%v", cursor, keys)
	}
	if object, err := reader.ReadObject(); err != nil || object != "PONG" {
		t.Fatalf("pipeline PING=%v err=%v", object, err)
	}
}
