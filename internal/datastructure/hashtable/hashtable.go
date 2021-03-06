package hashtable

import (
	"sync"
	"time"

	"github.com/HotPotatoC/kvstore/internal/logger"
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
	mtx     sync.RWMutex
}

// Bucket represents the hash table bucket
type Bucket struct {
	Head *Entry
}

// Entry represents an entry inside the bucket
type Entry struct {
	Key          string
	Value        string
	CreatedAt    time.Time
	ShouldExpire bool
	ExpireAfter  int
	Next         *Entry
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
	ht.mtx.Lock()
	defer ht.mtx.Unlock()

	initialSize := ht.nSize
	ht.insert(k, v, 0)
	count := ht.nSize - initialSize
	if count > 0 {
		ht.verifyLoadFactor()
	}
}

// SetEX inserts a new expirable key-value pair item into the hash table
func (ht *HashTable) SetEX(k string, v string, expiresAfter int) {
	ht.mtx.Lock()
	initialSize := ht.nSize
	ht.insert(k, v, expiresAfter)
	count := ht.nSize - initialSize
	if count > 0 {
		ht.verifyLoadFactor()
	}
	ht.mtx.Unlock()

	time.AfterFunc(time.Duration(expiresAfter)*time.Second, func() {
		ht.mtx.Lock()
		if entry := ht.lookup(k); entry.ShouldExpire {
			initialSize := ht.nSize
			ht.delete(k)
			if count := initialSize - ht.nSize; count > 0 {
				ht.verifyLoadFactor()
			}
		}
		ht.mtx.Unlock()
	})
}

// Get returns the value of the given key
// if the result is empty then returns an empty
// string ("")
func (ht *HashTable) Get(k string) string {
	ht.mtx.RLock()
	defer ht.mtx.RUnlock()
	result := ht.lookup(k)
	if result == nil {
		return ""
	}

	if result.ShouldExpire && time.Since(result.CreatedAt) > time.Duration(result.ExpireAfter)*time.Second {
		return ""
	}

	return result.Value
}

// Remove deletes an item by the given key
func (ht *HashTable) Remove(k string) int {
	ht.mtx.Lock()
	defer ht.mtx.Unlock()
	initialSize := ht.nSize
	ht.delete(k)
	count := initialSize - ht.nSize
	if count > 0 {
		ht.verifyLoadFactor()
	}
	return count
}

// Flush clears the bucket
func (ht *HashTable) Flush() {
	ht.mtx.Lock()
	defer ht.mtx.Unlock()
	ht.buckets = make([]*Bucket, DefaultSize)
	ht.nSize = 0
	ht.verifyLoadFactor()
}

// Iter represents an iterator for the hashtable
func (ht *HashTable) Iter() <-chan *Entry {
	ch := make(chan *Entry)
	go func() {
		ht.mtx.RLock()
		ht.iterate(ch)
		ht.mtx.RUnlock()
	}()
	return ch
}

// Exist returns true if an item with the given key exists
// otherwise returns false
func (ht *HashTable) Exist(k string) bool {
	ht.mtx.RLock()
	defer ht.mtx.RUnlock()
	return ht.lookup(k) != nil
}

// Size represents the size of the hash table
func (ht *HashTable) Size() int {
	ht.mtx.RLock()
	defer ht.mtx.RUnlock()
	return ht.nSize
}

func (ht *HashTable) insert(k string, v string, expiresAfter int) {
	logger.S().Debug("attempting to insert an entry to the hash table...")
	var entry *Entry
	index := ht.hashkey(k, len(ht.buckets))
	logger.S().Debugf("generated index: %d", index)

	if expiresAfter > 0 {
		logger.S().Debug("the given entry is expirable")
		logger.S().Debug("creating an expirable entry for the hash table...")
		entry = ht.newExpirableEntry(k, v, expiresAfter)
	} else {
		logger.S().Debug("creating an entry for the hash table...")
		entry = ht.newEntry(k, v)
	}

	if ht.buckets[index] == nil || ht.buckets[index].Head == nil {
		logger.S().Debug("bucket is empty creating one...")
		ht.buckets[index] = &Bucket{}
		entry.Next = ht.buckets[index].Head
		ht.buckets[index].Head = entry
		ht.nSize++
		logger.S().Debug("insert success")
		logger.S().Debugf("new hash table size: %d", ht.nSize)
		return
	}

	for iterator := ht.buckets[index].Head; iterator != nil; iterator = iterator.Next {
		if iterator.Next == nil {
			if iterator.Key == k {
				iterator.Value = v
				return
			}
			entry.Next = ht.buckets[index].Head
			ht.buckets[index].Head = entry
			break
		}

		if iterator.Next.Key == k {
			iterator.Next.Value = v
			break
		}
	}

	ht.nSize++
	logger.S().Debug("insert success")
	logger.S().Debugf("new hash table size: %d", ht.nSize)
}

