package server

import (
	"bytes"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
)

func TestCommandClassification(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		kind commandKind
		slow bool
	}{
		{"get", []string{"GeT", "key"}, commandGet, false},
		{"set", []string{"sEt", "key", "value"}, commandSet, false},
		{"ping", []string{"PiNg"}, commandPing, false},
		{"literal del", []string{"dEl", "key"}, commandDel, false},
		{"glob del", []string{"DeL", "key*"}, commandDel, false},
		{"escaped del", []string{"DEL", "key\\literal"}, commandDel, false},
		{"missing del", []string{"Del"}, commandDel, false},
		{"keys", []string{"KeYs", "*"}, commandOther, true},
		{"flush", []string{"FlUsHaLl"}, commandOther, true},
		{"client list", []string{"cLiEnT", "LiSt"}, commandOther, true},
		{"client kill", []string{"ClIeNt", "kIlL", "id", "2"}, commandOther, true},
		{"client name", []string{"CLIENT", "GeTnAmE"}, commandOther, false},
		{"unknown", []string{"NOPE"}, commandOther, false},
		{"near get", []string{"gex"}, commandOther, false},
		{"empty", nil, commandOther, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := make([][]byte, len(tc.args))
			for i, arg := range tc.args {
				args[i] = []byte(arg)
			}
			kind := classifyCommand(args)
			if kind != tc.kind || slowCommand(args, kind) != tc.slow {
				t.Fatalf("kind=%v slow=%t", kind, slowCommand(args, kind))
			}
		})
	}
}

func TestTrafficMixedCaseDispatchAndUnknown(t *testing.T) {
	s := newTestServer(t)
	c := openTestConn(s)
	var data []byte
	for _, args := range [][]string{{"sEt", "key", "value"}, {"gEt", "key"}, {"dEl", "key"}, {"PiNg"}, {"nOpE"}, {"cLiEnT", "SeTnAmE", "name"}, {"ClIeNt", "GeTnAmE"}} {
		data = append(data, protocol.MakeCommand(args...)...)
	}
	before := bytes.Clone(data)
	c.input = data
	traffic(t, s, c)
	want := "+OK\r\n$5\r\nvalue\r\n:1\r\n+PONG\r\n-ERR unknown command 'nope'\r\n+OK\r\n$4\r\nname\r\n"
	if c.output.String() != want {
		t.Fatalf("output=%q want %q", c.output.String(), want)
	}
	if !bytes.Equal(before, data) {
		t.Fatal("dispatch mutated input")
	}
}

func TestTrafficShutdownReadDoesNotLockWorkerAdmission(t *testing.T) {
	s := newTestServer(t)
	c := openTestConn(s)
	c.input = protocol.MakeCommand("PING")
	// A slow admission/close lock must not serialize the inline command path.
	s.workerMu.Lock()
	done := make(chan gnet.Action, 1)
	go func() { done <- s.OnTraffic(c) }()
	var action gnet.Action
	select {
	case action = <-done:
		s.workerMu.Unlock()
	case <-time.After(time.Second):
		s.workerMu.Unlock()
		<-done
		t.Fatal("inline traffic waited on worker admission lock")
	}
	if action != gnet.None || c.output.String() != "+PONG\r\n" {
		t.Fatalf("action=%v output=%q", action, c.output.String())
	}
	s.stopping.Store(true)
	c.input = protocol.MakeCommand("SET", "after-stop", "value")
	if s.OnTraffic(c) != gnet.Close {
		t.Fatal("traffic admitted after stopping")
	}
	if s.DB.Exists("after-stop") {
		t.Fatal("stopping traffic wrote a key")
	}
}

func TestKEYSKillPreservesFlagsAndOrder(t *testing.T) {
	s := newTestServer(t)
	releases := make([]chan struct{}, s.limits.workers)
	for i := range releases {
		releases[i] = blockWorker(t, s)
	}
	s.DB.Store(datastructure.NewItem("key:1", "value", 0))
	conn := openTestConn(s)
	state := conn.ctx.(*connectionState)
	c := state.client
	c.AddFlag(client.FlagReadOnly)
	conn.input = append(protocol.MakeCommand("KEYS", "key:*"), protocol.MakeCommand("PING")...)
	traffic(t, s, conn)
	if !state.running || !c.HasFlag(client.FlagBusy) || c.HasFlag(client.FlagNone) {
		t.Fatal("offload did not publish busy state")
	}
	var killerResponse bytes.Buffer
	s.killClient(&client.Client{}, &killerResponse, KillClientByID, c.ID)
	if killerResponse.String() != ":1\r\n" || conn.closed.Load() || !c.HasFlag(client.FlagCloseASAP) || !c.HasFlag(client.FlagReadOnly) {
		t.Fatal("concurrent kill lost flags or closed busy client before its response")
	}
	waitWake(t, conn)
	traffic(t, s, conn)
	if conn.output.Len() != 0 {
		t.Fatal("pending slow command produced an early response")
	}
	for _, release := range releases {
		close(release)
	}
	waitWake(t, conn)
	if action := s.OnTraffic(conn); action != gnet.Close {
		t.Fatalf("close request action=%v", action)
	}
	if conn.output.String() != "*1\r\n$5\r\nkey:1\r\n" {
		t.Fatalf("slow response or pipeline order lost: %q", conn.output.String())
	}
	if !c.HasFlag(client.FlagNone) || c.HasFlag(client.FlagBusy) || !c.HasFlag(client.FlagCloseASAP) || !c.HasFlag(client.FlagReadOnly) {
		t.Fatal("completion transition lost command state or options")
	}
}
