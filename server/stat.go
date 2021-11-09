package server

import (
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

// Stats is the stats for the server
type Stats struct {
	// StartTime is the time the server was started.
	StartTime time.Time `json:"start_time"`
	// NumCommands is the number of commands processed.
	NumCommands int64 `json:"num_commands"`
	// NumConnections is the number of connections received.
	NumConnections int64 `json:"num_connections"`
}

// infoCommand is the command to get server info
// TODO: implement
func infoCommand(c *client.Client) {
	c.Conn.AsyncWrite(protocol.MakeError("NOT_IMPLEMENTED"))
}
