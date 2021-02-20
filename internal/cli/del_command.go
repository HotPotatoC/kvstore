package cli

import (
	"fmt"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type delCommand struct {
	db *hashtable.HashTable
}

func MakeDelCommand(db *hashtable.HashTable) Command {
	return delCommand{
		db: db,
	}
}

func (c delCommand) String() string {
	return "del"
}

func (c delCommand) Execute(args []string) []byte {
	fmt.Printf("DelCommand called with: %s\n", args)
	return []byte("")
}
