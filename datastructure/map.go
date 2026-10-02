package datastructure

import (
	"errors"
	"hash/fnv"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const numShards = 256

// Item overhead estimates the item and map/TTL index entries. The budget does
// not include Go allocator slack, shard buckets, or transient network buffers.
const itemOverhead = int64(unsafe.Sizeof(Item{})) + 64

var ErrMaxMemory = errors.New("maxmemory limit reached")

func itemBytes(key, value string) int64 {
	return int64(len(key)) + int64(len(value)) + itemOverhead
}

// shard is a single partition of the sharded map.
type shard struct {
	mu      sync.RWMutex
	items   map[string]*Item
	ttlKeys map[string]struct{} // keys that have an expiry set
	bytes   int64               // accounted bytes, protected by mu
}

// Map is a thread-safe sharded map.
type Map struct {
	shards     [numShards]shard
	nSize      atomic.Int64
	usedMemory atomic.Int64
	maxMemory  atomic.Int64
	done       chan struct{}
	closeOnce  sync.Once
}

// NewMap returns a new Map.
func NewMap() *Map {
	m := &Map{done: make(chan struct{})}
	for i := range m.shards {
		m.shards[i].items = make(map[string]*Item)
		m.shards[i].ttlKeys = make(map[string]struct{})
	}
	go m.janitor()
	return m
}

// Close stops background expiry cleanup. It is safe to call repeatedly.
func (m *Map) Close() {
	m.closeOnce.Do(func() {
		if m.done != nil {
			close(m.done)
		}
	})
}

// shardIndex returns the shard index for the given key.
func shardIndex(k string) uint32 {
	return shardIndexBytes([]byte(k))
}

func shardIndexBytes(k []byte) uint32 {
	h := fnv.New32a()
	h.Write(k)
	return h.Sum32() % numShards
}

// SetMaxMemory sets the stored-data budget in bytes; zero means unlimited.
// Configure before accepting writes. Existing data is never evicted.
func (m *Map) SetMaxMemory(n int64) { m.maxMemory.Store(n) }

// UsedMemory returns accounted key/value and estimated metadata bytes.
func (m *Map) UsedMemory() int64 { return m.usedMemory.Load() }

// MaxMemory returns the configured stored-data budget; zero means unlimited.
func (m *Map) MaxMemory() int64 { return m.maxMemory.Load() }

// KeyspaceStats counts stored keys and expiry entries without visiting items.
// Expired keys remain counted until lazy or background cleanup removes them.
func (m *Map) KeyspaceStats() (keys, expires int64) {
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		keys += int64(len(s.items))
		expires += int64(len(s.ttlKeys))
		s.mu.RUnlock()
	}
	return
}

// reserveMemory atomically admits growth across all shards.
func (m *Map) reserveMemory(delta int64) bool {
	if delta == 0 {
		return true
	}
	if delta < 0 {
		m.usedMemory.Add(delta)
		return true
	}
	for {
		used := m.usedMemory.Load()
		limit := m.maxMemory.Load()
		if limit > 0 && (used > limit || delta > limit-used) {
			return false
		}
		if m.usedMemory.CompareAndSwap(used, used+delta) {
			return true
		}
	}
}

// Store stores a new key-value pair if it fits the configured budget.
func (m *Map) Store(v *Item) { _, _ = m.StoreLimited(v) }

// StoreLimited returns an error if the item would exceed the memory budget.
func (m *Map) StoreLimited(v *Item) (bool, error) { return m.store(v, false, false) }

// StoreIfAbsent stores the item only if its key is absent or expired.
func (m *Map) StoreIfAbsent(v *Item) bool {
	stored, _ := m.store(v, true, false)
	return stored
}

// StoreIfPresent replaces the item only if its key exists and is not expired.
func (m *Map) StoreIfPresent(v *Item) bool {
	stored, _ := m.store(v, false, true)
	return stored
}

func (m *Map) store(v *Item, onlyAbsent, onlyPresent bool) (bool, error) {
	s := &m.shards[shardIndex(v.Key)]
	s.mu.Lock()
	defer s.mu.Unlock()
	previous, exists := m.liveItemLocked(s, v.Key)
	if (onlyAbsent && exists) || (onlyPresent && !exists) {
		return false, nil
	}
	delta := itemBytes(v.Key, v.Data)
	if exists {
		delta -= itemBytes(previous.Key, previous.Data)
	}
	if !m.reserveMemory(delta) {
		return false, ErrMaxMemory
	}
	m.storeLocked(s, v, exists)
	s.bytes += delta
	return true, nil
}

// StoreBytes copies transient command data into a new immutable item.
func (m *Map) StoreBytes(key, value []byte, ttl time.Duration, onlyAbsent, onlyPresent bool) bool {
	stored, _ := m.StoreBytesLimited(key, value, ttl, onlyAbsent, onlyPresent)
	return stored
}

