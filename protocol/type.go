package protocol

import "strconv"

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
	CRLF = []byte{'\r', '\n'}
)

// MakeCommand creates a command protocol object.
func MakeCommand(args ...string) []byte {
	var b []byte
	b = append(b, Array)
	b = strconv.AppendInt(b, int64(len(args)), 10)
	b = append(b, CRLF...)
	for _, arg := range args {
		b = append(b, MakeBulkString(arg)...)
	}

	return b
}

// MakeSimpleString creates a simple string protocol object.
func MakeSimpleString(s string) []byte {
	var b []byte
	b = append(b, SimpleString)
	b = append(b, s...)
	b = append(b, CRLF...)
	return b
}

// MakeError creates an error protocol object.
func MakeError(s string) []byte {
	var b []byte
	b = append(b, Error)
	b = append(b, s...)
	b = append(b, CRLF...)
	return b
}

// MakeInteger creates an integer protocol object.
func MakeInteger(i int64) []byte {
	var b []byte
	b = append(b, Integer)
	b = strconv.AppendInt(b, i, 10)
	b = append(b, CRLF...)
	return b
}

// MakeBulkString creates a bulk string protocol object.
func MakeBulkString(s string) []byte {
	var b []byte
	b = append(b, BulkString)
	b = strconv.AppendInt(b, int64(len(s)), 10)
	b = append(b, CRLF...)
	b = append(b, s...)
	b = append(b, CRLF...)
	return b
}

// MakeNull creates a null protocol object.
func MakeNull() []byte {
	var b []byte
	b = append(b, BulkString)
	b = strconv.AppendInt(b, int64(-1), 10)
	b = append(b, CRLF...)
	return b
}

// MakeArray creates an array protocol object.
func MakeArray(args ...[]byte) []byte {
	var b []byte
	b = append(b, Array)
	b = strconv.AppendInt(b, int64(len(args)), 10)
	b = append(b, CRLF...)
	for _, arg := range args {
		b = append(b, arg...)
	}

	return b
}
