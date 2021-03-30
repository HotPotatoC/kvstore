package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
)

func TestCommandFLUSH(t *testing.T) {
	db := HTPopulate(10)

	cmd := command.New(db, nil, command.FLUSH)

	result := cmd.Execute([]string{""})
	if !bytes.Equal([]byte("OK"), result) {
		t.Errorf("Failed TestCommandFLUSH -> Expected: %s | Got: %s", []byte("OK"), result)
	}
}
