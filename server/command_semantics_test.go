package server

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func semanticMap(t *testing.T) *datastructure.Map {
	t.Helper()
	db := datastructure.NewMap()
	t.Cleanup(db.Close)
	return db
}

func semanticCommand(db *datastructure.Map, command string, args ...string) string {
	c := &client.Client{DB: db, Command: command, Argc: len(args)}
	for _, arg := range args {
		c.Argv = append(c.Argv, []byte(arg))
	}
	var response bytes.Buffer
	switch command {
	case "set":
		setCommand(c, &response)
	case "expire":
		expireCommand(c, &response)
	case "pexpire":
		pexpireCommand(c, &response)
	case "del":
		delCommand(c, &response)
	}
	return response.String()
}

func TestSETInvalidOptionsDoNotMutate(t *testing.T) {
	for _, options := range [][]string{
		{"unknown"}, {"EX"}, {"NX", "XX"}, {"EX", "30", "PX", "30"},
		{"EX", "30", "unknown"}, {"EX", "0"}, {"PX", "-1"},
		{"EX", "9223372037"}, {"PX", "9223372036855"},
		{"EX", "9223372036854775808"}, {"PX", "18446744073709551617"},
		{"EX", ""}, {"EX", "-"}, {"PX", "+1"}, {"EX", "01"}, {"PX", "1.5"},
	} {
		t.Run(strings.Join(options, "/"), func(t *testing.T) {
			for _, exists := range []bool{false, true} {
				db := semanticMap(t)
				previous := datastructure.NewItem("key", "previous", time.Minute)
				if exists {
					db.Store(previous)
				}
				args := append([]string{"key", "replacement"}, options...)
				if response := semanticCommand(db, "set", args...); !strings.HasPrefix(response, "-ERR ") {
					t.Fatalf("SET %v response=%q", options, response)
				}
				item, found := db.Get("key")
				if found != exists || (found && item != previous) {
					t.Fatal("invalid SET changed the key or its expiration")
				}
			}
		})
	}
}

func TestSETConditionsCombineWithExpiry(t *testing.T) {
	for _, condition := range []string{"NX", "XX"} {
		for _, expiry := range []string{"EX", "PX"} {
			for _, conditionFirst := range []bool{true, false} {
				order := "expiry-first"
				if conditionFirst {
					order = "condition-first"
				}
				t.Run(condition+"/"+expiry+"/"+order, func(t *testing.T) {
					db := semanticMap(t)
					if condition == "XX" {
						db.Store(datastructure.NewItem("key", "previous", 0))
					}
					options := []string{expiry, "60000", condition}
					if conditionFirst {
						options = []string{condition, expiry, "60000"}
					}
					args := append([]string{"key", "value"}, options...)
					before := time.Now()
					if response := semanticCommand(db, "set", args...); response != "+OK\r\n" {
						t.Fatalf("SET %v response=%q", options, response)
					}
					item, found := db.Get("key")
					wantTTL := 60000 * time.Second
					if expiry == "PX" {
						wantTTL = 60000 * time.Millisecond
					}
					if !found || item.Data != "value" || item.ExpiresAt.Before(before.Add(wantTTL)) || item.ExpiresAt.After(time.Now().Add(wantTTL)) {
						t.Fatal("SET did not publish the value and expected expiry together")
					}
					if condition == "NX" {
						if response := semanticCommand(db, "set", args...); response != "$-1\r\n" {
							t.Fatalf("second SET NX response=%q", response)
						}
						if current, _ := db.Get("key"); current != item {
							t.Fatal("failed SET NX changed expiry")
						}
					} else if response := semanticCommand(db, "set", append([]string{"missing", "value"}, options...)...); response != "$-1\r\n" {
						t.Fatalf("SET XX missing response=%q", response)
					}
				})
			}
		}
	}
}

func TestSETRepeatedCompatibleOptions(t *testing.T) {
	db := semanticMap(t)
	if response := semanticCommand(db, "set", "key", "value", "nx", "NX", "EX", "bad", "ex", "30"); response != "+OK\r\n" {
		t.Fatalf("compatible repeated SET options response=%q", response)
	}
}

