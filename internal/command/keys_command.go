package command

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type keysCommand struct {
	db *hashtable.HashTable
}

func makeKeysCommand(db *hashtable.HashTable) Command {
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
		b.WriteString(fmt.Sprintf("%d) %s", idx, entry.Key))
		b.WriteString("\n")
		idx++
	}

	return b.Bytes()
}
