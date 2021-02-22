package command

import "github.com/HotPotatoC/kvstore/pkg/hashtable"

type CommandOp int

const (
	SET CommandOp = iota
	GET
	DEL
	LIST
)

func (c CommandOp) String() string {
	return [...]string{"set", "get", "del", "list"}[c]
}

type Command interface {
	String() string
	Execute(args []string) []byte
}

func GetCommand(db *hashtable.HashTable, cmd CommandOp) Command {
	var command Command
	switch cmd {
	case SET:
		command = MakeSetCommand(db)
		break
	case GET:
		command = MakeGetCommand(db)
		break
	case DEL:
		command = MakeDelCommand(db)
		break
	case LIST:
		command = MakeListCommand(db)
		break
	default:
		command = nil
		break
	}

	return command
}
