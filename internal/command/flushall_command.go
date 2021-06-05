package command

import (
	"github.com/HotPotatoC/kvstore/internal/storage"
)

type flushAllCommand struct {
	db storage.Store
}

func makeFlushAllCommand(db storage.Store) Command {
	return flushAllCommand{
		db: db,
	}
}

func (c flushAllCommand) String() string {
	return "flushall"
}

func (c flushAllCommand) Execute(args []string) []byte {
	c.db.Flush()
	return []byte("OK")
}
