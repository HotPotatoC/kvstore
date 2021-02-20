package cli

import "github.com/HotPotatoC/kvstore/pkg/hashtable"

type Command interface {
	String() string
	Execute(args []string) []byte
}

func GetCommand(db *hashtable.HashTable, cmd string) Command {
	var command Command
	switch cmd {
	case "set":
		command = MakeSetCommand(db)
		break
	case "get":
		command = MakeGetCommand(db)
		break
	case "del":
		command = MakeDelCommand(db)
		break
	case "list":
		command = MakeListCommand(db)
		break
	default:
		command = nil
		break
	}

	return command
}
