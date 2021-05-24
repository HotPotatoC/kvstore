package storage

import (
	"github.com/HotPotatoC/kvstore/internal/datastructure/hashtable"
)

// Store is the representation of a key-value pair storage
type Store interface {
	// Set inserts a key-value data
	Set(key string, value string)

	// SetEX inserts an expirable key-value data
	SetEX(key string, value string, expiresAfter int)

	// Get returns the value of the given key
	Get(key string) string

	// Remove deletes an item by the given key
	Remove(key string) int

	// Flush deletes all keys
	Flush()

	// Iter represents an iterator for the table
	Iter() <-chan *hashtable.Entry

	// Exist returns true if an item with the given key exists otherwise returns false
	Exist(key string) bool

	// Size returns the amount of entries in the table
	Size() int
}

type storage struct {
	hashtable *hashtable.HashTable
}

// New creates a new storage for data structures
func New() Store {
	return &storage{
		hashtable: hashtable.New(),
	}
}

func (d *storage) Set(key string, value string) {
	d.hashtable.Set(key, value)
}

func (d *storage) SetEX(key string, value string, expiresAfter int) {
	d.hashtable.SetEX(key, value, expiresAfter)
}

func (d *storage) Get(key string) string {
	return d.hashtable.Get(key)
}

func (d *storage) Remove(key string) int {
	return d.hashtable.Remove(key)
}

func (d *storage) Flush() {
	d.hashtable.Flush()
}

func (d *storage) Iter() <-chan *hashtable.Entry {
	return d.hashtable.Iter()
}

func (d *storage) Exist(key string) bool {
	return d.hashtable.Exist(key)
}

func (d *storage) Size() int {
	return d.hashtable.Size()
}
