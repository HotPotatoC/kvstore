package command

import (
	"github.com/HotPotatoC/kvstore/internal/server/stats"
	"github.com/HotPotatoC/kvstore/internal/storage"
)

// Op represents the command operation code
type Op int

const (
	// SET inserts a new entry into the database
	SET Op = iota
	// SETEX inserts a new expirable entry into the database
	SETEX
	// GET returns the data in the database with the matching key
	GET
	// DEL remove an entry in the database with the matching key
	DEL
	// KEYS displays all the saved keys in the database
	KEYS
	// FLUSHALL delete all keys
	FLUSHALL
	// INFO displays the current status of the server (memory allocs, connected clients, uptime, etc.)
	INFO
)

func (c Op) String() string {
	return [...]string{"set", "setex", "get", "del", "keys", "flushall", "info"}[c]
}

func (c Op) Bytes() []byte {
	return [...][]byte{
		[]byte("set"),
		[]byte("setex"),
		[]byte("get"),
		[]byte("del"),
		[]byte("keys"),
		[]byte("flushall"),
		[]byte("info"),
	}[c]
}

// Args is the command operation required arguments
func (c Op) Args() string {
	return [...]string{
		"key value",
		"key value [exp seconds]",
		"key",
		"key",
		"",
		"",
		"",
	}[c]
}

// Description is the command operation description
func (c Op) Description() string {
	return [...]string{
		"Insert a new entry into the database",
		"Insert a new expirable entry into the database",
		"Return the data in the database with the matching key",
		"Remove an entry in the database with the matching key",
		"Display all the saved keys in the database",
		"Delete all keys",
		"Display the current stats of the server (OS, mem usage, total connections, etc.) in json format",
	}[c]
}

// Command is the set of methods for a commmand
type Command interface {
	String() string
	Execute(args []string) []byte
}

// New constructs the given command operation
func New(db storage.Store, stats *stats.Stats, cmd Op) Command {
	var command Command
	switch cmd {
	case SET:
		command = makeSetCommand(db)
	case SETEX:
		command = makeSetEXCommand(db)
	case GET:
		command = makeGetCommand(db)
	case DEL:
		command = makeDelCommand(db)
	case KEYS:
		command = makeKeysCommand(db)
	case FLUSHALL:
		command = makeFlushAllCommand(db)
	case INFO:
		command = makeInfoCommand(db, stats)
	default:
		command = nil
	}

	return command
}
