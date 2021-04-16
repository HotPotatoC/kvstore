package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
)

func TestCommandGET(t *testing.T) {
	db := NewTempDB(10)

	cmd := command.New(db, nil, command.GET)

	result := cmd.Execute([]string{"k5"})
	if !bytes.Equal([]byte("v5"), result) {
		t.Errorf("Failed TestCommandGET -> Expected: %s | Got: %s", []byte("v5"), result)
	}

	result = cmd.Execute([]string{""})
	if !bytes.Equal([]byte(command.ErrMissingKeyArg.Error()), result) {
		t.Errorf("Failed TestCommandGET -> Expected: %s | Got: %s", []byte("Missing key"), result)
	}

	result = cmd.Execute([]string{"k11"})
	if !bytes.Equal([]byte("<nil>"), result) {
		t.Errorf("Failed TestCommandGET -> Expected: %s | Got: %s", []byte("<nil>"), result)
	}
}
