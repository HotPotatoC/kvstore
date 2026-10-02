package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
)

type testConn struct {
	gnet.Conn
	input      []byte
	ctx        any
	output     bytes.Buffer
	capture    bool
	writeBytes int
	wakes      chan struct{}
	closed     atomic.Bool
	discards   []int
	outbound   int
	writeErr   error
}

func (c *testConn) Context() any       { return c.ctx }
func (c *testConn) SetContext(ctx any) { c.ctx = ctx }
func (c *testConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}
func (c *testConn) InboundBuffered() int       { return len(c.input) }
func (c *testConn) OutboundBuffered() int      { return c.outbound }
func (c *testConn) Peek(n int) ([]byte, error) { return c.input[:n], nil }
func (c *testConn) Discard(n int) (int, error) {
	if n <= 0 {
		panic("Discard(0) loses fragments")
	}
	c.discards = append(c.discards, n)
	c.input = c.input[n:]
	return n, nil
}
func (c *testConn) Write(data []byte) (int, error) {
	if c.writeErr != nil {
		return 0, c.writeErr
	}
	c.writeBytes += len(data)
	if c.capture {
		c.output.Write(data)
	}
	return len(data), nil
}
func (c *testConn) Wake(_ gnet.AsyncCallback) error {
	select {
	case c.wakes <- struct{}{}:
	default:
	}
	return nil
}
func (c *testConn) Close() error { c.closed.Store(true); return nil }
func newTestServer(t testing.TB) *Server {
	t.Helper()
	s := &Server{DB: datastructure.NewMap(), limits: serverLimits()}
	s.startWorkers()
	t.Cleanup(s.shutdown)
	return s
}
func openTestConn(s *Server) *testConn {
	c := &testConn{capture: true, wakes: make(chan struct{}, 100)}
	s.OnOpen(c)
	return c
}
func traffic(t *testing.T, s *Server, c *testConn) {
	t.Helper()
	if got := s.OnTraffic(c); got != gnet.None {
		t.Fatalf("traffic action=%v", got)
	}
}
func waitWake(t *testing.T, c *testConn) {
	t.Helper()
	select {
	case <-c.wakes:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not wake connection")
	}
}
func blockWorker(t *testing.T, s *Server) chan struct{} {
	t.Helper()
	release, started := make(chan struct{}), make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	s.jobs <- func() { close(started); <-release }
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not start")
	}
	return release
}

func TestTrafficRetainsFragments(t *testing.T) {
	s := newTestServer(t)
	c := openTestConn(s)
	data := protocol.MakeCommand("PiNg")
	for i, ch := range data {
		c.input = append(c.input, ch)
		traffic(t, s, c)
		if i < len(data)-1 && (c.output.Len() != 0 || len(c.discards) != 0 || len(c.wakes) != 0) {
			t.Fatalf("prefix %d produced output, discarded data, or woke", i+1)
		}
	}
	if c.output.String() != "+PONG\r\n" || len(c.input) != 0 {
		t.Fatalf("output=%q input=%q", c.output.String(), c.input)
	}
}

func TestTrafficBinarySETAndPipeline(t *testing.T) {
	s := newTestServer(t)
	c := openTestConn(s)
	value := bytes.Repeat([]byte{0, 255, '\r', '\n'}, 64<<10)
	key := "key\x00\r\n"
	data := append(protocol.MakeCommand("sEt", key, string(value)), protocol.MakeCommand("GET", key)...)
	original := bytes.Clone(data)
	c.input = data
	traffic(t, s, c)
	if !bytes.Equal(data, original) {
		t.Fatal("input mutated")
	}
	want := append([]byte("+OK\r\n"), protocol.MakeBulkString(string(value))...)
	if !bytes.Equal(c.output.Bytes(), want) {
		t.Fatalf("response length=%d want %d", c.output.Len(), len(want))
	}
	for i := range data {
		data[i] = 'x'
	}
	item, ok := s.DB.Get(key)
	if !ok || !bytes.Equal([]byte(item.Data), value) {
		t.Fatal("stored value aliases overwritten input")
	}
	if c.ctx.(*connectionState).client.Argv != nil {
		t.Fatal("client retains request arguments")
	}
}

func TestTrafficPipelineYields(t *testing.T) {
	s := newTestServer(t)
	s.limits.commands = 3
	c := openTestConn(s)
	c.input = bytes.Repeat(protocol.MakeCommand("PING"), 10)
	traffic(t, s, c)
	if c.output.String() != string(bytes.Repeat(protocol.RespPONG, 3)) || len(c.wakes) != 1 {
		t.Fatalf("first callback output=%q wakes=%d", c.output.String(), len(c.wakes))
	}
	for len(c.input) > 0 {
		waitWake(t, c)
		traffic(t, s, c)
	}
	if !bytes.Equal(c.output.Bytes(), bytes.Repeat(protocol.RespPONG, 10)) {
		t.Fatal("pipeline response loss")
	}
	c.input = []byte("*1\r\n$4\r\nPI")
	traffic(t, s, c)
	if len(c.wakes) != 0 {
		t.Fatal("incomplete frame spins via Wake")
	}
}

