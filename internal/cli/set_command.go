package cli

import (
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type setCommand struct {
	db *hashtable.HashTable
}

func MakeSetCommand(db *hashtable.HashTable) Command {
	return setCommand{
		db: db,
	}
}

func (c setCommand) String() string {
	return "set"
}

func (c setCommand) Execute(args []string) []byte {
	if len(args) < 2 {
		return []byte("Missing key/value arguments")
	}
	key := args[0]
	value := args[1]

	c.db.Set(key, value)

	return []byte("")
}
