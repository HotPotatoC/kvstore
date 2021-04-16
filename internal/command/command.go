package command

import (
	"github.com/HotPotatoC/kvstore/internal/database"
	"github.com/HotPotatoC/kvstore/internal/server/stats"
)

// Op represents the command type
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
	// LIST displays all the saved data in the database
	LIST
	// KEYS displays all the saved keys in the database
	KEYS
	// FLUSH delete all keys
	FLUSH
	// INFO displays the current status of the server (memory allocs, connected clients, uptime, etc.)
	INFO
)

func (c Op) String() string {
	return [...]string{"set", "setex", "get", "del", "list", "keys", "flush", "info"}[c]
}

// Args is the command required args. Used for information
func (c Op) Args() string {
	return [...]string{
		"key value",
		"key value [exp seconds]",
		"key",
		"key",
		"",
		"",
		"",
		"",
	}[c]
}

// Description is the command description. Used for information
func (c Op) Description() string {
	return [...]string{
		"Insert a new entry into the database",
		"Insert a new expirable entry into the database",
		"Return the data in the database with the matching key",
		"Remove an entry in the database with the matching key",
		"Display all the saved data in the database with the format [key] -> [value]",
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
func New(db database.Store, stats *stats.Stats, cmd Op) Command {
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
	case LIST:
		command = makeListCommand(db)
	case KEYS:
		command = makeKeysCommand(db)
	case FLUSH:
		command = makeFlushCommand(db)
	case INFO:
		command = makeInfoCommand(db, stats)
	default:
		command = nil
	}

	return command
}
