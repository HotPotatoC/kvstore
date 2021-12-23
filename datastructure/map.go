package datastructure

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Map is a thread-safe map.
type Map struct {
	items sync.Map
	nSize int64
}

// NewMap returns a new Map.
func NewMap() *Map {
	m := &Map{}

	go m.janitor()

	return m
}

// Store stores a new key-value pair.
func (m *Map) Store(v *Item) {
	m.items.Store(v.Key, v)
	atomic.AddInt64(&m.nSize, 1)
}

// Expire sets the expiration time of the key.
func (m *Map) Expire(k string, ttl time.Duration) int64 {
	v, ok := m.items.Load(k)
	if !ok {
		return 0
	}

	item := v.(*Item)

	if item.HasFlag(ItemFlagExpireNX) {
		item.RemoveFlag(ItemFlagExpireNX)
	}

	item.AddFlag(ItemFlagExpireXX)
	item.ExpiresAt = time.Now().Add(ttl)

	return atomic.LoadInt64(&m.nSize)
}

// Get returns the value of the key.
func (m *Map) Get(k string) (*Item, bool) {
	v, ok := m.items.Load(k)
	if !ok {
		return nil, false
	}

	if v.(*Item).HasFlag(ItemFlagExpireXX) && time.Now().After(v.(*Item).ExpiresAt) {
		m.items.Delete(k)
		atomic.AddInt64(&m.nSize, -1)
		return nil, false
	}

	return v.(*Item), true
}

// Delete deletes the key.
func (m *Map) Delete(k string) int64 {
	if _, ok := m.items.Load(k); !ok {
		return 0
	}
	m.items.Delete(k)

	prevNSize := atomic.LoadInt64(&m.nSize)
	atomic.AddInt64(&m.nSize, -1)

	// return the deleted amount
	return prevNSize - atomic.LoadInt64(&m.nSize)
}

// Len returns the number of items in the map.
func (m *Map) Len() int64 {
	return atomic.LoadInt64(&m.nSize)
}

// List returns all keys and values in a map
func (m *Map) List() map[string]*Item {
	items := make(map[string]*Item)

	m.items.Range(func(key, value interface{}) bool {
		items[key.(string)] = value.(*Item)
		return true
	})

	return items
}

// Keys returns the keys of the map.
func (m *Map) Keys() []string {
	var keys []string
	m.items.Range(func(k, v interface{}) bool {
		item := v.(*Item)
		if item.HasFlag(ItemFlagExpireNX) || time.Now().Before(item.ExpiresAt) {
			keys = append(keys, k.(string))
		}
		return true
	})
	return keys
}

// KeysWithPattern returns the keys of the map that match the pattern.
func (m *Map) KeysWithPattern(pattern string) []string {
	var keys []string
	m.items.Range(func(k, v interface{}) bool {
		key, item := k.(string), v.(*Item)
		if match, _ := filepath.Match(pattern, key); match && (item.HasFlag(ItemFlagExpireNX) || time.Now().Before(item.ExpiresAt)) {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}

// Exists checks if the key exists in the map.
func (m *Map) Exists(k string) bool {
	_, ok := m.items.Load(k)
	return ok
}

// Clear clears the map.
func (m *Map) Clear() int64 {
	prevNSize := atomic.LoadInt64(&m.nSize)
	var delNum int64
	m.items.Range(func(k, v interface{}) bool {
		m.items.Delete(k)
		atomic.AddInt64(&m.nSize, -1)
		delNum++
		return true
	})
	return prevNSize - atomic.LoadInt64(&m.nSize)
}

// janitor cleans up expired keys from the map.
// Runs every second.
func (m *Map) janitor() {
	for {
		time.Sleep(time.Second)
		m.items.Range(func(k, v interface{}) bool {
			item := v.(*Item)
			if item.HasFlag(ItemFlagExpireXX) && time.Now().After(item.ExpiresAt) {
				m.items.Delete(k)
				atomic.AddInt64(&m.nSize, -1)
			}
			return true
		})
	}
}
