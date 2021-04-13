package command

import (
	"fmt"

	"github.com/HotPotatoC/kvstore/database"
)

type delCommand struct {
	db database.Store
}

func makeDelCommand(db database.Store) Command {
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
