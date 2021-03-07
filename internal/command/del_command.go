package command

import (
	"fmt"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type delCommand struct {
	db *hashtable.HashTable
}

func makeDelCommand(db *hashtable.HashTable) Command {
	return delCommand{
		db: db,
	}
}

func (c delCommand) String() string {
	return "del"
}

func (c delCommand) Execute(args []string) []byte {
	key := args[0]
	if key == "" {
		return []byte("Missing key")
	}

	count := c.db.Remove(key)
	return []byte(fmt.Sprintf("%d", count))
}
