package datastructure

import (
	"time"
	"unsafe"
)

// Item represents an item stored in the map.
type Item struct {
	// Key is the key of the item.
	Key string
	// size of the item.
	Size uint32
	// Data stored by the item.
	Data interface{}
	// TTL of the item. (ms)
	TTL time.Duration
	// CreatedAt is the time when the item is created.
	CreatedAt time.Time
}

// NewItem creates a new item.
func NewItem(key string, data interface{}, ttl time.Duration) *Item {
	return &Item{
		Key:       key,
		Size:      uint32(unsafe.Sizeof(data)),
		Data:      data,
		TTL:       ttl,
		CreatedAt: time.Now(),
	}
}