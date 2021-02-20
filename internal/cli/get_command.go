package cli

import "fmt"

type getCommand struct{}

func MakeGetCommand() Command {
	return getCommand{}
}

func (c getCommand) String() string {
	return "get"
}

func (c getCommand) Execute(args string) {
	fmt.Printf("GetCommand called with: %s\n", args)
}
