package server

import (
	"fmt"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

const (
	// GenericErrorPrefix is the prefix for all generic errors
	GenericErrorPrefix = "ERR"
)

// NewGenericError returns a new generic error
func NewGenericError(msg string) []byte {
	return protocol.MakeError(fmt.Sprintf("%s %s", GenericErrorPrefix, msg))
}
