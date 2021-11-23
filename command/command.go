package command

import (
	"bytes"

	"github.com/HotPotatoC/kvstore-rewrite/client"
)

// Command represents the command
type Command struct {
	// Name is the command name
	Name string
	// Description is the command description
	Description string
	// Type is the command type (read, write, etc)
	Type Type
	// Proc is the command processor
	Proc Proc
	// SubCommands is the sub-commands
	SubCommands map[string]Command
}

// Proc is the command processor
type Proc func(client *client.Client)

// Type is the command type (read, write, etc)
type Type uint8

const (
	// Read is the read command type
	Read Type = 0x01
	// Write is the write command type
	Write Type = 0x02

	// ReadWrite is the read-write command type
	ReadWrite Type = Read | Write
)

func WrapArgsFromQuotes(args [][]byte) [][]byte {
	var wrappedArgs [][]byte
	var buf bytes.Buffer

	for _, arg := range args {
		if arg[0] == '"' {
			if buf.Len() > 0 {
				wrappedArgs = append(wrappedArgs, buf.Bytes())
				buf.Reset()
			}
			buf.Write(arg[1:])
		} else {
			buf.Write(arg)
		}
	}

	if buf.Len() > 0 {
		wrappedArgs = append(wrappedArgs, buf.Bytes())
	}

	return wrappedArgs
}
