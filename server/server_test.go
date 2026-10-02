package server

import (
	"bytes"
	"errors"
	"net"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
)

type pendingWrite struct {
	data     []byte
	callback gnet.AsyncCallback
}

type delayedConn struct {
	gnet.Conn
	writes []pendingWrite
	err    error
	closed bool
}

func (c *delayedConn) AsyncWrite(data []byte, callback gnet.AsyncCallback) error {
	c.writes = append(c.writes, pendingWrite{data, callback})
	return c.err
}

func (c *delayedConn) Context() any {
	panic("worker must not read connection context")
}

func (c *delayedConn) Close() error {
	c.closed = true
	return nil
}

func (c *delayedConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

func newTestServer() *Server {
	s := &Server{DB: datastructure.NewMap()}
	s.parserPool.New = func() any {
		br := bytes.NewReader(nil)
		return &parser{br: br, pr: protocol.NewReader(br)}
	}
	return s
}

func TestHandleRetainsPendingResponses(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "queued"},
		{name: "queued with submission error", err: errors.New("wakeup failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer()
			conn := &delayedConn{err: tc.err}
			c := &client.Client{Conn: conn, DB: s.DB}
			s.DB.Store(datastructure.NewItem("key", "value", 0))
			for i := 0; i < 100; i++ {
				s.handle(protocol.MakeCommand("PING"), c)
				s.handle(protocol.MakeCommand("GET", "key"), c)
			}
			if len(conn.writes) != 200 {
				t.Fatalf("writes = %d, want 200", len(conn.writes))
			}
			for i, write := range conn.writes {
				want := "+PONG\r\n"
				if i%2 == 1 {
					want = "$5\r\nvalue\r\n"
				}
				if string(write.data) != want {
					t.Errorf("pending write %d = %q, want %q", i, write.data, want)
				}
				if write.callback == nil {
					t.Fatal("missing write completion callback")
				}
				if err := write.callback(conn, nil); err != nil {
					t.Fatal(err)
				}
			}
			if c.Argv != nil {
				t.Fatal("client retains pooled arguments")
			}
		})
	}
}

func TestKillClientCompletesResponse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kind   KillClientType
		target any
	}{
		{"id", KillClientByID, int64(42)},
		{"address", KillClientByAddr, "127.0.0.1:12345"},
		{"name", KillClientByName, "target"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{}
			conn := &delayedConn{}
			s.clients.Store(int64(42), &client.Client{ID: 42, Name: "target", Conn: conn})
			var response bytes.Buffer
			s.killClient(&client.Client{}, &response, tc.kind, tc.target)
			if !conn.closed || response.String() != ":1\r\n" {
				t.Fatalf("closed = %t, response = %q", conn.closed, response.String())
			}
			if _, ok := s.clients.Load(int64(42)); ok {
				t.Fatal("killed client remains registered")
			}
		})
	}
}
