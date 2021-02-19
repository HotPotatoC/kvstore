package cli

import "fmt"

type listCommand struct{}

func MakeListCommand() Command {
	return listCommand{}
}

func (c listCommand) Execute(args string) {
	fmt.Printf("ListCommand called with: %s\n", args)
}
