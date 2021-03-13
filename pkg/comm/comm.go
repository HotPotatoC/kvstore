package comm

import (
	"net"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/tcp"
)

// Comm is the basic tcp communication
type Comm struct {
	Conn net.Conn
}

// New creates a new tcp comm
func New(addr string) (*Comm, error) {
	conn, err := newConnection(addr, time.Second*30)
	return &Comm{
		Conn: conn,
	}, err
}

func NewWithConn(conn net.Conn) *Comm {
	return &Comm{
		Conn: conn,
	}
}

// Connection returns the connection
func (c *Comm) Connection() net.Conn {
	return c.Conn
}

// Send writes data to the connection
func (c *Comm) Send(b []byte) (err error) {
	_, err = c.Conn.Write(b)
	return
}

// Read reads data from the connection
func (c *Comm) Read() (buffer []byte, n int, err error) {
	buffer = make([]byte, tcp.MaxTCPBufferSize)
	n, err = c.Conn.Read(buffer)
	return
}

func newConnection(addr string, timeout time.Duration) (net.Conn, error) {
	connection, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	if err := connection.SetDeadline(time.Now().Add(time.Hour * 1)); err != nil {
		return nil, err
	}

	if err := connection.SetReadDeadline(time.Now().Add(time.Hour * 1)); err != nil {
		return nil, err
	}

	if err := connection.SetWriteDeadline(time.Now().Add(time.Hour * 1)); err != nil {
		return nil, err
	}

	return connection, nil
}
