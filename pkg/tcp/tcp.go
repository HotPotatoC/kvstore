package tcp

import (
	"io"
	"net"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/logger"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

// Server represents a tcp server
type Server struct {
	listener net.Listener

	// Callbacks
	OnConnected    func(conn net.Conn)
	OnDisconnected func(conn net.Conn)
	OnMessage      func(conn net.Conn, msg []byte)
}

func init() {
	log = logger.New()
}

// New creates a new tcp server
func New() *Server {
	return &Server{
		OnConnected:    func(_ net.Conn) { /** no-op by default */ },
		OnDisconnected: func(_ net.Conn) { /** no-op by default */ },
		OnMessage:      func(_ net.Conn, _ []byte) { /** no-op by default */ },
	}
}

// Listen listens to the given address
func (s *Server) Listen(host string, port int) {
	server := makeTCPListener(host, port)
	s.listener = server

	go s.acceptConnections()
}

// Stop closes the tcp server
func (s *Server) Stop() {
	s.listener.Close()
}

func (s *Server) acceptConnections() {
	// Handles multiple client connections by spawning a goroutine for each client
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		s.OnConnected(conn)

		go func(conn net.Conn) {
			defer conn.Close()
			s.handleConnection(conn)
		}(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.OnDisconnected(conn)
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Hour)); err != nil {
		log.Fatalf("failed setting connection read deadline: %v", err)
	}

	buffer := make([]byte, 2048)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else if err != io.EOF {
				log.Fatalf("read error: %v", err)
				return
			}
		}

		if n == 0 {
			return
		}

		s.OnMessage(conn, buffer)
	}
}

func makeTCPListener(ip string, port int) net.Listener {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})
	if err != nil {
		log.Fatalf("failed creating tcp listener: %v", err)
	}

	return listener
}