func (ht *HashTable) delete(k string) {
	logger.S().Debugf("attempting to delete an entry with the key: %s", k)
	index := ht.hashkey(k, len(ht.buckets))
	logger.S().Debugf("generated index: %d", index)

	if ht.buckets[index] == nil || ht.buckets[index].Head == nil {
		logger.S().Debug("entry to delete was not found")
		logger.S().Debug("cancelling deletion")
		return
	}

	if ht.buckets[index].Head.Key == k {
		ht.buckets[index].Head = ht.buckets[index].Head.Next
		ht.nSize--
		logger.S().Debug("delete success")
		logger.S().Debugf("new hash table size: %d", ht.nSize)
		return
	}

	iterator := ht.buckets[index].Head
	for iterator.Next != nil {
		if iterator.Next.Key == k {
			iterator.Next = iterator.Next.Next
			ht.nSize--
			logger.S().Debug("delete success")
			logger.S().Debugf("new hash table size: %d", ht.nSize)
			return
		}
		iterator = iterator.Next
	}
}

func (ht *HashTable) lookup(k string) *Entry {
	logger.S().Debugf("finding an entry with the key: %s", k)
	index := ht.hashkey(k, len(ht.buckets))
	logger.S().Debugf("generated index: %d", index)
	if ht.buckets[index] == nil {
		logger.S().Debug("lookup process did not found any entry")
		logger.S().Debug("returning nil")
		return nil
	}

	iterator := ht.buckets[index].Head
	for iterator != nil {
		if iterator.Key == k {
			logger.S().Debug("lookup success")
			logger.S().Debug("returning the entry")
			return iterator
		}
		iterator = iterator.Next
	}

	logger.S().Debug("lookup process did not found any entry")
	logger.S().Debug("returning nil")
	return nil
}

func (ht *HashTable) iterate(ch chan<- *Entry) {
	for _, bucket := range ht.buckets {
		if bucket != nil {
			for entry := bucket.Head; entry != nil; entry = entry.Next {
				ch <- entry
			}
		}
	}
	close(ch)
}

func (ht *HashTable) loadFactor() float32 {
	return float32(ht.nSize) / float32(len(ht.buckets))
}

func (ht *HashTable) verifyLoadFactor() {
	logger.S().Debug("verifying hash table load factor to increase/decrease the amount of buckets")
	if ht.nSize == 0 {
		logger.S().Debug("hash table is empty")
		return
	}

	lf := ht.loadFactor()
	if lf > maxLoadFactor {
		logger.S().Debugf("hash table load factor exceeds the maximum load factor (%f)", lf)
		logger.S().Debugf("increasing hash table capacity %d -> %d", ht.nSize, ht.nSize*2)
		ht.updateCapacity(ht.nSize * 2)
	} else if lf < minLoadFactor {
		logger.S().Debugf("hash table load factor below the minimum load factor (%f)", lf)
		logger.S().Debugf("decreasing hash table capacity %d -> %d", ht.nSize, len(ht.buckets)/2)
		ht.updateCapacity(len(ht.buckets) / 2)
	}
	logger.S().Debug("hash table capacity didn't change")
}

func (ht *HashTable) updateCapacity(size int) {
	newTable := newHashTable(size)
	for _, bucket := range ht.buckets {
		for bucket != nil && bucket.Head != nil {
			newTable.insert(bucket.Head.Key, bucket.Head.Value, int(bucket.Head.ExpireAfter))
			bucket.Head = bucket.Head.Next
		}
	}
	ht.buckets = newTable.buckets
}

func (ht *HashTable) newEntry(key, value string) *Entry {
	return &Entry{
		Key:          key,
		Value:        value,
		CreatedAt:    time.Now(),
		ShouldExpire: false,
		ExpireAfter:  0,
		Next:         nil,
	}
}

func (ht *HashTable) newExpirableEntry(key, value string, expiresAfter int) *Entry {
	return &Entry{
		Key:          key,
		Value:        value,
		CreatedAt:    time.Now(),
		ShouldExpire: true,
		ExpireAfter:  expiresAfter,
		Next:         nil,
	}
}

func (ht *HashTable) hashkey(key string, size int) uint64 {
	return xxhash.Sum64([]byte(key)) % uint64(size)
}
