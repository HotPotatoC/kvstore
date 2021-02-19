package tcp

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/logger"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

type server struct {
	listener net.Listener

	// Callbacks
	OnConnected    func(conn net.Conn)
	OnDisconnected func(conn net.Conn)
	OnMessage      func(conn net.Conn, msg []byte)
}

func init() {
	log = logger.NewLogger()
}

func New() *server {
	return &server{
		OnConnected:    func(_ net.Conn) { noop() },
		OnDisconnected: func(_ net.Conn) { noop() },
		OnMessage:      func(_ net.Conn, _ []byte) { noop() },
	}
}

func (s *server) Listen(host string, port int) {
	server := makeTCPListener(host, port)
	s.listener = server

	go s.acceptConnections()
}

func (s *server) Stop() {
	s.listener.Close()
}

func (s *server) acceptConnections() {
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

func (s *server) handleConnection(conn net.Conn) {
	defer s.OnDisconnected(conn)
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Hour)); err != nil {
		log.Fatalf("failed setting connection read deadline: %v", err)
	}

	for {
		data, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else if err != io.EOF {
				log.Fatalf("read error: %v", err)
				return
			}
		}

		if data == nil {
			return
		}

		s.OnMessage(conn, data)
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

// No operation
func noop() {}
