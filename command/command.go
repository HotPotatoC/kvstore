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

func Parse(data []interface{}) ([]byte, [][]byte) {
	recvCmd, rawRecvArgv := bytes.ToLower(data[0].([]byte)), data[1:]

	var recvArgv [][]byte
	for _, v := range rawRecvArgv {
		recvArgv = append(recvArgv, v.([]byte))
	}

	return recvCmd, recvArgv
}
