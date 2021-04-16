package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

func TestCommandSETEX(t *testing.T) {
	db := hashtable.New()

	cmd := command.New(db, nil, command.SETEX)

	result := cmd.Execute([]string{"key", "value", "5"})
	if !bytes.Equal([]byte("OK"), result) {
		t.Errorf("Failed TestCommandSETEX -> Expected: %s | Got: %s", []byte("OK"), result)
	}

	result = cmd.Execute([]string{"key"})
	if !bytes.Equal([]byte(command.ErrInvalidArgLength.Error()), result) {
		t.Errorf("Failed TestCommandSETEX -> Expected: %s | Got: %s", []byte(command.ErrInvalidArgLength.Error()), result)
	}

	result = cmd.Execute([]string{"", "value", "5"})
	if !bytes.Equal([]byte(command.ErrMissingKeyArg.Error()), result) {
		t.Errorf("Failed TestCommandSETEX -> Expected: %s | Got: %s", []byte(command.ErrMissingKeyArg.Error()), result)
	}

	result = cmd.Execute([]string{"key", "value", "invalid-seconds"})
	if !bytes.Equal([]byte("invalid expiry seconds provided"), result) {
		t.Errorf("Failed TestCommandSETEX -> Expected: %s | Got: %s", []byte("invalid expiry seconds provided"), result)
	}
}
