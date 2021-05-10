package command

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/internal/storage"
)

type keysCommand struct {
	db storage.Store
}

func makeKeysCommand(db storage.Store) Command {
	return keysCommand{
		db: db,
	}
}

func (c keysCommand) String() string {
	return "list"
}

func (c keysCommand) Execute(args []string) []byte {
	var b bytes.Buffer
	idx := 1
	for entry := range c.db.Iter() {
		// Prevent from sending more than 50 keys
		if idx > 50 {
			b.WriteString("...\n\n")
			break
		}
		b.WriteString(fmt.Sprintf("%d) %s\n", idx, entry.Key))
		idx++
	}

	b.WriteString(fmt.Sprintf("%d keys\n", c.db.Size()))
	return b.Bytes()
}
