package command

import (
	"github.com/HotPotatoC/kvstore/internal/storage"
)

type flushCommand struct {
	db storage.Store
}

func makeFlushCommand(db storage.Store) Command {
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
