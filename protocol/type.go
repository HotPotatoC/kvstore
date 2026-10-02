package protocol

import (
	"bytes"
	"strconv"
)

const (
	// SimpleString represents a string.
	SimpleString byte = '+'
	// Error represents an error.
	Error byte = '-'
	// Integer represents an integer.
	Integer byte = ':'
	// BulkString represents a bulk string.
	BulkString byte = '$'
	// Array represents an array.
	Array byte = '*'
)

var (
	// CRLF represents the end of a line.
	CRLF      = []byte{'\r', '\n'}
	nullBytes = []byte("$-1\r\n")

	// RespOK is the pre-computed RESP OK response.
	RespOK = []byte("+OK\r\n")
	// RespPONG is the pre-computed RESP PONG response.
	RespPONG = []byte("+PONG\r\n")
	// RespNull is the pre-computed RESP null bulk string response.
	RespNull = nullBytes
)

// WriteSimpleString writes a simple string RESP response to w.
func WriteSimpleString(w *bytes.Buffer, s string) {
	w.WriteByte(SimpleString)
	w.WriteString(s)
	w.Write(CRLF)
}

// WriteError writes an error RESP response to w.
func WriteError(w *bytes.Buffer, s string) {
	w.WriteByte(Error)
	w.WriteString(s)
	w.Write(CRLF)
}

// WriteInteger writes an integer RESP response to w.
func WriteInteger(w *bytes.Buffer, i int64) {
	w.WriteByte(Integer)
	var buf [20]byte
	w.Write(strconv.AppendInt(buf[:0], i, 10))
	w.Write(CRLF)
}

// WriteBulkString writes a bulk string RESP response to w.
func WriteBulkString(w *bytes.Buffer, s string) {
	w.WriteByte(BulkString)
	var buf [20]byte
	w.Write(strconv.AppendInt(buf[:0], int64(len(s)), 10))
	w.Write(CRLF)
	w.WriteString(s)
	w.Write(CRLF)
}

// WriteNull writes a null bulk string RESP response to w.
func WriteNull(w *bytes.Buffer) {
	w.Write(nullBytes)
}

// MakeCommand creates a command protocol object.
func MakeCommand(args ...string) []byte {
	totalSize := 1 + numDigits(int64(len(args))) + 2 // *N\r\n
	for _, arg := range args {
		totalSize += 1 + numDigits(int64(len(arg))) + 2 + len(arg) + 2 // $L\r\nARG\r\n
	}

	b := make([]byte, 0, totalSize)
	b = append(b, Array)
	b = strconv.AppendInt(b, int64(len(args)), 10)
	b = append(b, CRLF...)
	for _, arg := range args {
		b = append(b, BulkString)
		b = strconv.AppendInt(b, int64(len(arg)), 10)
		b = append(b, CRLF...)
		b = append(b, arg...)
		b = append(b, CRLF...)
	}

	return b
}

// MakeSimpleString creates a simple string protocol object.
func MakeSimpleString(s string) []byte {
	size := 1 + len(s) + 2 // +s\r\n
	b := make([]byte, 0, size)
	b = append(b, SimpleString)
	b = append(b, s...)
	b = append(b, CRLF...)
	return b
}

// MakeError creates an error protocol object.
func MakeError(s string) []byte {
	size := 1 + len(s) + 2 // -s\r\n
	b := make([]byte, 0, size)
	b = append(b, Error)
	b = append(b, s...)
	b = append(b, CRLF...)
	return b
}

// MakeInteger creates an integer protocol object.
func MakeInteger(i int64) []byte {
	size := 1 + numDigits(i) + 2 // :i\r\n
	if i < 0 {
		size++ // For the '-' sign
	}
	b := make([]byte, 0, size)
	b = append(b, Integer)
	b = strconv.AppendInt(b, i, 10)
	b = append(b, CRLF...)
	return b
}

var (
	respTrue  = []byte(":1\r\n")
	respFalse = []byte(":0\r\n")
)

// MakeBool creates a bool protocol object. (basically an integer with value 1 or 0)
func MakeBool(b bool) []byte {
	if b {
		return respTrue
	}
	return respFalse
}

// MakeBulkString creates a bulk string protocol object.
func MakeBulkString(s string) []byte {
	size := 1 + numDigits(int64(len(s))) + 2 + len(s) + 2 // $L\r\ns\r\n
	b := make([]byte, 0, size)
	b = append(b, BulkString)
	b = strconv.AppendInt(b, int64(len(s)), 10)
	b = append(b, CRLF...)
	b = append(b, s...)
	b = append(b, CRLF...)
	return b
}

// MakeNull creates a null protocol object.
func MakeNull() []byte {
	return nullBytes
}

// MakeArray creates an array protocol object.
func MakeArray(args ...[]byte) []byte {
	totalSize := 1 + numDigits(int64(len(args))) + 2 // *N\r\n
	for _, arg := range args {
		totalSize += len(arg)
	}

	b := make([]byte, 0, totalSize)
	b = append(b, Array)
	b = strconv.AppendInt(b, int64(len(args)), 10)
	b = append(b, CRLF...)
	for _, arg := range args {
		b = append(b, arg...)
	}

	return b
}

func numDigits(n int64) int {
	if n < 0 {
		n = -n
	}

	switch {
	case n < 10:
		return 1
	case n < 100:
		return 2
	case n < 1000:
		return 3
	case n < 10000:
		return 4
	case n < 100000:
		return 5
	case n < 1000000:
		return 6
	case n < 10000000:
		return 7
	case n < 100000000:
		return 8
	case n < 1000000000:
		return 9
	case n < 10000000000:
		return 10
	case n < 100000000000:
		return 11
	case n < 1000000000000:
		return 12
	case n < 10000000000000:
		return 13
	case n < 100000000000000:
		return 14
	case n < 1000000000000000:
		return 15
	case n < 10000000000000000:
		return 16
	case n < 100000000000000000:
		return 17
	case n < 1000000000000000000:
		return 18
	default:
		// Max int64 is 9,223,372,036,854,775,807, which has 19 digits.
		return 19
	}
}
