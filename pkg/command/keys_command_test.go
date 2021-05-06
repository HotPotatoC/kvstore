package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/pkg/command"
)

func TestCommandKEYS(t *testing.T) {
	db := NewTempDB(10)

	cmd := command.New(db, nil, command.KEYS)

	result := cmd.Execute([]string{})

	expected := []byte{49, 41, 32, 107, 49, 48, 10, 50, 41, 32, 107, 53, 10, 51, 41, 32, 107, 50, 10, 52, 41, 32, 107, 52, 10, 53, 41, 32, 107, 54, 10, 54, 41, 32, 107, 57, 10, 55, 41, 32, 107, 56, 10, 56, 41, 32, 107, 51, 10, 57, 41, 32, 107, 55, 10, 49, 48, 41, 32, 107, 49, 10, 49, 48, 32, 107, 101, 121, 115, 10}

	if !bytes.Equal(expected, result) {
		t.Errorf("Failed TestCommandKEYS -> Expected: %s | Got: %s", expected, result)
	}
}
