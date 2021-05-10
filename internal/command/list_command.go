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
	i := 0
	for entry := range c.db.Iter() {
		// Prevent from sending more than 50 items
		if i > 50 {
			b.WriteString("...\n\n")
			break
		}
		b.WriteString(fmt.Sprintf("%s -> \"%s\"\n", entry.Key, entry.Value))
		i++
	}

	b.WriteString(fmt.Sprintf("%d items\n", c.db.Size()))
	return b.Bytes()
}
