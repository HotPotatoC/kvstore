package cli

import "fmt"

type delCommand struct{}

func MakeDelCommand() Command {
	return delCommand{}
}

func (c delCommand) String() string {
	return "del"
}

func (c delCommand) Execute(args string) {
	fmt.Printf("DelCommand called with: %s\n", args)
}
