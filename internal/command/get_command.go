package command

import (
	"github.com/HotPotatoC/kvstore/internal/database"
)

type getCommand struct {
	db database.Store
}

func makeGetCommand(db database.Store) Command {
	return getCommand{
		db: db,
	}
}

func (c getCommand) String() string {
	return "get"
}

func (c getCommand) Execute(args []string) []byte {
	key := args[0]
	if key == "" {
		return []byte(ErrMissingKeyArg.Error())
	}

	result := c.db.Get(key)
	if result == "" {
		return []byte("<nil>")
	}

	return []byte(result)
}
