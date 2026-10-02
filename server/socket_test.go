package server

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
	"github.com/spf13/viper"
)

type socketServer struct {
	*Server
	ready chan gnet.Engine
}

func (s *socketServer) OnBoot(engine gnet.Engine) gnet.Action {
	action := s.Server.OnBoot(engine)
	s.ready <- engine
	return action
}
func startSocketServer(t *testing.T) (*Server, string) {
	t.Helper()
	s := newTestServer(t)
	s.limits.commands = 7
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	wrapper := &socketServer{Server: s, ready: make(chan gnet.Engine, 1)}
	done := make(chan error, 1)
	go func() { done <- gnet.Run(wrapper, "tcp://"+addr, gnet.WithNumEventLoop(1)) }()
	var engine gnet.Engine
	select {
	case engine = <-wrapper.ready:
	case err := <-done:
		t.Fatalf("server startup: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("server startup timed out")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("stop server: %v", err)
		}
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("server exit: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("server did not exit")
		}
	})
	return s, addr
}
func dialSocket(t *testing.T, addr string) net.Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	c.SetDeadline(time.Now().Add(5 * time.Second))
	return c
}
func sendSocket(t *testing.T, c net.Conn, data []byte) {
	t.Helper()
	if _, err := io.Copy(c, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
}
func readSocket(t *testing.T, c net.Conn, want []byte) {
	t.Helper()
	got := make([]byte, len(want))
	if _, err := io.ReadFull(c, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("response length=%d want=%d; prefixes=%q / %q", len(got), len(want), got[:min(80, len(got))], want[:min(80, len(want))])
	}
}

func TestSocketFragmentedBinaryAndLargePipeline(t *testing.T) {
	_, addr := startSocketServer(t)
	c := dialSocket(t, addr)
	ping := protocol.MakeCommand("PING")
	split := len(ping) - 3
	sendSocket(t, c, ping[:split])
	c.SetReadDeadline(time.Now().Add(40 * time.Millisecond))
	var probe [1]byte
	if n, err := c.Read(probe[:]); n != 0 || err == nil {
		t.Fatalf("fragment generated response: n=%d err=%v", n, err)
	}
	c.SetDeadline(time.Now().Add(5 * time.Second))
	sendSocket(t, c, ping[split:])
	readSocket(t, c, protocol.RespPONG)
	value := bytes.Repeat([]byte{0, 255, '\r', '\n'}, 64<<10)
	set := protocol.MakeCommand("SET", "binary", string(value))
	sendSocket(t, c, set[:1025])
	sendSocket(t, c, set[1025:])
	readSocket(t, c, protocol.RespOK)
	sendSocket(t, c, protocol.MakeCommand("GET", "binary"))
	readSocket(t, c, protocol.MakeBulkString(string(value)))
	pipeline := bytes.Repeat(protocol.MakeCommand("PING"), 2048)
	pipeline = append(pipeline, '\r', '\n')
	pipeline = append(pipeline, protocol.MakeCommand("ECHO", "sentinel\x00")...)
	sendSocket(t, c, pipeline)
	want := append(bytes.Repeat(protocol.RespPONG, 2048), protocol.MakeBulkString("sentinel\x00")...)
	readSocket(t, c, want)
}

func TestSocketSlowOrderingAndClose(t *testing.T) {
	s, addr := startSocketServer(t)
	s.DB.Store(datastructure.NewItem("key", "value", 0))
	releases := make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	slow := dialSocket(t, addr)
	sendSocket(t, slow, append(protocol.MakeCommand("KEYS", "*"), protocol.MakeCommand("PING")...))
	fast := dialSocket(t, addr)
	sendSocket(t, fast, protocol.MakeCommand("PING"))
	readSocket(t, fast, protocol.RespPONG)
	slow.SetReadDeadline(time.Now().Add(40 * time.Millisecond))
	var probe [1]byte
	if n, err := slow.Read(probe[:]); n != 0 || err == nil {
		t.Fatalf("slow pipeline replied before completion: n=%d err=%v", n, err)
	}
	slow.SetDeadline(time.Now().Add(5 * time.Second))
	for _, release := range releases {
		close(release)
	}
	readSocket(t, slow, []byte("*1\r\n$3\r\nkey\r\n+PONG\r\n"))
	// A queued command may finish after its socket closes. The server must release
	// its result and continue serving the unrelated connection.
	releases = make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	sendSocket(t, slow, protocol.MakeCommand("KEYS", "*"))
	slow.Close()
	for _, release := range releases {
		close(release)
	}
	sendSocket(t, fast, protocol.MakeCommand("PING"))
	readSocket(t, fast, protocol.RespPONG)
}

func TestDefaultLifecycleAndWorkerDrain(t *testing.T) {
	old := viper.Get("database.path")
	viper.Set("database.path", filepath.Join(t.TempDir(), "dump.kvsdb"))
	t.Cleanup(func() { viper.Set("database.path", old) })
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.shutdown)
	if s.limits.loops != 4 || s.limits.workers != 4 || s.jobs == nil {
		t.Fatalf("defaults=%+v", s.limits)
	}
	release := blockWorker(t, s)
	done := make(chan struct{})
	go func() { s.shutdown(); close(done) }()
	select {
	case <-done:
		t.Fatal("shutdown did not wait for worker")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not drain worker")
	}
	state := &connectionState{client: &client.Client{Conn: &testConn{wakes: make(chan struct{}, 1)}}}
	if s.submitSlow(state, [][]byte{[]byte("KEYS"), []byte("*")}, commandOther) {
		t.Fatal("shutdown accepted new worker job")
	}
}

