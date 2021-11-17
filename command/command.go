package command

import "github.com/HotPotatoC/kvstore-rewrite/client"

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

// NewCommand creates a new command
func NewCommand(name, desc string, t Type, proc Proc) *Command {
	return &Command{Name: name, Description: desc, Type: t, Proc: proc}
}