package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
)

func TestCommandPING(t *testing.T) {
	db := NewTempDB(10)

	cmd := command.New(db, nil, command.PING)

	result := cmd.Execute([]string{})

	expected := []byte("PONG")

	if !bytes.Equal(expected, result) {
		t.Errorf("Failed TestCommandPING -> Expected: %s | Got: %s", expected, result)
	}
}
