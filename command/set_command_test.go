package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/command"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

func TestCommandSET(t *testing.T) {
	db := hashtable.New()

	cmd := command.New(db, nil, command.SET)

	result := cmd.Execute([]string{"key", "value"})
	if !bytes.Equal([]byte("OK"), result) {
		t.Errorf("Failed TestCommandSET -> Expected: %s | Got: %s", []byte("OK"), result)
	}

	result = cmd.Execute([]string{"key"})
	if !bytes.Equal([]byte(command.ErrMissingKeyValueArg.Error()), result) {
		t.Errorf("Failed TestCommandSET -> Expected: %s | Got: %s", []byte(command.ErrMissingKeyValueArg.Error()), result)
	}

	result = cmd.Execute([]string{"", "value"})
	if !bytes.Equal([]byte(command.ErrMissingKeyArg.Error()), result) {
		t.Errorf("Failed TestCommandSET -> Expected: %s | Got: %s", []byte(command.ErrMissingKeyArg.Error()), result)
	}
}
