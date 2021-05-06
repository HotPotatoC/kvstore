package database

import (
	"github.com/HotPotatoC/kvstore/pkg/datastructure/hashtable"
)

// Store is the representation of a key-value pair database
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

type database struct {
	table *hashtable.HashTable
}

// New creates a new database
func New() Store {
	return &database{
		table: hashtable.New(),
	}
}

func (d *database) Set(key string, value string) {
	d.table.Set(key, value)
}

func (d *database) SetEX(key string, value string, expiresAfter int) {
	d.table.SetEX(key, value, expiresAfter)
}

func (d *database) Get(key string) string {
	return d.table.Get(key)
}

func (d *database) Remove(key string) int {
	return d.table.Remove(key)
}

func (d *database) Flush() {
	d.table.Flush()
}

func (d *database) Iter() <-chan *hashtable.Entry {
	return d.table.Iter()
}

func (d *database) Exist(key string) bool {
	return d.table.Exist(key)
}

func (d *database) Size() int {
	return d.table.Size()
}
