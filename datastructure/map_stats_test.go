package datastructure

import (
	"testing"
	"time"
)

func TestKeyspaceStatsFollowMutations(t *testing.T) {
	m := NewMap()
	t.Cleanup(m.Close)
	check := func(wantKeys, wantExpires int64) {
		t.Helper()
		if keys, expires := m.KeyspaceStats(); keys != wantKeys || expires != wantExpires {
			t.Fatalf("keys=%d expires=%d want %d/%d", keys, expires, wantKeys, wantExpires)
		}
	}
	m.Store(NewItem("key", "value", 0))
	check(1, 0)
	m.Expire("key", time.Hour)
	check(1, 1)
	m.Expire("key", time.Hour)
	check(1, 1)
	m.StoreIfAbsent(NewItem("key", "ignored", 0))
	check(1, 1)
	m.Store(NewItem("key", "replacement", 0))
	check(1, 0)
	expired := NewItem("expired", "value", time.Hour)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	m.Store(expired)
	check(2, 1)
	m.Get("expired")
	check(1, 0)
	m.Store(NewItem("second", "value", time.Hour))
	m.Delete("second")
	check(1, 0)
	m.Clear()
	check(0, 0)
}