func TestShutdownCallbackDefersPersistence(t *testing.T) {
	old := viper.Get("database.path")
	path := filepath.Join(t.TempDir(), "dump.kvsdb")
	viper.Set("database.path", path)
	t.Cleanup(func() { viper.Set("database.path", old) })
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.shutdown)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s.DB.Store(datastructure.NewItem("before", "value", 0))
	s.OnShutdown(gnet.Engine{})
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("OnShutdown persisted before event loops exited")
	}
	// Simulate the last already-admitted inline write while loops finish.
	s.DB.Store(datastructure.NewItem("last", "acknowledged", 0))
	s.shutdown()
	db, err := disk.OpenKVSDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	snapshot, err := db.Read()
	if err != nil {
		t.Fatal(err)
	}
	item, ok := snapshot.Get("last")
	if !ok || item.Data != "acknowledged" {
		t.Fatal("final inline write missing from persisted snapshot")
	}
}

type delayedBootServer struct {
	*Server
	entered, resume chan struct{}
}

func (s *delayedBootServer) OnBoot(engine gnet.Engine) gnet.Action {
	close(s.entered)
	<-s.resume
	return s.Server.OnBoot(engine)
}

func TestStopBeforeOnBoot(t *testing.T) {
	s := newTestServer(t)
	wrapper := &delayedBootServer{Server: s, entered: make(chan struct{}), resume: make(chan struct{})}
	t.Cleanup(func() {
		select {
		case <-wrapper.resume:
		default:
			close(wrapper.resume)
		}
	})
	done := make(chan error, 1)
	go func() { done <- gnet.Run(wrapper, "tcp://127.0.0.1:0") }()
	select {
	case <-wrapper.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("OnBoot not called")
	}
	stopped := make(chan struct{})
	go func() { s.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop before OnBoot hung")
	}
	close(wrapper.resume)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("stopped server started or hung")
	}
	s.engMu.Lock()
	err := s.eng.Validate()
	s.engMu.Unlock()
	if err == nil {
		t.Fatal("OnBoot published an engine that never starts")
	}
}

func TestRunWithoutAddressesCleansUp(t *testing.T) {
	old := viper.Get("server.addrs")
	viper.Set("server.addrs", []string{})
	t.Cleanup(func() { viper.Set("server.addrs", old) })
	s := newTestServer(t)
	if err := s.Run(); err == nil {
		t.Fatal("empty bind addresses accepted")
	}
	if s.submitSlow(&connectionState{}, [][]byte{[]byte("KEYS"), []byte("*")}, commandOther) {
		t.Fatal("failed Run left worker admission open")
	}
}
