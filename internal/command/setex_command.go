package command

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/HotPotatoC/kvstore/internal/database"
)

type setEXCommand struct {
	db database.Store
}

func makeSetEXCommand(db database.Store) Command {
	return setEXCommand{
		db: db,
	}
}

func (c setEXCommand) String() string {
	return "setex"
}

func (c setEXCommand) Execute(args []string) []byte {
	if len(args) < 3 {
		return []byte(ErrInvalidArgLength.Error())
	}

	key := args[0]
	if key == "" {
		return []byte(ErrMissingKeyArg.Error())
	}

	restOfArgs := strings.Join(args[1:], " ")
	r := csv.NewReader(strings.NewReader(restOfArgs))
	r.Comma = ' '

	fields, err := r.Read()
	if err != nil {
		return []byte(ErrInternalError.Error())
	}

	value := fields[0]

	expiresAfter, err := strconv.ParseInt(fields[1], 0, 32)
	if err != nil {
		return []byte("invalid expiry seconds provided")
	}

	c.db.SetEX(key, value, int(expiresAfter))

	return []byte("OK")
}
