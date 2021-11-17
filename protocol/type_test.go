package protocol_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func TestWriter_MakeCommand(t *testing.T) {
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
			res := protocol.MakeCommand(tt.args...)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}

func TestWriter_MakeSimpleString(t *testing.T) {
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
			res := protocol.MakeSimpleString(tt.args)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}

func TestWriter_MakeError(t *testing.T) {
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
			res := protocol.MakeError(tt.args)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}

func TestWriter_MakeInteger(t *testing.T) {
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
			res := protocol.MakeInteger(tt.args)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}

func Test_MakeBool(t *testing.T) {
	tc := []struct {
		name string
		args bool
		exp  []byte
	}{
		{name: "true", args: true, exp: []byte(":1\r\n")},
		{name: "false", args: false, exp: []byte(":0\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			res := protocol.MakeBool(tt.args)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}

func TestWriter_MakeBulkString(t *testing.T) {
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
			res := protocol.MakeBulkString(tt.args)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}

func TestWriter_MakeArray(t *testing.T) {
	tc := []struct {
		name string
		args [][]byte
		exp  []byte
	}{
		{name: "empty", args: [][]byte{}, exp: []byte("*0\r\n")},
		{name: "1", args: [][]byte{protocol.MakeBulkString("1")}, exp: []byte("*1\r\n$1\r\n1\r\n")},
		{name: "2", args: [][]byte{protocol.MakeBulkString("1"), protocol.MakeBulkString("2")}, exp: []byte("*2\r\n$1\r\n1\r\n$1\r\n2\r\n")},
		{name: "3", args: [][]byte{protocol.MakeBulkString("1"), protocol.MakeBulkString("2"), protocol.MakeBulkString("3")}, exp: []byte("*3\r\n$1\r\n1\r\n$1\r\n2\r\n$1\r\n3\r\n")},
		{name: "SET key value", args: [][]byte{protocol.MakeBulkString("SET"), protocol.MakeBulkString("key"), protocol.MakeBulkString("value")}, exp: []byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			res := protocol.MakeArray(tt.args...)
			if !bytes.Equal(res, tt.exp) {
				t.Errorf("expected %#v, got %#v", string(tt.exp), string(res))
			}
		})
	}
}
