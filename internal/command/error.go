package command

import (
	"errors"
)

var (
	// ErrMissingKeyArg is thrown when the client did not send the key argument for the command
	ErrMissingKeyArg = errors.New("missing key argument")

	// ErrMissingValueArg is thrown when the client did not send the value argument for the command
	ErrMissingValueArg = errors.New("missing value argument")

	// ErrMissingKeyValueArg is thrown when the client did not send both the key and value arguments for the command
	ErrMissingKeyValueArg = errors.New("missing key/value arguments")

	// ErrInvalidArgLength is thrown when the client provides missing arguments or does not satisfy the amount of required args for the command
	ErrInvalidArgLength = errors.New("missing arguments")

	// ErrCommandDoesNotExist is thrown when the command does not exists in the command package
	ErrCommandDoesNotExist = errors.New("command does not exists")

	// ErrInternalError is thrown when an error on the server side has occured
	ErrInternalError = errors.New("internal error")
)
