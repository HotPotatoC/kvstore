package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

func TestCommandSET(t *testing.T) {
	db := hashtable.New()

	cmd := command.New(db, command.SET)

	result := cmd.Execute([]string{"key", "value"})
	if !bytes.Equal([]byte("OK"), result) {
		t.Errorf("Failed TestCommandSET -> Expected: %s | Got: %s", []byte("OK"), result)
	}

	result = cmd.Execute([]string{"key"})
	if !bytes.Equal([]byte("Missing key/value arguments"), result) {
		t.Errorf("Failed TestCommandSET -> Expected: %s | Got: %s", []byte("Missing key/value arguments"), result)
	}

	result = cmd.Execute([]string{"", "value"})
	if !bytes.Equal([]byte("Missing key"), result) {
		t.Errorf("Failed TestCommandSET -> Expected: %s | Got: %s", []byte("Missing key"), result)
	}
}
