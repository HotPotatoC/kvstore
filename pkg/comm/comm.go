package comm

import (
	"net"
	"time"
)

type Comm struct {
	connection net.Conn
}

func New(addr string) (*Comm, error) {
	conn, err := newConnection(addr, time.Second*30)
	return &Comm{
		connection: conn,
	}, err
}

func (c *Comm) Connection() net.Conn {
	return c.connection
}

func (c *Comm) Send(b []byte) (err error) {
	_, err = c.connection.Write(b)
	return
}

func (c *Comm) Read() (buffer []byte, n int, err error) {
	buffer = make([]byte, 2048)
	n, err = c.connection.Read(buffer)
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
