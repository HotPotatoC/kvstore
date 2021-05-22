package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/peterh/liner"
)

func newTerminal() *liner.State {
	liner := liner.NewLiner()

	liner.SetCtrlCAborts(true)
	liner.SetCompleter(func(s string) (c []string) {
		commands := []command.Op{
			command.SET,
			command.SETEX,
			command.GET,
			command.DEL,
			command.KEYS,
			command.FLUSHALL,
			command.INFO,
		}
		for _, n := range commands {
			if strings.HasPrefix(n.String(), strings.ToLower(s)) {
				c = append(c, n.String())
			}
		}
		return
	})

	if f, err := os.Open(filepath.Join(os.TempDir(), ".kvstore-cli-history")); err == nil {
		liner.ReadHistory(f)
		f.Close()
	}

	return liner
}
