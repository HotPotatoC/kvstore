package command

import (
	"errors"
)

var (
	ErrMissingKeyArg       = errors.New("missing key argument")
	ErrMissingValueArg     = errors.New("missing value argument")
	ErrMissingKeyValueArg  = errors.New("missing key/value arguments")
	ErrCommandDoesNotExist = errors.New("command does not exists")

	ErrInternalError = errors.New("internal error")
)
