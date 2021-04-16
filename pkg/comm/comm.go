package comm

import (
	"net"
)

// Comm is a basic stream communication
type Comm struct {
	conn net.Conn

	cfg *Config
}

// New creates a new comm
func New(cfg *Config) (*Comm, error) {
	cfg.init()
	conn, err := newConnection(cfg)
	return &Comm{
		conn: conn,
		cfg:  cfg,
	}, err
}

// NewWithConn creates a new comm with the given connection
func NewWithConn(conn net.Conn) *Comm {
	return &Comm{
		conn: conn,
	}
}

// Connection returns the connection
func (c *Comm) Connection() net.Conn {
	return c.conn
}

// Send writes data to the connection
func (c *Comm) Send(b []byte) (err error) {
	_, err = c.conn.Write(b)
	return
}

// Read reads data from the connection
func (c *Comm) Read() (buffer []byte, n int, err error) {
	buffer = make([]byte, 2048)
	n, err = c.conn.Read(buffer)
	return
}

func newConnection(cfg *Config) (net.Conn, error) {
	connection, err := net.DialTimeout("tcp", cfg.Addr, cfg.DialTimeout)
	if err != nil {
		return nil, err
	}

	if err := connection.SetReadDeadline(cfg.ReadDeadline); err != nil {
		return nil, err
	}

	if err := connection.SetWriteDeadline(cfg.WriteDeadline); err != nil {
		return nil, err
	}

	return connection, nil
}
