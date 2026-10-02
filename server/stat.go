package server

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/build"
	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

// Stats is the stats for the server
type Stats struct {
	// StartTime is the time the server was started.
	StartTime time.Time `json:"start_time"`
	// NumCommands is the number of commands processed.
	NumCommands atomic.Int64
	// NumConnections is the number of connections received.
	NumConnections      atomic.Int64
	rejectedConnections atomic.Int64
	errorReplies        atomic.Int64
	rejectedRequests    atomic.Int64
	oomErrors           atomic.Int64
	overloadErrors      atomic.Int64
}

// infoCommand is the command to get server info
func (s *Server) infoCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc > 1 {
		res.Write(NewGenericError("wrong number of arguments for 'info' command"))
		return
	}
	section := "default"
	if c.Argc == 1 {
		section = strings.ToLower(string(c.Argv[0]))
	}
	var info bytes.Buffer
	for _, name := range []string{"server", "clients", "memory", "stats", "persistence", "keyspace"} {
		if section != "default" && section != "all" && section != name {
			continue
		}
		if info.Len() > 0 {
			info.WriteString("\r\n")
		}
		switch name {
		case "server":
			fmt.Fprintf(&info, "# Server\r\nkvstore_version:%s\r\ngo_version:%s\r\nprocess_id:%d\r\nuptime_in_seconds:%d\r\n", build.Version, runtime.Version(), s.PID, int64(time.Since(s.StartTime)/time.Second))
		case "clients":
			fmt.Fprintf(&info, "# Clients\r\nconnected_clients:%d\r\nmaxclients:%d\r\n", s.activeClients.Load(), s.limits.maxClients)
		case "memory":
			fmt.Fprintf(&info, "# Memory\r\nused_memory:%d\r\nmaxmemory:%d\r\nmaxmemory_policy:noeviction\r\nmemory_accounting:keys_values_estimated_metadata\r\n", s.DB.UsedMemory(), s.DB.MaxMemory())
		case "stats":
			fmt.Fprintf(&info, "# Stats\r\ntotal_connections_received:%d\r\nrejected_connections:%d\r\ntotal_commands_processed:%d\r\ntotal_error_replies:%d\r\nrejected_requests:%d\r\noom_errors:%d\r\noverload_errors:%d\r\n", s.NumConnections.Load(), s.rejectedConnections.Load(), s.NumCommands.Load(), s.errorReplies.Load(), s.rejectedRequests.Load(), s.oomErrors.Load(), s.overloadErrors.Load())
		case "persistence":
			stats := disk.SnapshotStats{LastStatus: "never"}
			enabled, saving := 0, 0
			if s.kvsDB != nil {
				enabled = 1
				stats = s.kvsDB.SnapshotStats()
			}
			if stats.InProgress {
				saving = 1
			}
			fmt.Fprintf(&info, "# Persistence\r\nsnapshot_enabled:%d\r\nsnapshot_in_progress:%d\r\nsnapshot_last_save_status:%s\r\nsnapshot_save_failures:%d\r\nsnapshot_last_successful_save_time:%d\r\n", enabled, saving, stats.LastStatus, stats.Failures, stats.LastSuccess)
		case "keyspace":
			keys, expires := s.DB.KeyspaceStats()
			fmt.Fprintf(&info, "# Keyspace\r\ndb0:keys=%d,expires=%d\r\n", keys, expires)
		}
	}
	if c.OutputLimit > 0 && info.Len() > c.OutputLimit-32 {
		res.Write(NewGenericError("response exceeds output limit"))
		return
	}
	protocol.WriteBulkString(res, info.String())
}
