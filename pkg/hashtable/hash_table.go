package hashtable

import (
	"hash/fnv"
	"sync"
)

// HashTable data structure
type HashTable struct {
	table Dict
	m     sync.RWMutex
}

// Dict represents a map
type Dict map[int]string

// NewHashTable returns a new Hash Table
func NewHashTable() *HashTable {
	return &HashTable{}
}

// Set inserts a new key-value pair item into the hash table
func (ht *HashTable) Set(k string, v string) {
	ht.m.Lock()

	if ht.table == nil {
		ht.init()
	}

	ht.insert(k, v)
	ht.m.Unlock()
}

// Remove deletes an item by the given key
func (ht *HashTable) Remove(k string) int {
	ht.m.Lock()
	defer ht.m.Unlock()
	initialSize := len(ht.table)
	ht.del(k)
	return initialSize - len(ht.table)
}

// Get returns the value of the given key
func (ht *HashTable) Get(k string) string {
	ht.m.RLock()
	defer ht.m.RUnlock()
	return ht.lookup(k)
}

// List returns the table
func (ht *HashTable) List() Dict {
	ht.m.RLock()
	defer ht.m.RUnlock()

	return ht.table
}

// Exist returns true if an item with the given key exists
// otherwise returns false
func (ht *HashTable) Exist(k string) bool {
	ht.m.RLock()
	defer ht.m.RUnlock()
	return ht.lookup(k) != ""
}

// Size represents the size of the hash table
func (ht *HashTable) Size() int {
	ht.m.RLock()
	defer ht.m.RUnlock()
	return len(ht.table)
}

func (ht *HashTable) init() {
	ht.table = make(map[int]string)
}

func (ht *HashTable) insert(k string, v string) {
	ht.table[ht.hashkey(k)] = v
}

func (ht *HashTable) del(k string) {
	if k == "*" {
		for key := range ht.table {
			delete(ht.table, key)
		}
		return
	}
	delete(ht.table, ht.hashkey(k))
}

func (ht *HashTable) lookup(k string) string {
	return ht.table[ht.hashkey(k)]
}

func (ht *HashTable) hashkey(k string) int {
	h32 := fnv.New32a()
	h32.Write([]byte(k))
	return int(h32.Sum32())
}
