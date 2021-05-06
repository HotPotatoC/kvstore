package command

import (
	"github.com/HotPotatoC/kvstore/pkg/database"
)

type flushCommand struct {
	db database.Store
}

func makeFlushCommand(db database.Store) Command {
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
