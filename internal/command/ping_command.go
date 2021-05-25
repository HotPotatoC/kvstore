package command

import (
	"github.com/HotPotatoC/kvstore/internal/storage"
)

type pingCommand struct {
	db storage.Store
}

func makePingCommand(db storage.Store) Command {
	return pingCommand{
		db: db,
	}
}

func (c pingCommand) String() string {
	return "ping"
}

func (c pingCommand) Execute(args []string) []byte {
	return []byte("PONG")
}
