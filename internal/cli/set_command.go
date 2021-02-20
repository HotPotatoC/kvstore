package cli

import "fmt"

type setCommand struct{}

func MakeSetCommand() Command {
	return setCommand{}
}

func (c setCommand) String() string {
	return "set"
}

func (c setCommand) Execute(args string) {
	fmt.Printf("SetCommand called with: %s\n", args)
}
