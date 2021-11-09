package datastructure

import (
	"time"
	"unsafe"
)

// Item represents an item stored in the map.
type Item struct {
	// Key is the key of the item.
	Key string
	// Type of the item.
	Type ValueType
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
func NewItem(key string, data interface{}, vt ValueType, ttl time.Duration) *Item {
	return &Item{
		Key:       key,
		Type:      vt,
		Size:      uint32(unsafe.Sizeof(data)),
		Data:      data,
		TTL:       ttl,
		CreatedAt: time.Now(),
	}
}

// ValueType is the type of the value.
type ValueType uint32

const (
	// TypeString is the type of string.
	TypeString ValueType = 0x01
	// TypeList is the type of list.
	TypeList ValueType = 0x02
	// TypeCounter is the type of counter.
	TypeCounter ValueType = 0x03
)
