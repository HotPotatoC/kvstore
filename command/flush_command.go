package command

import (
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type flushCommand struct {
	db *hashtable.HashTable
}

func makeFlushCommand(db *hashtable.HashTable) Command {
	return flushCommand{
		db: db,
	}
}

func (c flushCommand) String() string {
	return "flush"
}

func (c flushCommand) Execute(args []string) []byte {
	c.db.Flush()
	return []byte("OK")
}
