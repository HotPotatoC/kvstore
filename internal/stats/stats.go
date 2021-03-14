package stats

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/utils"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

// Stats stores the information of the kvstore server
type Stats struct {
	OS                    string        `json:"os"`
	ARCH                  string        `json:"os_arch"`
	GoVersion             string        `json:"go_version"`
	ProcessID             int           `json:"process_id"`
	TCPHost               string        `json:"tcp_host"`
	TCPPort               int           `json:"tcp_port"`
	Uptime                time.Duration `json:"server_uptime"`
	UptimeHuman           string        `json:"server_uptime_human"`
	ConnectedCount        uint64        `json:"connected_clients"`
	TotalConnectionsCount uint64        `json:"total_connections_count"`
	MemoryUsage           uint64        `json:"memory_usage"`
	MemoryUsageHuman      string        `json:"memory_usage_human"`
	MemoryTotalAlloc      uint64        `json:"memory_total_alloc"`
}

// Init sets the default values
func (s *Stats) Init() {
	if s.OS == "" {
		s.OS = runtime.GOOS
	}

	if s.ARCH == "" {
		s.ARCH = runtime.GOARCH
	}

	if s.GoVersion == "" {
		s.GoVersion = runtime.Version()
	}

	s.ConnectedCount = 0
	s.TotalConnectionsCount = 0

	s.ProcessID = os.Getpid()
}

// UpdateUptime updates the Uptime and UptimeHuman fields to the current time
func (s *Stats) UpdateUptime() {
	s.Uptime = time.Since(startTime)
	s.UptimeHuman = utils.FormatDuration(time.Since(startTime))
}

// UpdateMemStats updates the MemoryUsage, MemoryUsageHuman and MemoryTotalAlloc fields with the
// current memory usage, and total allocations
func (s *Stats) UpdateMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s.MemoryUsage = m.Alloc
	s.MemoryUsageHuman = utils.ByteCount(m.Alloc)
	s.MemoryTotalAlloc = m.TotalAlloc
}

// JSON returns the json representation of this struct
func (s *Stats) JSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "\t")
}
