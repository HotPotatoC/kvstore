package datastructure

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Map is a thread-safe map.
type Map struct {
	Items sync.Map
	nSize int64
}

// NewMap returns a new Map.
func NewMap() *Map {
	return &Map{}
}

// Store stores a new key-value pair.
func (m *Map) Store(v *Item) {
	m.Items.Store(v.Key, v)
	atomic.AddInt64(&m.nSize, 1)
	if v.TTL > 0 {
		time.AfterFunc(v.TTL, func() {
			m.Items.Delete(v.Key)
			atomic.AddInt64(&m.nSize, -1)
		})
	}
}

// Expire sets the expiration time of the key.
// TODO: reset the timer of a key that is already has an expiry
func (m *Map) Expire(k string, ttl time.Duration) int64 {
	v, ok := m.Items.Load(k)
	if !ok {
		return 0
	}

	v.(*Item).TTL = ttl
	v.(*Item).Flag |= ItemFlagExpireXX
	if ttl > 0 {
		time.AfterFunc(ttl, func() {
			m.Items.Delete(k)
			atomic.AddInt64(&m.nSize, -1)
		})
	}

	return atomic.LoadInt64(&m.nSize)
}

// Get returns the value of the key.
func (m *Map) Get(k string) (*Item, bool) {
	v, ok := m.Items.Load(k)
	if !ok {
		return nil, false
	}

	if v.(*Item).Flag&ItemFlagExpireXX != 0 && time.Since(v.(*Item).CreatedAt) > v.(*Item).TTL {
		m.Items.Delete(k)
		return nil, false
	}

	return v.(*Item), true
}

// Delete deletes the key.
func (m *Map) Delete(k string) int64 {
	if _, ok := m.Items.Load(k); !ok {
		return 0
	}
	m.Items.Delete(k)

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

	m.Items.Range(func(key, value interface{}) bool {
		items[key.(string)] = value.(*Item)
		return true
	})

	return items
}

// Keys returns the keys of the map.
func (m *Map) Keys() []string {
	var keys []string
	m.Items.Range(func(k, _ interface{}) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

// KeysWithPattern returns the keys of the map that match the pattern.
func (m *Map) KeysWithPattern(pattern string) []string {
	var keys []string
	m.Items.Range(func(k, _ interface{}) bool {
		if strings.Contains(k.(string), pattern) {
			keys = append(keys, k.(string))
		}
		return true
	})
	return keys
}

// Values returns the values of the map.
func (m *Map) Values() []*Item {
	values := make([]*Item, 0, atomic.LoadInt64(&m.nSize))
	m.Items.Range(func(k, v interface{}) bool {
		values = append(values, v.(*Item))
		return true
	})
	return values
}

// Exists checks if the key exists in the map.
func (m *Map) Exists(k string) bool {
	_, ok := m.Items.Load(k)
	return ok
}

// Clear clears the map.
func (m *Map) Clear() int64 {
	prevNSize := atomic.LoadInt64(&m.nSize)
	var delNum int64
	m.Items.Range(func(k, v interface{}) bool {
		m.Items.Delete(k)
		atomic.AddInt64(&m.nSize, -1)
		delNum++
		return true
	})
	return prevNSize - atomic.LoadInt64(&m.nSize)
}
