package command

import (
	"fmt"

	"github.com/HotPotatoC/kvstore/internal/storage"
)

type delCommand struct {
	db storage.Store
}

func makeDelCommand(db storage.Store) Command {
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
		return []byte(ErrMissingKeyArg.Error())
	}

	count := c.db.Remove(key)
	return []byte(fmt.Sprintf("%d", count))
}
