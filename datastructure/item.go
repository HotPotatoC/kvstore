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
	flag := ItemFlagNone
	if ttl == 0 {
		flag |= ItemFlagExpireNX
	} else if ttl > 0 {
		flag |= ItemFlagExpireXX
	}

	return &Item{
		Key:       key,
		Size:      uint32(unsafe.Sizeof(data)),
		Data:      data,
		Flag:      flag,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
}
