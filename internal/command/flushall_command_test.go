package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
)

func TestCommandFLUSHALL(t *testing.T) {
	db := NewTempDB(10)

	cmd := command.New(db, nil, command.FLUSHALL)

	result := cmd.Execute([]string{""})
	if !bytes.Equal([]byte("OK"), result) {
		t.Errorf("Failed TestCommandFLUSHALL -> Expected: %s | Got: %s", []byte("OK"), result)
	}
}
