package datastructure

import (
	"hash/fnv"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const numShards = 256

// shard is a single partition of the sharded map.
type shard struct {
	mu      sync.RWMutex
	items   map[string]*Item
	ttlKeys map[string]struct{} // keys that have an expiry set
}

// Map is a thread-safe sharded map.
type Map struct {
	shards [numShards]shard
	nSize  atomic.Int64
}

// NewMap returns a new Map.
func NewMap() *Map {
	m := &Map{}
	for i := range m.shards {
		m.shards[i].items = make(map[string]*Item)
		m.shards[i].ttlKeys = make(map[string]struct{})
	}
	go m.janitor()
	return m
}

// shardIndex returns the shard index for the given key.
func shardIndex(k string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(k))
	return h.Sum32() % numShards
}

// Store stores a new key-value pair.
func (m *Map) Store(v *Item) {
	m.store(v, false, false)
}

// StoreIfAbsent stores the item only if its key is absent or expired.
func (m *Map) StoreIfAbsent(v *Item) bool {
	return m.store(v, true, false)
}

// StoreIfPresent replaces the item only if its key exists and is not expired.
func (m *Map) StoreIfPresent(v *Item) bool {
	return m.store(v, false, true)
}

func (m *Map) store(v *Item, onlyAbsent, onlyPresent bool) bool {
	s := &m.shards[shardIndex(v.Key)]
	s.mu.Lock()
	defer s.mu.Unlock()
	previous, exists := s.items[v.Key]
	if exists && previous.HasFlag(ItemFlagExpireXX) && time.Now().After(previous.ExpiresAt) {
		delete(s.items, v.Key)
		delete(s.ttlKeys, v.Key)
		m.nSize.Add(-1)
		exists = false
	}
	if (onlyAbsent && exists) || (onlyPresent && !exists) {
		return false
	}
	if s.items == nil {
		s.items = make(map[string]*Item)
	}
	if !exists {
		m.nSize.Add(1)
	}
	s.items[v.Key] = v
	if v.HasFlag(ItemFlagExpireXX) {
		if s.ttlKeys == nil {
			s.ttlKeys = make(map[string]struct{})
		}
		s.ttlKeys[v.Key] = struct{}{}
	} else {
		delete(s.ttlKeys, v.Key)
	}
	return true
}

// Expire sets the expiration time of the key.
func (m *Map) Expire(k string, ttl time.Duration) int64 {
	s := &m.shards[shardIndex(k)]
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[k]
	if !ok {
		return 0
	}

	updated := *item
	updated.RemoveFlag(ItemFlagExpireNX)
	updated.AddFlag(ItemFlagExpireXX)
	updated.ExpiresAt = time.Now().Add(ttl)
	s.items[k] = &updated
	if s.ttlKeys == nil {
		s.ttlKeys = make(map[string]struct{})
	}
	s.ttlKeys[k] = struct{}{}

	return m.nSize.Load()
}

// Get returns the value of the key.
func (m *Map) Get(k string) (*Item, bool) {
	s := &m.shards[shardIndex(k)]
	s.mu.RLock()
	item, ok := s.items[k]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !item.HasFlag(ItemFlagExpireXX) || !time.Now().After(item.ExpiresAt) {
		return item, true
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok = s.items[k]
	if !ok {
		return nil, false
	}
	if item.HasFlag(ItemFlagExpireXX) && time.Now().After(item.ExpiresAt) {
		delete(s.items, k)
		delete(s.ttlKeys, k)
		m.nSize.Add(-1)
		return nil, false
	}
	return item, true
}

// Delete deletes the key.
func (m *Map) Delete(k string) int64 {
	if k == "*" {
		return m.Clear()
	}

	// Literal keys take precedence over glob patterns.
	s := &m.shards[shardIndex(k)]
	s.mu.Lock()
	if _, exists := s.items[k]; exists {
		delete(s.items, k)
		delete(s.ttlKeys, k)
		m.nSize.Add(-1)
		s.mu.Unlock()
		return 1
	}
	s.mu.Unlock()

	if !strings.ContainsAny(k, "*?[\\") {
		return 0
	}
	if _, err := filepath.Match(k, ""); err != nil {
		return 0
	}

	var deletedN int64
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.Lock()
		var shardDeleted int64
		for key := range s.items {
			if match, _ := filepath.Match(k, key); match {
				delete(s.items, key)
				delete(s.ttlKeys, key)
				shardDeleted++
			}
		}
		m.nSize.Add(-shardDeleted)
		deletedN += shardDeleted
		s.mu.Unlock()
	}
	return deletedN
}

// Len returns the number of items in the map.
func (m *Map) Len() int64 {
	return m.nSize.Load()
}

// List returns all keys and values in a map.
func (m *Map) List() map[string]*Item {
	items := make(map[string]*Item)
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		for k, v := range s.items {
			items[k] = v
		}
		s.mu.RUnlock()
	}
	return items
}

// Keys returns the keys of the map.
func (m *Map) Keys() []string {
	var keys []string
	now := time.Now()
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		for k, item := range s.items {
			if item.HasFlag(ItemFlagExpireNX) || now.Before(item.ExpiresAt) {
				keys = append(keys, k)
			}
		}
		s.mu.RUnlock()
	}
	return keys
}

// KeysWithPattern returns the keys of the map that match the pattern.
func (m *Map) KeysWithPattern(pattern string) []string {
	var keys []string
	now := time.Now()
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		for k, item := range s.items {
			if match, _ := filepath.Match(pattern, k); match && (item.HasFlag(ItemFlagExpireNX) || now.Before(item.ExpiresAt)) {
				keys = append(keys, k)
			}
		}
		s.mu.RUnlock()
	}
	return keys
}

// Exists checks if the key exists in the map.
func (m *Map) Exists(k string) bool {
	idx := shardIndex(k)
	s := &m.shards[idx]
	s.mu.RLock()
	_, ok := s.items[k]
	s.mu.RUnlock()
	return ok
}

// Clear clears the map.
func (m *Map) Clear() int64 {
	var clearedN int64
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.Lock()
		cleared := int64(len(s.items))
		s.items = make(map[string]*Item)
		s.ttlKeys = make(map[string]struct{})
		m.nSize.Add(-cleared)
		clearedN += cleared
		s.mu.Unlock()
	}
	return clearedN
}

// janitor cleans up expired keys from the map.
// Runs every second, only scanning keys with TTL.
func (m *Map) janitor() {
	for {
		time.Sleep(time.Second)
		now := time.Now()
		for i := range m.shards {
			s := &m.shards[i]

			// First pass: find expired keys (RLock)
			s.mu.RLock()
			var expired []string
			for k := range s.ttlKeys {
				item, ok := s.items[k]
				if ok && item.HasFlag(ItemFlagExpireXX) && now.After(item.ExpiresAt) {
					expired = append(expired, k)
				}
			}
			s.mu.RUnlock()

			if len(expired) == 0 {
				continue
			}

			// Second pass: recheck expiry before deleting (Lock)
			s.mu.Lock()
			for _, k := range expired {
				item, ok := s.items[k]
				if ok && item.HasFlag(ItemFlagExpireXX) && now.After(item.ExpiresAt) {
					delete(s.items, k)
					delete(s.ttlKeys, k)
					m.nSize.Add(-1)
				}
			}
			s.mu.Unlock()
		}
	}
}
