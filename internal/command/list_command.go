package command

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/internal/database"
)

type listCommand struct {
	db database.Store
}

func makeListCommand(db database.Store) Command {
	return listCommand{
		db: db,
	}
}

func (c listCommand) String() string {
	return "list"
}

func (c listCommand) Execute(args []string) []byte {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%d items\n", c.db.Size()))
	for entry := range c.db.Iter() {
		b.WriteString(fmt.Sprintf("%s -> \"%s\"\n", entry.Key, entry.Value))
	}

	return b.Bytes()
}
