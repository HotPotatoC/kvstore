package command

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type listCommand struct {
	db *hashtable.HashTable
}

func makeListCommand(db *hashtable.HashTable) Command {
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
		for entry != nil {
			b.WriteString(fmt.Sprintf("%s -> \"%s\"\n", entry.Key, entry.Value))
			entry = entry.Next
		}

	}

	return b.Bytes()
}
