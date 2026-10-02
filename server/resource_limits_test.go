package server

import (
	"bytes"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
)

func TestFlushFailureKeepsPipelineAligned(t *testing.T) {
	s := newTestServer(t)
	db, err := disk.OpenKVSDB(filepath.Join(t.TempDir(), "dump.kvsdb"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	s.kvsDB = db
	c := openTestConn(s)
	c.input = append(protocol.MakeCommand("FLUSHALL"), protocol.MakeCommand("PING")...)
	traffic(t, s, c)
	waitWake(t, c)
	traffic(t, s, c)
	got := c.output.Bytes()
	if !bytes.HasPrefix(got, []byte("-ERR ")) || bytes.Count(got, []byte("\r\n")) != 2 || !bytes.HasSuffix(got, protocol.RespPONG) {
		t.Fatalf("unexpected pipeline replies: %q", got)
	}
	s.kvsDB = nil // Already closed; avoid shutdown snapshot attempt.
}

func TestMaxClientsConcurrentAdmissionAndRelease(t *testing.T) {
	s := newTestServer(t)
	s.limits.maxClients = 5
	var accepted atomic.Int64
	var wg sync.WaitGroup
	conns := make([]*testConn, 40)
	for i := range conns {
		conns[i] = &testConn{}
		wg.Add(1)
		go func(c *testConn) {
			defer wg.Done()
			out, action := s.OnOpen(c)
			if action == gnet.None {
				accepted.Add(1)
			} else if action != gnet.Close || string(out) != "-ERR max number of clients reached\r\n" || c.Context() != nil {
				t.Errorf("unexpected rejection: %q %v", out, action)
			}
		}(conns[i])
	}
	wg.Wait()
	if accepted.Load() != 5 || s.activeClients.Load() != 5 {
		t.Fatalf("accepted=%d active=%d", accepted.Load(), s.activeClients.Load())
	}
	for _, c := range conns {
		s.OnClose(c, nil)
	}
	if s.activeClients.Load() != 0 {
		t.Fatal("connection slots leaked")
	}
	c := &testConn{}
	if _, action := s.OnOpen(c); action != gnet.None {
		t.Fatal("released slot unavailable")
	}
	s.OnClose(c, nil)
	s.OnClose(c, nil)
	if s.activeClients.Load() != 0 {
		t.Fatal("duplicate close changed active count")
	}
}

func TestMaxClientsZeroDisablesLimit(t *testing.T) {
	s := newTestServer(t)
	s.limits.maxClients = 0
	for i := 0; i < 20; i++ {
		c := &testConn{}
		if _, action := s.OnOpen(c); action != gnet.None {
			t.Fatal("unlimited connection rejected")
		}
		s.OnClose(c, nil)
	}
}
