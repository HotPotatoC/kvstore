package cli

import (
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
	fmt.Printf("ListCommand called with: %s\n", args)
	return []byte("")
}
