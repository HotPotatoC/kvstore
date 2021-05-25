package command

import (
	"fmt"
	"strings"

	"github.com/HotPotatoC/kvstore/internal/server/stats"
	"github.com/HotPotatoC/kvstore/internal/storage"
)

// Op represents the command operation code
type Op int

const (
	// SET inserts a new entry into the database
	SET Op = iota
	// SETEX inserts a new expirable entry into the database
	SETEX
	// GET returns the data in the database with the matching key
	GET
	// DEL remove an entry in the database with the matching key
	DEL
	// KEYS displays all the saved keys in the database
	KEYS
	// FLUSHALL delete all keys
	FLUSHALL
	// INFO displays the current status of the server (memory allocs, connected clients, uptime, etc.)
	INFO
	// Pings the server
	PING
)

const (
	ReadMode    = "r"
	WriteMode   = "w"
	PersistMode = "+"
)

func (c Op) String() string {
	return [...]string{
		"set",
		"setex",
		"get",
		"del",
		"keys",
		"flushall",
		"info",
		"ping",
	}[c]
}

func (c Op) Bytes() []byte {
	return [...][]byte{
		[]byte("set"),
		[]byte("setex"),
		[]byte("get"),
		[]byte("del"),
		[]byte("keys"),
		[]byte("flushall"),
		[]byte("info"),
		[]byte("ping"),
	}[c]
}

// Args is the command operation required arguments
func (c Op) Args() string {
	return [...]string{
		"key value",
		"key value [exp seconds]",
		"key",
		"key",
		"",
		"",
		"",
		"",
	}[c]
}

// Description is the command operation description
func (c Op) Description() string {
	return [...]string{
		"Insert a new entry into the database",
		"Insert a new expirable entry into the database",
		"Return the data in the database with the matching key",
		"Remove an entry in the database with the matching key",
		"Display all the saved keys in the database",
		"Delete all keys",
		"Display the current stats of the server (OS, mem usage, total connections, etc.) in json format",
		"Pings the server",
	}[c]
}

// Opts is the command operation options which can be
// Read (r), Write (w) and Persist (+)
func (c Op) Opts() string {
	return [...]string{
		fmt.Sprintf("%s%s", PersistMode, WriteMode),
		fmt.Sprintf("%s%s", PersistMode, WriteMode),
		ReadMode,
		fmt.Sprintf("%s%s", PersistMode, WriteMode),
		ReadMode,
		fmt.Sprintf("%s%s", PersistMode, WriteMode),
		ReadMode,
		ReadMode,
	}[c]
}

// Command is the set of methods for a commmand
type Command interface {
	String() string
	Execute(args []string) []byte
}

type Options struct {
	Op      Op
	Full    string
	Command string
	Args    []string
	Mode    []string
}

// New constructs the given command operation
func New(db storage.Store, stats *stats.Stats, cmd Op) Command {
	var command Command
	switch cmd {
	case SET:
		command = makeSetCommand(db)
	case SETEX:
		command = makeSetEXCommand(db)
	case GET:
		command = makeGetCommand(db)
	case DEL:
		command = makeDelCommand(db)
	case KEYS:
		command = makeKeysCommand(db)
	case FLUSHALL:
		command = makeFlushAllCommand(db)
	case INFO:
		command = makeInfoCommand(db, stats)
	case PING:
		command = makePingCommand(db)
	default:
		command = nil
	}

	return command
}

func Parse(input string) (Options, error) {
	raw := strings.Fields(string(input))
	cmd := strings.ToLower(
		strings.TrimSpace(raw[0]))
	args := strings.Split(
		strings.TrimSpace(
			strings.TrimPrefix(
				string(input), raw[0])),
		" ")

	var op Op
	switch string(cmd) {
	case SET.String():
		op = SET
	case SETEX.String():
		op = SETEX
	case DEL.String():
		op = DEL
	case FLUSHALL.String():
		op = FLUSHALL
	case GET.String():
		op = GET
	case KEYS.String():
		op = KEYS
	case INFO.String():
		op = INFO
	case PING.String():
		op = PING
	default:
		return Options{}, ErrCommandDoesNotExist
	}

	return Options{
		Op:      op,
		Full:    fmt.Sprintf("%s %s", op.String(), strings.Join(args, " ")),
		Command: op.String(),
		Args:    args,
		Mode:    strings.Split(op.Opts(), ""),
	}, nil
}
