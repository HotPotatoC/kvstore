package command_test

import (
	"bytes"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/command"
)

func equal(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}

	return true
}

func TestWrapArgsFromQuotes(t *testing.T) {
	tc := []struct {
		name string
		args [][]byte
		exp  [][]byte
	}{
		{
			name: "Hello World",
			args: [][]byte{{byte('"')}, []byte("Hello"), {' '}, []byte("World"), {'"'}},
			exp:  [][]byte{[]byte("Hello World")},
		},
		{
			name: `{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}`,
			args: [][]byte{{byte('"')}, []byte("{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}"), {'"'}},
			exp:  [][]byte{[]byte("{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}")},
		},
		{
			name: `[\"Hello\",\"World\"]`,
			args: [][]byte{{byte('"')}, []byte("[\"Hello\",\"World\"]"), {'"'}},
			exp:  [][]byte{[]byte("[\"Hello\",\"World\"]")},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			got := command.WrapArgsFromQuotes(tt.args)
			if !equal(got, tt.exp) {
				t.Errorf("expected %#v, got %#v", tt.exp, got)
			}
		})
	}
}
