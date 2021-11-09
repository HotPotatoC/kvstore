package protocol_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func TestWriter_WriteCommand(t *testing.T) {
	tc := []struct {
		name string
		args []string
		exp  []byte
	}{
		{name: "set", args: []string{"set", "key", "value"}, exp: []byte("*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")},
		{name: "get", args: []string{"get", "key"}, exp: []byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n")},
		{name: "del", args: []string{"del", "key"}, exp: []byte("*2\r\n$3\r\ndel\r\n$3\r\nkey\r\n")},
		{name: "incr", args: []string{"incr", "key"}, exp: []byte("*2\r\n$4\r\nincr\r\n$3\r\nkey\r\n")},
		{name: "decr", args: []string{"decr", "key"}, exp: []byte("*2\r\n$4\r\ndecr\r\n$3\r\nkey\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := protocol.NewWriter(&buf)
			w.WriteCommand(tt.args...)
			if !bytes.Equal(buf.Bytes(), tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), buf.String())
			}
		})
	}
}

func TestWriter_WriteSimpleString(t *testing.T) {
	tc := []struct {
		name string
		args string
		exp  []byte
	}{
		{name: "OK", args: "OK", exp: []byte("+OK\r\n")},
		{name: "PONG", args: "PONG", exp: []byte("+PONG\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := protocol.NewWriter(&buf)
			w.WriteSimpleString(tt.args)
			if !bytes.Equal(buf.Bytes(), tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), buf.String())
			}
		})
	}
}

func TestWriter_WriteError(t *testing.T) {
	tc := []struct {
		name string
		args string
		exp  []byte
	}{
		{name: "ERR", args: "ERR", exp: []byte("-ERR\r\n")},
		{name: "ERR with message", args: "ERR message", exp: []byte("-ERR message\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := protocol.NewWriter(&buf)
			w.WriteError(tt.args)
			if !bytes.Equal(buf.Bytes(), tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), buf.String())
			}
		})
	}
}

func TestWriter_WriteInteger(t *testing.T) {
	tc := []struct {
		name string
		args int64
		exp  []byte
	}{
		{name: "0", args: 0, exp: []byte(":0\r\n")},
		{name: "1", args: 1, exp: []byte(":1\r\n")},
		{name: "10", args: 10, exp: []byte(":10\r\n")},
		{name: "100", args: 100, exp: []byte(":100\r\n")},
		{name: "1000", args: 1000, exp: []byte(":1000\r\n")},
		{name: "10000", args: 10000, exp: []byte(":10000\r\n")},
		{name: "100000", args: 100000, exp: []byte(":100000\r\n")},
		{name: "1000000", args: 1000000, exp: []byte(":1000000\r\n")},
		{name: "10000000", args: 10000000, exp: []byte(":10000000\r\n")},
		{name: "-1", args: -1, exp: []byte(":-1\r\n")},
		{name: "-10", args: -10, exp: []byte(":-10\r\n")},
		{name: "-100", args: -100, exp: []byte(":-100\r\n")},
		{name: "-1000", args: -1000, exp: []byte(":-1000\r\n")},
		{name: "-10000", args: -10000, exp: []byte(":-10000\r\n")},
		{name: "-100000", args: -100000, exp: []byte(":-100000\r\n")},
		{name: "-1000000", args: -1000000, exp: []byte(":-1000000\r\n")},
		{name: "-10000000", args: -10000000, exp: []byte(":-10000000\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := protocol.NewWriter(&buf)
			w.WriteInteger(tt.args)
			if !bytes.Equal(buf.Bytes(), tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), buf.String())
			}
		})
	}
}

func TestWriter_WriteBulkString(t *testing.T) {
	tc := []struct {
		name string
		args string
		exp  []byte
	}{
		{name: "empty", args: "", exp: []byte("$0\r\n\r\n")},
		{name: "1", args: "1", exp: []byte("$1\r\n1\r\n")},
		{name: "10", args: "1234567890", exp: []byte("$10\r\n1234567890\r\n")},
		{name: "Hello World", args: "Hello World", exp: []byte("$11\r\nHello World\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := protocol.NewWriter(&buf)
			w.WriteBulkString(tt.args)
			if !bytes.Equal(buf.Bytes(), tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), buf.String())
			}
		})
	}
}
