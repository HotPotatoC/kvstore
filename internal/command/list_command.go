package command

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type listCommand struct {
	db *hashtable.HashTable
}

func MakeListCommand(db *hashtable.HashTable) Command {
	return listCommand{
		db: db,
	}
}

func (c listCommand) String() string {
	return "list"
}

func (c listCommand) Execute(args []string) []byte {
	var b bytes.Buffer
	i := 1

	for k, v := range c.db.List() {
		b.WriteString(fmt.Sprintf("%d) [%d]: \"%s\"\n", i, k, v))
		i++
	}

	return b.Bytes()
}
