package hashtable

import (
	"hash/fnv"
	"sync"
)

const (
	minLoadFactor = 0.25
	maxLoadFactor = 0.75
	// DefaultSize is the default size of the table
	DefaultSize = 16
)

// HashTable data structure
type HashTable struct {
	table []*Bucket
	nSize int
	m     sync.RWMutex
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

// NewHashTable returns a new Hash Table
func NewHashTable() *HashTable {
	return &HashTable{
		table: make([]*Bucket, DefaultSize),
		nSize: 0,
	}
}

// newHashTable returns a new Hash Table
func newHashTable(tableSize int) *HashTable {
	return &HashTable{
		table: make([]*Bucket, DefaultSize),
		nSize: 0,
	}
}

// Set inserts a new key-value pair item into the hash table
func (ht *HashTable) Set(k string, v string) {
	ht.m.Lock()

	initialSize := ht.nSize
	ht.insert(k, v)
	count := initialSize - ht.nSize
	if count > 0 {
		ht.verifyLoadFactorExpansion()
	}
	ht.m.Unlock()
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

// List returns the table
func (ht *HashTable) List() []*Bucket {
	ht.m.RLock()
	defer ht.m.RUnlock()

	return ht.table
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
	index := ht.hashkey(k, len(ht.table))
	entry := ht.newEntry(k, v)
	if ht.lookup(k) == nil {
		ht.table[index] = &Bucket{}
		entry.Next = ht.table[index].Head
		ht.table[index].Head = entry
	} else {
		iterator := ht.table[index].Head
		for {
			if iterator.Next != nil {
				data := iterator.Next
				if data.Key == k {
					data.Value = v
					break
				}
			} else {
				entry.Next = ht.table[index].Head
				ht.table[index].Head = entry
				break
			}
			iterator = iterator.Next
		}
	}

	ht.nSize++
}

func (ht *HashTable) del(k string) {
	index := ht.hashkey(k, len(ht.table))
	if ht.table[index] == nil {
		return
	}
	if ht.table[index].Head.Key == k {
		ht.table[index].Head = ht.table[index].Head.Next
		ht.nSize--
		return
	}

	iterator := ht.table[index].Head
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
	index := ht.hashkey(k, len(ht.table))
	if ht.table[index] == nil {
		return nil
	}
	iterator := ht.table[index].Head
	for iterator != nil {
		if iterator.Key == k {
			return iterator
		}
		iterator = ht.table[index].Head.Next
	}
	return nil
}

func (ht *HashTable) loadFactor() float64 {
	return float64(ht.nSize) / float64(ht.nSize)
}

func (ht *HashTable) verifyLoadFactorExpansion() {
	if ht.nSize == 0 {
		return
	} else {
		lf := ht.loadFactor()
		if lf > maxLoadFactor {
			newTable := newHashTable(ht.nSize * 2)
			for _, record := range ht.table {
				if record != nil && record.Head != nil {
					newTable.Set(record.Head.Key, record.Head.Value)
					record.Head = record.Head.Next
				}
			}
			ht.table = newTable.table
		} else if lf < minLoadFactor {
			newTable := newHashTable(len(ht.table) / 2)
			for _, record := range ht.table {
				if record != nil && record.Head != nil {
					newTable.Set(record.Head.Key, record.Head.Value)
					record.Head = record.Head.Next
				}
			}
			ht.table = newTable.table
		}
	}
}

func (ht *HashTable) newEntry(key, value string) *Entry {
	return &Entry{
		Key:   key,
		Value: value,
	}
}
func (ht *HashTable) hashkey(k string, size int) int {
	h32 := fnv.New32a()
	h32.Write([]byte(k))
	return int(h32.Sum32()) % size
}
