package protocol

import (
	"bufio"
	"errors"
	"io"
)

var (
	// ErrInvalidSyntax is returned when the reader encounters an invalid
	// syntax.
	ErrInvalidSyntax = errors.New("invalid syntax")
	// ErrMalformedLength is returned when the reader encounters a malformed
	// length.
	ErrMalformedLength = errors.New("malformed length")
)

// Reader is a protocol reader.
type Reader struct {
	br *bufio.Reader
}

// NewReader returns a new protocol reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{br: bufio.NewReader(r)}
}

// ReadObject reads an object from the reader.
func (r *Reader) ReadObject() (interface{}, error) {
	// read the line from the stream
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case SimpleString, Error:
		// Avoid allocation for frequent "+OK" and "+PONG"
		if string(line[1:]) == "OK" {
			return "OK", nil
		}
		if string(line[1:]) == "PONG" {
			return "PONG", nil
		}

		return string(line[1:]), nil
	case Integer:
		return r.parseInt(line[1:])
	case BulkString:
		n, err := r.parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(r.br, p)
		if err != nil {
			return nil, err
		}
		if line, err := r.readLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, ErrInvalidSyntax
		}

		return p, nil
	case Array:
		len, err := r.parseLen(line[1:])
		if err != nil {
			return nil, err
		}

		if len == -1 {
			return nil, nil
		}

		result := make([]interface{}, len)
		for i := 0; i < len; i++ {
			result[i], err = r.ReadObject()
			if err != nil {
				return nil, err
			}
		}

		return result, nil
	}

	return nil, ErrInvalidSyntax
}

// readLine reads a line from the reader.
func (r *Reader) readLine() ([]byte, error) {
	// read the line from the stream using ReadSlice to avoid allocations
	// Reference: https://github.com/gomodule/redigo/blob/master/redis/conn.go#L543
	p, err := r.br.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) {
		buf := append([]byte{}, p...)

		for err == bufio.ErrBufferFull {
			p, err = r.br.ReadSlice('\n')
			buf = append(buf, p...)
		}

		p = buf
	}
	if err != nil {
		return nil, err
	}

	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, ErrInvalidSyntax
	}

	return p[:i], nil
}

// parseLen parses a length from the given protocol data.
func (r *Reader) parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, ErrMalformedLength
	}

	if len(p) == 2 && p[0] == '-' && p[1] == '1' {
		return -1, nil
	}

	len := 0
	for _, b := range p {
		if b == '\r' {
			break
		}
		if b < '0' || b > '9' {
			return -1, ErrMalformedLength
		}
		len *= 10
		len += int(b - '0')
	}

	return len, nil
}

// parseInt parses an integer from given protocol data.
func (r *Reader) parseInt(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, ErrMalformedLength
	}

	var negation bool
	if p[0] == '-' {
		negation = true
		p = p[1:]
	}

	var result int
	for _, b := range p {
		if b < '0' || b > '9' {
			return -1, ErrMalformedLength
		}

		result = result*10 + int(b-'0')
	}

	if negation {
		result = -result
	}

	return result, nil
}
