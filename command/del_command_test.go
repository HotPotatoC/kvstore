package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/command"
)

func TestCommandDEL(t *testing.T) {
	db := HTPopulate(10)

	cmd := command.New(db, nil, command.DEL)

	result := cmd.Execute([]string{"k5"})
	if !bytes.Equal([]byte("1"), result) {
		t.Errorf("Failed TestCommandDEL -> Expected: %s | Got: %s", []byte("1"), result)
	}

	result = cmd.Execute([]string{""})
	if !bytes.Equal([]byte(command.ErrMissingKeyArg.Error()), result) {
		t.Errorf("Failed TestCommandDEL -> Expected: %s | Got: %s", []byte(command.ErrMissingKeyArg.Error()), result)
	}

	result = cmd.Execute([]string{"k11"})
	if !bytes.Equal([]byte("0"), result) {
		t.Errorf("Failed TestCommandDEL -> Expected: %s | Got: %s", []byte("0"), result)
	}
}
