package command_test

import (
	"encoding/json"
	"testing"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/stats"
)

func TestCommandINFO(t *testing.T) {
	stats := &stats.Stats{}

	cmd := command.New(nil, stats, command.INFO)

	result := cmd.Execute([]string{})

	if !json.Valid(result) {
		t.Errorf("Failed TestCommandINFO -> Expected: %s | Got: %s", "a_valid_json_result", result)
	}
}
