package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
)

func TestCommandDEL(t *testing.T) {
	db := HTPopulate(10)

	cmd := command.New(db, command.DEL)

	result := cmd.Execute([]string{"k5"})
	if !bytes.Equal([]byte("1"), result) {
		t.Errorf("Failed TestCommandDEL -> Expected: %s | Got: %s", []byte("1"), result)
	}

	result = cmd.Execute([]string{""})
	if !bytes.Equal([]byte("Missing key"), result) {
		t.Errorf("Failed TestCommandDEL -> Expected: %s | Got: %s", []byte("Missing key"), result)
	}

	result = cmd.Execute([]string{"k11"})
	if !bytes.Equal([]byte("0"), result) {
		t.Errorf("Failed TestCommandDEL -> Expected: %s | Got: %s", []byte("0"), result)
	}
}
