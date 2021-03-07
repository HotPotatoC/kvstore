package command

import (
	"bytes"

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
	for _, bucket := range c.db.List() {
		if bucket != nil && bucket.Head != nil {
			b.WriteString(bucket.Head.Key)
			b.WriteString("\n")
		}
	}

	return b.Bytes()
}
