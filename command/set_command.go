package command

import (
	"strings"

	"github.com/HotPotatoC/kvstore/database"
)

type setCommand struct {
	db database.Store
}

func makeSetCommand(db database.Store) Command {
	return setCommand{
		db: db,
	}
}

func (c setCommand) String() string {
	return "set"
}

func (c setCommand) Execute(args []string) []byte {
	if len(args) < 2 {
		return []byte(ErrMissingKeyValueArg.Error())
	}

	key := args[0]
	if key == "" {
		return []byte(ErrMissingKeyArg.Error())
	}

	value := strings.Join(args[1:], " ")

	c.db.Set(key, value)

	return []byte("OK")
}
