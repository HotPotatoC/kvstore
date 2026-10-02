package server

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
)

func infoPayload(t *testing.T, response string) string {
	t.Helper()
	header, payload, ok := strings.Cut(response, "\r\n")
	if !ok || !strings.HasPrefix(header, "$") {
		t.Fatalf("not a bulk INFO response: %q", response)
	}
	n, err := strconv.Atoi(header[1:])
	if err != nil || len(payload) != n+2 || !strings.HasSuffix(payload, "\r\n") {
		t.Fatalf("invalid INFO bulk length: %q", response)
	}
	return payload[:n]
}

func infoValues(payload string) map[string]string {
	values := make(map[string]string)
	for _, line := range strings.Split(payload, "\r\n") {
		if key, value, ok := strings.Cut(line, ":"); ok {
			values[key] = value
		}
	}
	return values
}

func testInfo(s *Server, args ...string) string {
	c := &client.Client{Argc: len(args)}
	for _, arg := range args {
		c.Argv = append(c.Argv, []byte(arg))
	}
	var response bytes.Buffer
	s.infoCommand(c, &response)
	return response.String()
}

func TestINFOSectionsAndLimits(t *testing.T) {
	s := newTestServer(t)
	s.DB.SetMaxMemory(12345)
	s.DB.Store(datastructure.NewItem("permanent", "value", 0))
	s.DB.Store(datastructure.NewItem("expiring", "value", time.Hour))
	sections := []string{"Server", "Clients", "Memory", "Stats", "Persistence", "Keyspace"}
	for _, args := range [][]string{nil, {"default"}, {"ALL"}} {
		payload := infoPayload(t, testInfo(s, args...))
		for _, section := range sections {
			if !strings.Contains(payload, "# "+section+"\r\n") {
				t.Fatalf("INFO %v missing %s", args, section)
			}
		}
		values := infoValues(payload)
		if values["maxmemory"] != "12345" || values["used_memory"] != strconv.FormatInt(s.DB.UsedMemory(), 10) || values["db0"] != "keys=2,expires=1" || values["snapshot_last_save_status"] != "never" {
			t.Fatalf("unexpected INFO values: %v", values)
		}
	}
	for _, section := range sections {
		payload := infoPayload(t, testInfo(s, strings.ToUpper(section)))
		if strings.Count(payload, "# ") != 1 || !strings.HasPrefix(payload, "# "+section+"\r\n") {
			t.Fatalf("INFO %s response=%q", section, payload)
		}
	}
	if payload := infoPayload(t, testInfo(s, "unknown")); payload != "" {
		t.Fatalf("unknown INFO section response=%q", payload)
	}
	if response := testInfo(s, "all", "stats"); !strings.HasPrefix(response, "-ERR ") {
		t.Fatalf("INFO extra argument response=%q", response)
	}
	var response bytes.Buffer
	s.infoCommand(&client.Client{OutputLimit: 40}, &response)
	if response.String() != "-ERR response exceeds output limit\r\n" {
		t.Fatalf("bounded INFO response=%q", response.String())
	}
}

func TestINFOTrafficCountersInlineAndWorker(t *testing.T) {
	s := newTestServer(t)
	conn := openTestConn(s)
	var requests []byte
	for _, args := range [][]string{{"SET", "key", "value"}, {"GET"}, {"NOPE"}, {"PING"}} {
		requests = append(requests, protocol.MakeCommand(args...)...)
	}
	conn.input = requests
	traffic(t, s, conn)
	if s.NumCommands.Load() != 4 || s.errorReplies.Load() != 2 {
		t.Fatalf("inline counts commands=%d errors=%d", s.NumCommands.Load(), s.errorReplies.Load())
	}
	conn.output.Reset()
	conn.input = protocol.MakeCommand("INFO", "all", "extra")
	traffic(t, s, conn)
	waitWake(t, conn)
	traffic(t, s, conn)
	if s.NumCommands.Load() != 5 || s.errorReplies.Load() != 3 || !strings.HasPrefix(conn.output.String(), "-ERR ") {
		t.Fatal("worker INFO errors were not counted")
	}
	conn.output.Reset()
	conn.input = protocol.MakeCommand("INFO", "stats")
	traffic(t, s, conn)
	waitWake(t, conn)
	traffic(t, s, conn)
	values := infoValues(infoPayload(t, conn.output.String()))
	if values["total_commands_processed"] != "6" || values["total_error_replies"] != "3" || values["total_connections_received"] != "1" {
		t.Fatalf("traffic INFO stats=%v", values)
	}
}

func TestINFOResourceRejectionCounters(t *testing.T) {
	s := newTestServer(t)
	s.limits.maxClients = 1
	conn := openTestConn(s)
	rejected := &testConn{}
	if _, action := s.OnOpen(rejected); action != gnet.Close || s.rejectedConnections.Load() != 1 || s.activeClients.Load() != 1 || s.NumConnections.Load() != 1 {
		t.Fatal("maxclients rejection not counted")
	}
	s.DB.SetMaxMemory(1)
	conn.input = protocol.MakeCommand("SET", "key", "value")
	traffic(t, s, conn)
	if s.oomErrors.Load() != 1 || s.rejectedRequests.Load() != 1 || s.errorReplies.Load() != 2 {
		t.Fatal("OOM rejection not counted")
	}
	releases := make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	for i := 0; i < s.limits.queue; i++ {
		s.jobs <- func() {}
	}
	conn.input = protocol.MakeCommand("INFO")
	traffic(t, s, conn)
	if s.overloadErrors.Load() != 1 || s.rejectedRequests.Load() != 2 || s.errorReplies.Load() != 3 || s.NumCommands.Load() != 1 {
		t.Fatal("worker overload rejection not counted")
	}
	for _, release := range releases {
		close(release)
	}
	conn.input = []byte("*1\r\n:1\r\n")
	if action := s.OnTraffic(conn); action != gnet.Close || s.rejectedRequests.Load() != 3 || s.errorReplies.Load() != 4 {
		t.Fatal("protocol rejection not counted")
	}
	s.OnClose(conn, nil)
	if s.activeClients.Load() != 0 {
		t.Fatal("closed client remains counted")
	}
}

func TestINFOClientSubcommandStillReturnsClientDetails(t *testing.T) {
	s := newTestServer(t)
	conn := openTestConn(s)
	conn.input = protocol.MakeCommand("CLIENT", "INFO")
	traffic(t, s, conn)
	payload := infoPayload(t, conn.output.String())
	if !strings.HasPrefix(payload, "id=1 addr=") || strings.Contains(payload, "# Server") {
		t.Fatalf("CLIENT INFO response=%q", payload)
	}
}

func TestINFOConcurrentCommandsAndSnapshots(t *testing.T) {
	s := newTestServer(t)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := &client.Client{DB: s.DB}
			for j := 0; j < 100; j++ {
				var response bytes.Buffer
				s.processCommand([][]byte{[]byte("set"), []byte("key"), []byte("value")}, commandSet, c, &response)
				s.DB.Expire("key", time.Hour)
			}
		}()
	}
	for i := 0; i < 100; i++ {
		infoPayload(t, testInfo(s))
	}
	wg.Wait()
	if s.NumCommands.Load() != 400 {
		t.Fatalf("concurrent commands=%d", s.NumCommands.Load())
	}
}
