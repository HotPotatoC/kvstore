package protocol

import (
	"bufio"
	"bytes"
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
	br  *bufio.Reader
	buf []byte
}

// NewReader returns a new protocol reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		br:  bufio.NewReader(r),
		buf: make([]byte, 0, 64*1024),
	}
}

// ReadObject reads an object from the reader.
func (r *Reader) ReadObject() (any, error) {
	// read the line from the stream
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case SimpleString, Error:
		// Avoid allocation for frequent "+OK" and "+PONG"
		val := line[1:]
		if bytes.Equal(val, []byte("OK")) {
			return "OK", nil
		}
		if bytes.Equal(val, []byte("PONG")) {
			return "PONG", nil
		}

		return string(val), nil
	case Integer:
		return r.parseInt(line[1:])
	case BulkString:
		n, err := r.parseLen(line[1:])
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, nil // Nil bulk string
		}

		// Read data and trailing CRLF into the reusable buffer
		totalLen := n + 2
		if cap(r.buf) < totalLen {
			r.buf = make([]byte, totalLen)
		} else {
			r.buf = r.buf[:totalLen]
		}

		if _, err = io.ReadFull(r.br, r.buf); err != nil {
			return nil, err
		}

		if r.buf[n] != '\r' || r.buf[n+1] != '\n' {
			return nil, ErrInvalidSyntax
		}

		return r.buf[:n], nil
	case Array:
		n, err := r.parseLen(line[1:])
		if err != nil {
			return nil, err
		}

		if n == -1 {
			return nil, nil
		}

		result := make([]any, n)
		var dataBuf bytes.Buffer
		for i := 0; i < n; i++ {
			// Read the next object in the array
			obj, err := r.ReadObject()
			if err != nil {
				return nil, err
			}

			if val, ok := obj.([]byte); ok {
				// Get the current position in the buffer, which is the start of our new slice.
				start := dataBuf.Len()
				// Append the data from the temporary read buffer into our persistent array buffer.
				dataBuf.Write(val)
				// Store a slice that points to the data we just wrote inside dataBuf.
				result[i] = dataBuf.Bytes()[start:]
			} else {
				// For other types (int, string, nil), no copy is needed.
				result[i] = obj
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
		buf := make([]byte, len(p), len(p)*2)
		copy(buf, p)

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
