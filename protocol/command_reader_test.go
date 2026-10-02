package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"testing"
)

func TestParseCommandBinaryAndReuse(t *testing.T) {
	want := [][]byte{[]byte("SET"), {0, '\r', '\n', 255}, {}}
	data := MakeCommand("SET", string(want[1]), "")
	original := bytes.Clone(data)
	scratch := make([][]byte, 0, 8)
	args, consumed, err := ParseCommand(data, scratch)
	if err != nil || consumed != len(data) || len(args) != len(want) {
		t.Fatalf("args=%q consumed=%d err=%v", args, consumed, err)
	}
	for i := range want {
		if !bytes.Equal(args[i], want[i]) {
			t.Fatalf("arg %d = %q, want %q", i, args[i], want[i])
		}
		if cap(args[i]) != len(args[i]) {
			t.Fatalf("arg %d has spare capacity", i)
		}
	}
	if &args[0] != &scratch[:cap(scratch)][0] {
		t.Fatal("scratch not reused")
	}
	if !bytes.Equal(data, original) {
		t.Fatal("parser mutated input")
	}
	args[1][0] = 1
	if bytes.Equal(data, original) {
		t.Fatal("arguments do not alias input")
	}
}

func TestParseCommandFragmented(t *testing.T) {
	for _, data := range [][]byte{MakeCommand("SET", "key", "a\x00b\r\nc"), MakeCommand(""), []byte("*0\r\n")} {
		for n := 0; n < len(data); n++ {
			_, consumed, err := ParseCommand(data[:n], make([][]byte, 0, 8))
			if !errors.Is(err, io.ErrUnexpectedEOF) || consumed != 0 {
				t.Fatalf("prefix %d of %q: consumed=%d err=%v", n, data, consumed, err)
			}
		}
		args, consumed, err := ParseCommand(data, nil)
		if err != nil || consumed != len(data) {
			t.Fatalf("full frame args=%q consumed=%d err=%v", args, consumed, err)
		}
	}
}

func TestParseCommandConcatenated(t *testing.T) {
	first, second := MakeCommand("PING"), MakeCommand("GET", "key")
	data := append(bytes.Clone(first), second...)
	args, consumed, err := ParseCommand(data, nil)
	if err != nil || consumed != len(first) || string(args[0]) != "PING" {
		t.Fatalf("first args=%q consumed=%d err=%v", args, consumed, err)
	}
	args, used, err := ParseCommand(data[consumed:], args)
	if err != nil || used != len(second) || string(args[0]) != "GET" || string(args[1]) != "key" {
		t.Fatalf("second args=%q consumed=%d err=%v", args, used, err)
	}
}

func TestParseCommandRejectsMalformed(t *testing.T) {
	for _, data := range []string{
		"+PING\r\n", "*-1\r\n", "*+1\r\n", "*\r\n", "*1\n", "*1\rx", "*x\r\n",
		"*1\r\n$-1\r\n", "*1\r\n$-2\r\n", "*1\r\n$\r\n", "*1\r\n:1\r\n", "*1\r\n*0\r\n",
		"*1\r\n$1\r\nx\ny", "*1\r\n$1\r\nx\ry", "*1\r\n$1\nx\r\n",
		"*" + strconv.FormatUint(uint64(^uint(0)>>1), 10) + "0\r\n",
		"*1\r\n$" + strconv.FormatUint(uint64(^uint(0)>>1), 10) + "0\r\n",
	} {
		t.Run(fmt.Sprintf("%q", data), func(t *testing.T) {
			_, consumed, err := ParseCommand([]byte(data), nil)
			if err == nil || errors.Is(err, io.ErrUnexpectedEOF) || consumed != 0 {
				t.Fatalf("consumed=%d err=%v", consumed, err)
			}
		})
	}
}

var perfArgs [][]byte

func BenchmarkParseCommandSET(b *testing.B) {
	for _, size := range []int{32, 1024, 65536} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			payload := MakeCommand("SET", "key:0000000123", string(bytes.Repeat([]byte("x"), size)))
			args := make([][]byte, 0, 8)
			b.ReportAllocs()
			b.SetBytes(int64(len(payload)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var consumed int
				var err error
				args, consumed, err = ParseCommand(payload, args)
				if err != nil {
					b.Fatal(err)
				}
				if len(args) != 3 || consumed != len(payload) {
					b.Fatal("bad command")
				}
				perfArgs = args
			}
		})
	}
}

func TestParseCommandLimits(t *testing.T) {
	for _, input := range []string{"*1025\r\n", "*1\r\n$8388609\r\n"} {
		_, n, err := ParseCommand([]byte(input), nil)
		if !errors.Is(err, ErrCommandTooLarge) || n != 0 {
			t.Fatalf("input=%q consumed=%d err=%v", input, n, err)
		}
	}
}

func TestParseCommandPipeSeparator(t *testing.T) {
	data := append([]byte("\r\n\r\n"), MakeCommand("ECHO", "sentinel")...)
	for n := 0; n < len(data); n++ {
		_, consumed, err := ParseCommand(data[:n], nil)
		if !errors.Is(err, io.ErrUnexpectedEOF) || consumed != 0 {
			t.Fatalf("prefix %d consumed=%d err=%v", n, consumed, err)
		}
	}
	args, n, err := ParseCommand(data, nil)
	if err != nil || n != len(data) || len(args) != 2 || string(args[1]) != "sentinel" {
		t.Fatalf("args=%q consumed=%d err=%v", args, n, err)
	}
}

func TestParseCommandTotalLengthLimit(t *testing.T) {
	// A second maximum-size bulk would exceed the total frame limit. Reject its
	// declaration without waiting for or allocating its body.
	data := []byte("*2\r\n$8388608\r\n")
	data = append(data, bytes.Repeat([]byte("x"), MaxBulkLength)...)
	data = append(data, []byte("\r\n$8388608\r\n")...)
	_, n, err := ParseCommand(data, nil)
	if !errors.Is(err, ErrCommandTooLarge) || n != 0 {
		t.Fatalf("consumed=%d err=%v", n, err)
	}
	// Zero arguments still have to obey the encoded frame length limit.
	header := append([]byte("*"), bytes.Repeat([]byte("0"), MaxCommandLength)...)
	header = append(header, '\r', '\n')
	_, n, err = ParseCommand(header, nil)
	if !errors.Is(err, ErrCommandTooLarge) || n != 0 {
		t.Fatalf("zero-array header consumed=%d err=%v", n, err)
	}
	separators := bytes.Repeat([]byte("\r\n"), MaxCommandLength/2)
	_, n, err = ParseCommand(separators, nil)
	if !errors.Is(err, ErrCommandTooLarge) || n != 0 {
		t.Fatalf("separators consumed=%d err=%v", n, err)
	}
}
