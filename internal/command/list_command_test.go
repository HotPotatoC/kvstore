package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
)

func TestCommandLIST(t *testing.T) {
	db := NewTempDB(10)

	cmd := command.New(db, nil, command.LIST)

	result := cmd.Execute([]string{})

	expected := []byte{107, 49, 48, 32, 45, 62, 32, 34, 118, 49, 48, 34, 10, 107, 53, 32, 45, 62, 32, 34, 118, 53, 34, 10, 107, 50, 32, 45, 62, 32, 34, 118, 50, 34, 10, 107, 52, 32, 45, 62, 32, 34, 118, 52, 34, 10, 107, 54, 32, 45, 62, 32, 34, 118, 54, 34, 10, 107, 57, 32, 45, 62, 32, 34, 118, 57, 34, 10, 107, 56, 32, 45, 62, 32, 34, 118, 56, 34, 10, 107, 51, 32, 45, 62, 32, 34, 118, 51, 34, 10, 107, 55, 32, 45, 62, 32, 34, 118, 55, 34, 10, 107, 49, 32, 45, 62, 32, 34, 118, 49, 34, 10, 49, 48, 32, 105, 116, 101, 109, 115, 10}

	if !bytes.Equal(expected, result) {
		t.Errorf("Failed TestCommandLIST -> Expected: %s | Got: %s", expected, result)
	}
}
