package protocol

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

/*=======================================================================/
	Currently not used, atm using type.go since it's a lot faster but
	allocates more. Saving for later. :)
/=======================================================================*/

var (
	// ErrInvalidType is returned when the type is not valid.
	ErrInvalidType = errors.New("invalid type")
	// ErrInvalidSimpleString is returned when a simple string is not valid.
	ErrInvalidSimpleString = errors.New("invalid simple string")
	// ErrInvalidBulkString is returned when a bulk string is not valid.
	ErrInvalidBulkString = errors.New("invalid bulk string")
	// ErrInvalidInteger is returned when an integer is not valid.
	ErrInvalidInteger = errors.New("invalid integer")
)

// Writer is a protocol writer.
type Writer struct {
	bw *bufio.Writer
}

// NewWriter returns a new writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{bw: bufio.NewWriter(w)}
}

// WriteCommand writes a command.
func (w *Writer) WriteCommand(args ...string) error {
	w.bw.WriteByte(Array)
	w.bw.WriteString(strconv.Itoa(len(args)))
	w.bw.Write(CRLF)

	for _, arg := range args {
		w.bw.WriteByte(BulkString)
		w.bw.WriteString(strconv.Itoa(len(arg)))
		w.bw.Write(CRLF)
		w.bw.WriteString(arg)
		w.bw.Write(CRLF)
	}

	return w.bw.Flush()
}

// WriteSimpleString writes a simple string.
func (w *Writer) WriteSimpleString(s string) error {
	w.bw.WriteByte(SimpleString)
	w.bw.WriteString(s)
	w.bw.Write(CRLF)
	return w.bw.Flush()
}

// WriteError writes an error.
func (w *Writer) WriteError(s string) error {
	w.bw.WriteByte(Error)
	w.bw.WriteString(s)
	w.bw.Write(CRLF)
	return w.bw.Flush()
}

// WriteInteger writes an integer.
func (w *Writer) WriteInteger(i int64) error {
	w.bw.WriteByte(Integer)
	w.bw.WriteString(strconv.FormatInt(i, 10))
	w.bw.Write(CRLF)
	return w.bw.Flush()
}

// WriteBulkString writes a bulk string.
func (w *Writer) WriteBulkString(s string) error {
	w.bw.WriteByte(BulkString)
	w.bw.WriteString(strconv.Itoa(len(s)))
	w.bw.Write(CRLF)
	w.bw.WriteString(s)
	w.bw.Write(CRLF)
	return w.bw.Flush()
}
