package cli

type Command interface {
	Execute(args string)
}

func GetCommand(cmd string) Command {
	var command Command
	switch cmd {
	case "set":
		command = MakeSetCommand()
		break
	case "get":
		command = MakeGetCommand()
		break
	case "del":
		command = MakeDelCommand()
		break
	case "list":
		command = MakeListCommand()
		break
	default:
		command = nil
		break
	}

	return command
}
