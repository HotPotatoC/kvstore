package hashtable

import (
	"sync"

	"github.com/cespare/xxhash/v2"
)

const (
	minLoadFactor = 0.25
	maxLoadFactor = 0.75
	// DefaultSize is the default size of the table
	DefaultSize = 16
)

// HashTable data structure
type HashTable struct {
	buckets []*Bucket
	nSize   int
	m       sync.RWMutex
}

// Bucket represents the hash table bucket
type Bucket struct {
	Head *Entry
}

type Entry struct {
	Key   string
	Value string
	Next  *Entry
}

// New returns a new Hash Table
func New() *HashTable {
	return &HashTable{
		buckets: make([]*Bucket, DefaultSize),
		nSize:   0,
	}
}

// newHashTable returns a new Hash Table
func newHashTable(size int) *HashTable {
	return &HashTable{
		buckets: make([]*Bucket, size),
		nSize:   0,
	}
}

// Set inserts a new key-value pair item into the hash table
func (ht *HashTable) Set(k string, v string) {
	ht.m.Lock()
	defer ht.m.Unlock()

	initialSize := ht.nSize
	ht.insert(k, v)
	count := ht.nSize - initialSize
	if count > 0 {
		ht.verifyLoadFactorExpansion()
	}
}

// Get returns the value of the given key
// if the result is empty then returns an empty
// string ("")
func (ht *HashTable) Get(k string) string {
	ht.m.RLock()
	defer ht.m.RUnlock()
	result := ht.lookup(k)
	if result == nil {
		return ""
	}
	return result.Value
}

// Remove deletes an item by the given key
func (ht *HashTable) Remove(k string) int {
	ht.m.Lock()
	defer ht.m.Unlock()
	initialSize := ht.nSize
	ht.del(k)
	count := initialSize - ht.nSize
	if count > 0 {
		ht.verifyLoadFactorExpansion()
	}
	return count
}

// List returns the table
func (ht *HashTable) List() []*Bucket {
	ht.m.RLock()
	defer ht.m.RUnlock()

	return ht.buckets
}

// Exist returns true if an item with the given key exists
// otherwise returns false
func (ht *HashTable) Exist(k string) bool {
	ht.m.RLock()
	defer ht.m.RUnlock()
	return ht.lookup(k) != nil
}

// Size represents the size of the hash table
func (ht *HashTable) Size() int {
	ht.m.RLock()
	defer ht.m.RUnlock()
	return ht.nSize
}

func (ht *HashTable) insert(k string, v string) {
	index := ht.hashkey(k, len(ht.buckets))
	entry := ht.newEntry(k, v)
	if ht.buckets[index] == nil {
		ht.buckets[index] = &Bucket{}
		entry.Next = ht.buckets[index].Head
		ht.buckets[index].Head = entry
	} else {
		iterator := ht.buckets[index].Head
		for {
			if iterator.Next != nil {
				data := iterator.Next
				if data.Key == k {
					data.Value = v
					break
				}
			} else {
				entry.Next = ht.buckets[index].Head
				ht.buckets[index].Head = entry
				break
			}
			iterator = iterator.Next
		}
	}

	ht.nSize++
}

func (ht *HashTable) del(k string) {
	index := ht.hashkey(k, len(ht.buckets))

	if ht.buckets[index] == nil || ht.buckets[index].Head == nil {
		return
	}

	if ht.buckets[index].Head.Key == k {
		ht.buckets[index].Head = ht.buckets[index].Head.Next
		ht.nSize--
		return
	}

	iterator := ht.buckets[index].Head
	for iterator.Next != nil {
		if iterator.Next.Key == k {
			iterator.Next = iterator.Next.Next
			ht.nSize--
			return
		}
		iterator = iterator.Next
	}
}

func (ht *HashTable) lookup(k string) *Entry {
	index := ht.hashkey(k, len(ht.buckets))
	if ht.buckets[index] == nil {
		return nil
	}
	iterator := ht.buckets[index].Head
	for iterator != nil {
		if iterator.Key == k {
			return iterator
		}
		iterator = iterator.Next
	}
	return nil
}

func (ht *HashTable) loadFactor() float64 {
	return float64(ht.nSize) / float64(len(ht.buckets))
}

func (ht *HashTable) verifyLoadFactorExpansion() {
	if ht.nSize == 0 {
		return
	}

	lf := ht.loadFactor()
	if lf > maxLoadFactor {
		newTable := newHashTable(ht.nSize * 2)
		for _, bucket := range ht.buckets {
			for bucket != nil && bucket.Head != nil {
				newTable.insert(bucket.Head.Key, bucket.Head.Value)
				bucket.Head = bucket.Head.Next
			}
		}
		ht.buckets = newTable.buckets
	} else if lf < minLoadFactor {
		newTable := newHashTable(len(ht.buckets) / 2)
		for _, bucket := range ht.buckets {
			for bucket != nil && bucket.Head != nil {
				newTable.insert(bucket.Head.Key, bucket.Head.Value)
				bucket.Head = bucket.Head.Next
			}
		}
		ht.buckets = newTable.buckets
	}
}

func (ht *HashTable) newEntry(key, value string) *Entry {
	return &Entry{
		Key:   key,
		Value: value,
		Next:  nil,
	}
}

func (ht *HashTable) hashkey(key string, size int) uint64 {
	return xxhash.Sum64([]byte(key)) % uint64(size)
}
