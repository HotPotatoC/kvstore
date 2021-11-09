package protocol_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func TestReader(t *testing.T) {
	tc := []struct {
		name string
		st   []byte
		exp  interface{}
		err  error
	}{
		{name: "SimpleString", st: []byte("+OK\r\n"), exp: "OK", err: nil},
		{name: "SimpleString", st: []byte("+PONG\r\n"), exp: "PONG", err: nil},
		{name: "Error", st: []byte("-ERR Generic error\r\n"), exp: "ERR Generic error", err: nil},
		{name: "Integer", st: []byte(":123\r\n"), exp: 123, err: nil},
		{name: "Negative Integer", st: []byte(":-123\r\n"), exp: -123, err: nil},
		{name: "BulkString", st: []byte("$3\r\nfoo\r\n"), exp: []byte("foo"), err: nil},
		{name: "Empty BulkString", st: []byte("$0\r\n\r\n"), exp: []byte(""), err: nil},
		{name: "Array", st: []byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"), exp: []interface{}{[]byte("foo"), []byte("bar")}, err: nil},
		{name: "Array with nil", st: []byte("*2\r\n$3\r\nfoo\r\n$-1\r\n"), exp: []interface{}{[]byte("foo"), nil}, err: nil},
		{name: "Array with nil and empty bulk string", st: []byte("*2\r\n$3\r\nfoo\r\n$0\r\n\r\n"), exp: []interface{}{[]byte("foo"), []byte("")}, err: nil},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			r := protocol.NewReader(bytes.NewBuffer(tt.st))
			got, err := r.ReadObject()
			if err != tt.err {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}

			if !reflect.DeepEqual(got, tt.exp) {
				t.Errorf("Expected %#v, got %#v", tt.exp, got)
			}
		})
	}
}
