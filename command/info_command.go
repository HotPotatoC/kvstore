package command

import (
	"github.com/HotPotatoC/kvstore/database"
	"github.com/HotPotatoC/kvstore/server/stats"
)

type infoCommand struct {
	db database.Store
	stats *stats.Stats
}

func makeInfoCommand(db database.Store, stats *stats.Stats) Command {
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
