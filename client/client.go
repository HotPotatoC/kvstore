package client

import (
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/panjf2000/gnet"
)

// Flags is a bitmask of client options.
type Flags uint32

const (
	// FlagNone indicates no flags.
	FlagNone Flags = 1 << iota
	// FlagReadOnly is a client option to set read-only behavior.
	FlagReadOnly
	// FlagBusy is a client option to set busy behavior.
	FlagBusy
	// FlagCloseASAP this client option will close the connection as soon as
	// the server replies.
	FlagCloseASAP
)

func (f Flags) String() string {
	var s string
	if f&FlagNone != 0 {
		return "N"
	}
	if f&FlagReadOnly != 0 {
		s += "r"
	}
	if f&FlagBusy != 0 {
		s += "b"
	}
	if f&FlagCloseASAP != 0 {
		s += "c"
	}
	return s
}

// Client is a client that connects to a server.
type Client struct {
	// ID is the incremental ID of the client.
	ID int64
	// Name is the name of the client.
	Name string
	// Flags is a bitmask of client options.
	Flags Flags
	// Conn is the underlying connection.
	Conn gnet.Conn
	// DB is the database client.
	DB *datastructure.Map
	// kvsDB is the file used to persist the data structure.
	KVSDB *disk.KVSDB
	// Command is the command that the client is currently executing.
	Command string
	// Argc is the number of arguments excluding the command.
	Argc int
	// Argv is the arguments excluding the command.
	Argv [][]byte
	// CreateTime is the time when the client is created.
	CreateTime time.Time
}

// HasFlag returns true if the client has the specified flag.
func (c *Client) HasFlag(flag Flags) bool {
	return c.Flags&flag != 0
}

// AddFlag adds the specified flag to the client.
func (c *Client) AddFlag(flag Flags) {
	c.Flags |= flag
}

// RemoveFlag removes the specified flag from the client.
func (c *Client) RemoveFlag(flag Flags) {
	c.Flags &= ^flag
}