// StoreBytesLimited checks conditions and reserves the memory budget before
// copying request data. Overwrites reuse the immutable stored key.
func (m *Map) StoreBytesLimited(key, value []byte, ttl time.Duration, onlyAbsent, onlyPresent bool) (bool, error) {
	s := &m.shards[shardIndexBytes(key)]
	s.mu.Lock()
	defer s.mu.Unlock()
	previous, exists := m.liveItemLocked(s, string(key))
	if (onlyAbsent && exists) || (onlyPresent && !exists) {
		return false, nil
	}
	delta := int64(len(key)) + int64(len(value)) + itemOverhead
	if exists {
		delta -= itemBytes(previous.Key, previous.Data)
	}
	if !m.reserveMemory(delta) {
		return false, ErrMaxMemory
	}
	v := NewItem("", string(value), ttl)
	if exists {
		v.Key = previous.Key
	} else {
		v.Key = string(key)
	}
	m.storeLocked(s, v, exists)
	s.bytes += delta
	return true, nil
}

// removeLocked releases all accounting for an existing key. Caller holds s.mu.
func (m *Map) removeLocked(s *shard, key string, item *Item) {
	delete(s.items, key)
	delete(s.ttlKeys, key)
	n := itemBytes(item.Key, item.Data)
	s.bytes -= n
	m.usedMemory.Add(-n)
	m.nSize.Add(-1)
}

// liveItemLocked treats expired keys as absent. The caller holds s.mu.
func (m *Map) liveItemLocked(s *shard, key string) (*Item, bool) {
	previous, exists := s.items[key]
	if exists && previous.HasFlag(ItemFlagExpireXX) && !time.Now().Before(previous.ExpiresAt) {
		m.removeLocked(s, key, previous)
		return nil, false
	}
	return previous, exists
}

// storeLocked publishes an owned item. The caller holds s.mu.
func (m *Map) storeLocked(s *shard, v *Item, exists bool) {
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
}

// Expire sets the expiration time of the key.
func (m *Map) Expire(k string, ttl time.Duration) int64 {
	s := &m.shards[shardIndex(k)]
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := m.liveItemLocked(s, k)
	if !ok {
		return 0
	}
	if ttl <= 0 {
		m.removeLocked(s, k, item)
		return 1
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

	return 1
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
	if !item.HasFlag(ItemFlagExpireXX) || time.Now().Before(item.ExpiresAt) {
		return item, true
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok = s.items[k]
	if !ok {
		return nil, false
	}
	if item.HasFlag(ItemFlagExpireXX) && !time.Now().Before(item.ExpiresAt) {
		m.removeLocked(s, k, item)
		return nil, false
	}
	return item, true
}

// Delete deletes one literal key. Expired keys count as absent.
func (m *Map) Delete(k string) int64 {
	s := &m.shards[shardIndex(k)]
	s.mu.Lock()
	defer s.mu.Unlock()
	item, exists := m.liveItemLocked(s, k)
	if !exists {
		return 0
	}
	m.removeLocked(s, k, item)
	return 1
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
	keys, _ := m.keys("", true, 0)
	return keys
}

// KeysWithPattern returns the keys of the map that match the pattern.
func (m *Map) KeysWithPattern(pattern string) []string {
	keys, _ := m.keys(pattern, false, 0)
	return keys
}

// KeysWithPatternLimit stops before the response estimate exceeds maxBytes.
// The estimate reserves 32 bytes for the array and len(key)+32 per key, matching
// the server's existing output limit. Non-positive limits allow all matches.
func (m *Map) KeysWithPatternLimit(pattern string, maxBytes int) ([]string, bool) {
	return m.keys(pattern, pattern == "*", maxBytes)
}

func (m *Map) keys(pattern string, all bool, maxBytes int) ([]string, bool) {
	var keys []string
	remaining := maxBytes - 32
	now := time.Now()
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		for k, item := range s.items {
			if !item.HasFlag(ItemFlagExpireNX) && !now.Before(item.ExpiresAt) {
				continue
			}
			if !all {
				if match, _ := filepath.Match(pattern, k); !match {
					continue
				}
			}
			if maxBytes > 0 {
				if len(k) > remaining-32 {
					s.mu.RUnlock()
					return nil, true
				}
				remaining -= len(k) + 32
			}
			keys = append(keys, k)
		}
		s.mu.RUnlock()
	}
	return keys, false
}

// Exists checks if the key exists in the map.
func (m *Map) Exists(k string) bool {
	_, ok := m.Get(k)
	return ok
}

// Clear clears the map.
func (m *Map) Clear() int64 {
	var clearedN int64
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.Lock()
		cleared := int64(len(s.items))
		m.usedMemory.Add(-s.bytes)
		s.bytes = 0
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
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
		}
		now := time.Now()
		for i := range m.shards {
			s := &m.shards[i]

			// First pass: find expired keys (RLock)
			s.mu.RLock()
			var expired []string
			for k := range s.ttlKeys {
				item, ok := s.items[k]
				if ok && item.HasFlag(ItemFlagExpireXX) && !now.Before(item.ExpiresAt) {
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
				if ok && item.HasFlag(ItemFlagExpireXX) && !now.Before(item.ExpiresAt) {
					m.removeLocked(s, k, item)
				}
			}
			s.mu.Unlock()
		}
	}
}
