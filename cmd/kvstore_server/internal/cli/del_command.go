package cli

import (
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type delCommand struct {
	db *hashtable.HashTable
}

func MakeDelCommand(db *hashtable.HashTable) Command {
	return delCommand{
		db: db,
	}
}

func (c delCommand) String() string {
	return "del"
}

func (c delCommand) Execute(args []string) []byte {
	key := args[0]

	c.db.Remove(key)
	return []byte("OK")
}