func TestExpireRejectsInvalidArgumentsWithoutMutation(t *testing.T) {
	for _, command := range []string{"expire", "pexpire"} {
		for _, args := range [][]string{
			{}, {"key"}, {"key", "30", "NX"}, {"key", "30", "unknown"},
			{"key", ""}, {"key", "-"}, {"key", "1.5"}, {"key", "+1"},
			{"key", "9223372036854775808"}, {"key", "-9223372036854775809"},
			{"key", "9223372036855"},
		} {
			t.Run(command+"/"+strings.Join(args, "/"), func(t *testing.T) {
				db := semanticMap(t)
				previous := datastructure.NewItem("key", "value", time.Minute)
				db.Store(previous)
				if response := semanticCommand(db, command, args...); !strings.HasPrefix(response, "-ERR ") {
					t.Fatalf("%s %v response=%q", command, args, response)
				}
				if item, found := db.Get("key"); !found || item != previous {
					t.Fatal("invalid expiration changed the item")
				}
			})
		}
		if command == "expire" {
			for _, ttl := range []string{"9223372037", "-9223372036854775808", "-9223372036854776"} {
				db := semanticMap(t)
				previous := datastructure.NewItem("key", "value", 0)
				db.Store(previous)
				if response := semanticCommand(db, command, "key", ttl); !strings.HasPrefix(response, "-ERR ") {
					t.Fatalf("EXPIRE duration overflow response=%q", response)
				}
				if item, found := db.Get("key"); !found || item != previous {
					t.Fatal("overflowing expiration changed the item")
				}
			}
		}
	}
}

func TestExpireImmediateDeletionAndExpiredAbsence(t *testing.T) {
	for _, command := range []string{"expire", "pexpire"} {
		ttls := []string{"0", "-1", "-9223372036854775"}
		if command == "pexpire" {
			ttls = append(ttls, "-9223372036854775808")
		}
		for _, ttl := range ttls {
			db := semanticMap(t)
			db.Store(datastructure.NewItem("key", "value", 0))
			db.Store(datastructure.NewItem("untouched", "value", 0))
			if response := semanticCommand(db, command, "key", ttl); response != ":1\r\n" {
				t.Fatalf("%s %s response=%q", command, ttl, response)
			}
			if _, found := db.Get("key"); found || db.Len() != 1 {
				t.Fatal("nonpositive expiration did not delete immediately")
			}
		}
		db := semanticMap(t)
		expired := datastructure.NewItem("expired", "value", time.Minute)
		expired.ExpiresAt = time.Now().Add(-time.Second)
		db.Store(expired)
		if response := semanticCommand(db, command, "expired", "30"); response != ":0\r\n" {
			t.Fatalf("%s resurrected expired key: %q", command, response)
		}
		if _, found := db.Get("expired"); found {
			t.Fatal("expiration resurrected expired key")
		}
	}
}

func TestDELLiteralMultipleKeys(t *testing.T) {
	db := semanticMap(t)
	for _, key := range []string{"*", "key:*", "key:one", "[abc]", "a", "untouched"} {
		db.Store(datastructure.NewItem(key, "value", 0))
	}
	if response := semanticCommand(db, "del", "*", "key:*", "[abc]", "*", "missing"); response != ":3\r\n" {
		t.Fatalf("literal multi-key DEL response=%q", response)
	}
	if response := semanticCommand(db, "del", "key:*", "a*"); response != ":0\r\n" {
		t.Fatalf("missing pattern DEL response=%q", response)
	}
	for _, key := range []string{"key:one", "a", "untouched"} {
		if _, found := db.Get(key); !found {
			t.Fatalf("literal DEL deleted %q", key)
		}
	}
	expired := datastructure.NewItem("expired", "value", time.Minute)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	db.Store(expired)
	if response := semanticCommand(db, "del", "expired"); response != ":0\r\n" {
		t.Fatalf("DEL counted an expired key: %q", response)
	}
	if response := semanticCommand(db, "del"); !strings.HasPrefix(response, "-ERR ") {
		t.Fatalf("DEL missing key response=%q", response)
	}
}

func TestSETMemoryLimitReturnsOOMWithoutMutation(t *testing.T) {
	db := semanticMap(t)
	previous := datastructure.NewItem("key", "value", time.Minute)
	db.Store(previous)
	db.SetMaxMemory(db.UsedMemory())
	for _, args := range [][]string{
		{"another", "value"}, {"key", "longer replacement", "PX", "30000"},
	} {
		if response := semanticCommand(db, "set", args...); !strings.HasPrefix(response, "-OOM ") {
			t.Fatalf("SET at memory limit response=%q", response)
		}
		if item, found := db.Get("key"); !found || item != previous {
			t.Fatal("OOM changed the existing item or its expiry")
		}
		if _, found := db.Get("another"); found {
			t.Fatal("OOM published a new key")
		}
	}
	if response := semanticCommand(db, "set", "key", "longer replacement", "NX", "EX", "30"); response != "$-1\r\n" {
		t.Fatalf("failed condition at memory limit response=%q", response)
	}
}
