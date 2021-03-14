package command

import (
	"github.com/HotPotatoC/kvstore/internal/stats"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

// Op represents the command type
type Op int

const (
	// SET inserts a new entry into the database
	SET Op = iota
	// GET returns the data in the database with the matching key
	GET
	// DEL remove an entry in the database with the matching key
	DEL
	// LIST displays all the saved data in the database
	LIST
	// KEYS displays all the saved keys in the database
	KEYS

	// INFO displays the current status of the server (memory allocs, connected clients, uptime, etc.)
	INFO
)

func (c Op) String() string {
	return [...]string{"set", "get", "del", "list", "keys", "info"}[c]
}

// Command is the set of methods for a commmand
type Command interface {
	String() string
	Execute(args []string) []byte
}

// New constructs the given command operation
func New(db *hashtable.HashTable, stats *stats.Stats, cmd Op) Command {
	var command Command
	switch cmd {
	case SET:
		command = makeSetCommand(db)
		break
	case GET:
		command = makeGetCommand(db)
		break
	case DEL:
		command = makeDelCommand(db)
		break
	case LIST:
		command = makeListCommand(db)
		break
	case KEYS:
		command = makeKeysCommand(db)
		break
	case INFO:
		command = makeInfoCommand(db, stats)
		break
	default:
		command = nil
		break
	}

	return command
}
