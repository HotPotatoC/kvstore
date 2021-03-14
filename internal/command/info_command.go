package command

import (
	"github.com/HotPotatoC/kvstore/internal/stats"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)

type infoCommand struct {
	db *hashtable.HashTable
	stats *stats.Stats
}

func makeInfoCommand(db *hashtable.HashTable, stats *stats.Stats) Command {
	return infoCommand{
		db: db,
		stats: stats,
	}
}

func (c infoCommand) String() string {
	return "info"
}

func (c infoCommand) Execute(args []string) []byte {
	c.stats.UpdateMemStats()
	c.stats.UpdateUptime()

	infoData, err := c.stats.JSON()
	if err != nil {
		return []byte(ErrInternalError.Error())
	}
	return infoData
}
