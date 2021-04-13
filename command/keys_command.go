package command

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/database"
)

type keysCommand struct {
	db database.Store
}

func makeKeysCommand(db database.Store) Command {
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