func TestSlowCommandOrderingAndUnrelatedClient(t *testing.T) {
	s := newTestServer(t) // Occupy every worker to deterministically pause KEYS.
	releases := make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	s.DB.Store(datastructure.NewItem("key", "value", 0))
	c := openTestConn(s)
	c.input = append(protocol.MakeCommand("KEYS", "*"), protocol.MakeCommand("PING")...)
	traffic(t, s, c)
	if c.output.Len() != 0 || !c.ctx.(*connectionState).running {
		t.Fatal("slow command did not pause its pipeline")
	}
	other := openTestConn(s)
	other.input = protocol.MakeCommand("PING")
	traffic(t, s, other)
	if other.output.String() != "+PONG\r\n" {
		t.Fatal("slow command blocked unrelated client")
	}
	for _, release := range releases {
		close(release)
	}
	waitWake(t, c)
	traffic(t, s, c)
	if c.output.String() != "*1\r\n$3\r\nkey\r\n+PONG\r\n" {
		t.Fatalf("unordered output=%q", c.output.String())
	}
}

func TestWorkerOverloadAndClose(t *testing.T) {
	s := newTestServer(t)
	releases := make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	for i := 0; i < s.limits.queue; i++ {
		s.jobs <- func() {}
	}
	c := openTestConn(s)
	c.input = append(protocol.MakeCommand("KEYS", "*"), protocol.MakeCommand("PING")...)
	traffic(t, s, c)
	if c.output.String() != "-ERR server overloaded\r\n+PONG\r\n" {
		t.Fatalf("overload output=%q", c.output.String())
	}
	for _, release := range releases {
		close(release)
	}
	// Wait for the queue to drain before admitting the next command.
	barrier := make(chan struct{})
	s.jobs <- func() { close(barrier) }
	<-barrier
	releases = make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	c = openTestConn(s)
	c.input = protocol.MakeCommand("KEYS", "*")
	traffic(t, s, c)
	s.OnClose(c, nil)
	for _, release := range releases {
		close(release)
	}
	s.shutdown()
	state := c.ctx.(*connectionState)
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.result != nil || len(c.wakes) != 0 {
		t.Fatal("closed connection retains or receives worker result")
	}
}

func TestTrafficLimitsAndWriteFailure(t *testing.T) {
	for _, test := range []struct {
		name string
		set  func(*Server, *testConn)
	}{
		{"input", func(s *Server, c *testConn) { s.limits.input = 4; c.input = []byte("*1\r\n$") }},
		{"output", func(s *Server, c *testConn) { c.outbound = s.limits.output; c.input = protocol.MakeCommand("PING") }},
		{"write", func(_ *Server, c *testConn) {
			c.writeErr = errors.New("write failed")
			c.input = protocol.MakeCommand("PING")
		}},
		{"protocol", func(_ *Server, c *testConn) { c.input = []byte("*1\r\n:1\r\n") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := newTestServer(t)
			c := openTestConn(s)
			test.set(s, c)
			if s.OnTraffic(c) != gnet.Close {
				t.Fatal("expected close")
			}
		})
	}
}

func TestClientMetadataAndKillArgumentCount(t *testing.T) {
	s := newTestServer(t)
	conn := openTestConn(s)
	c := conn.ctx.(*connectionState).client
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			c.SetName(fmt.Sprint(i))
			c.AddFlag(client.FlagBusy)
			c.RemoveFlag(client.FlagBusy)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			var res bytes.Buffer
			s.clientListSubCommand(c, &res)
			clientInfoSubCommand(c, &res)
		}
	}()
	wg.Wait()
	var res bytes.Buffer
	c.Argv = [][]byte{[]byte("kill"), []byte("id")}
	c.Argc = 2
	s.clientKillSubCommand(c, &res)
	if !bytes.HasPrefix(res.Bytes(), []byte("-ERR")) {
		t.Fatal("CLIENT KILL missing target not rejected")
	}
	c.Argv = [][]byte{[]byte("kill"), []byte("unknown"), []byte("target")}
	c.Argc = 3
	res.Reset()
	s.clientKillSubCommand(c, &res)
	if !bytes.HasPrefix(res.Bytes(), []byte("-ERR")) {
		t.Fatal("CLIENT KILL unknown filter not rejected")
	}
	c.Argv = nil
}

func TestKillClientCompletesResponse(t *testing.T) {
	for _, kind := range []KillClientType{KillClientByID, KillClientByAddr, KillClientByName} {
		s := newTestServer(t)
		conn := openTestConn(s)
		c := conn.ctx.(*connectionState).client
		c.SetName("target")
		var target any = c.ID
		if kind == KillClientByAddr {
			target = c.RemoteAddr
		}
		if kind == KillClientByName {
			target = c.Name()
		}
		var res bytes.Buffer
		s.killClient(&client.Client{}, &res, kind, target)
		if !conn.closed.Load() || res.String() != ":1\r\n" {
			t.Fatalf("closed=%t response=%q", conn.closed.Load(), res.String())
		}
		s.OnClose(conn, nil)
		if _, ok := s.clients.Load(c.ID); ok {
			t.Fatal("closed client remains registered")
		}
	}
}

func TestTrafficPipeSeparatorFragments(t *testing.T) {
	s := newTestServer(t)
	c := openTestConn(s)
	c.input = append(protocol.MakeCommand("PING"), '\r')
	traffic(t, s, c)
	if c.output.String() != "+PONG\r\n" || string(c.input) != "\r" {
		t.Fatalf("output=%q input=%q", c.output.String(), c.input)
	}
	c.input = append(c.input, '\n')
	traffic(t, s, c)
	if string(c.input) != "\r\n" || len(c.wakes) != 0 {
		t.Fatal("fragmented separator lost or spun")
	}
	c.input = append(c.input, protocol.MakeCommand("ECHO", "sentinel")...)
	traffic(t, s, c)
	if c.output.String() != "+PONG\r\n$8\r\nsentinel\r\n" {
		t.Fatalf("output=%q", c.output.String())
	}
}
