package protocol

import (
	"errors"
	"io"
)

const (
	MaxCommandArgs   = 1024
	MaxBulkLength    = 8 << 20
	MaxCommandLength = 16 << 20
)

var ErrCommandTooLarge = errors.New("command exceeds protocol limits")

// ParseCommand parses one flat RESP array of bulk strings. The returned slices
// alias data, so the caller must own data until command execution completes.
// scratch reuses argument storage. On error, no bytes are consumed.
func ParseCommand(data []byte, scratch [][]byte) ([][]byte, int, error) {
	// redis-cli --pipe separates its ECHO sentinel with an extra CRLF.
	start := 0
	for start < len(data) && data[start] == '\r' {
		if start+1 == len(data) {
			return nil, 0, io.ErrUnexpectedEOF
		}
		if data[start+1] != '\n' {
			return nil, 0, ErrInvalidSyntax
		}
		start += 2
		if start >= MaxCommandLength {
			return nil, 0, ErrCommandTooLarge
		}
	}
	n, used, err := commandLength(data[start:], Array)
	if err != nil {
		return nil, 0, err
	}
	offset := start + used
	if offset > MaxCommandLength || n > MaxCommandArgs {
		return nil, 0, ErrCommandTooLarge
	}
	args := scratch[:0]
	for i := 0; i < n; i++ {
		length, used, err := commandLength(data[offset:], BulkString)
		if err != nil {
			return nil, 0, err
		}
		if length > MaxBulkLength {
			return nil, 0, ErrCommandTooLarge
		}
		offset += used
		if offset > MaxCommandLength || length > MaxCommandLength-offset-2 {
			return nil, 0, ErrCommandTooLarge
		}
		if length > len(data)-offset {
			return nil, 0, io.ErrUnexpectedEOF
		}
		end := offset + length
		if end == len(data) {
			return nil, 0, io.ErrUnexpectedEOF
		}
		if data[end] != '\r' {
			return nil, 0, ErrInvalidSyntax
		}
		if end+1 == len(data) {
			return nil, 0, io.ErrUnexpectedEOF
		}
		if data[end+1] != '\n' {
			return nil, 0, ErrInvalidSyntax
		}
		args = append(args, data[offset:end:end])
		offset = end + 2
	}
	return args, offset, nil
}

func commandLength(data []byte, prefix byte) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	if data[0] != prefix {
		return 0, 0, ErrInvalidSyntax
	}
	n := 0
	const maxInt = int(^uint(0) >> 1)
	for i := 1; i < len(data); i++ {
		if i >= MaxCommandLength {
			return 0, 0, ErrCommandTooLarge
		}
		ch := data[i]
		if ch == '\r' {
			if i == 1 {
				return 0, 0, ErrMalformedLength
			}
			if i+1 == len(data) {
				return 0, 0, io.ErrUnexpectedEOF
			}
			if data[i+1] != '\n' {
				return 0, 0, ErrInvalidSyntax
			}
			return n, i + 2, nil
		}
		if ch < '0' || ch > '9' || n > (maxInt-int(ch-'0'))/10 {
			return 0, 0, ErrMalformedLength
		}
		n = n*10 + int(ch-'0')
	}
	return 0, 0, io.ErrUnexpectedEOF
}
