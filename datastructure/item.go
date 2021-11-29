package datastructure

import (
	"time"
	"unsafe"
)

type ItemFlag uint32

const (
	// ItemFlagNone is the default flag.
	ItemFlagNone ItemFlag = 0
	// ItemExpireNX indicates that the item has no expiry.
	ItemFlagExpireNX ItemFlag = 1 << iota
	// ItemExpireXX indicates that the item has an expiry.
	ItemFlagExpireXX
)

// Item represents an item stored in the map.
type Item struct {
	// Key is the key of the item.
	Key string
	// size of the item.
	Size uint32
	// Data stored by the item.
	Data interface{}
	// Flag is a bitmask of item options.
	Flag ItemFlag
	// ExpiresAt is the time when the item expires.
	ExpiresAt time.Time
	// CreatedAt is the time when the item is created.
	CreatedAt time.Time
}

// NewItem creates a new item.
func NewItem(key string, data interface{}, ttl time.Duration) *Item {
	item := &Item{
		Key:       key,
		Size:      uint32(unsafe.Sizeof(data)),
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	if ttl == 0 {
		item.Flag |= ItemFlagExpireNX
	} else if ttl > 0 {
		item.Flag |= ItemFlagExpireXX
	}

	return item
}

// HasFlag returns true if the item has the given flag.
func (i *Item) HasFlag(flag ItemFlag) bool {
	return i.Flag&flag != 0
}

// AddFlag adds the given flag to the item.
func (i *Item) AddFlag(flag ItemFlag) {
	i.Flag |= flag
}

// RemoveFlag removes the given flag from the item.
func (i *Item) RemoveFlag(flag ItemFlag) {
	i.Flag &= ^flag
}
