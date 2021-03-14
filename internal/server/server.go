package server

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/packet"
	"github.com/HotPotatoC/kvstore/internal/stats"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

// Server represents the database server
type Server struct {
	db *hashtable.HashTable

	// Info
	*stats.Stats
	Version string `json:"version"`
	Build   string `json:"build"`
}

func init() {
	log = logger.NewLogger()
}

// New creates a new kvstore server
func New(version, build string) *Server {
	return &Server{
		db:      hashtable.New(),
		Version: version,
		Build:   build,
		Stats: &stats.Stats{},
	}
}

// Start runs the server
func (s *Server) Start(host string, port int) {
	log.Info("KVStore is starting...")
	log.Infof("version=%s build=%s pid=%d", s.Version, s.Build, os.Getpid())
	log.Info("starting tcp server...")
	tcpServer := tcp.New()

	tcpServer.OnConnected = s.onConnected
	tcpServer.OnDisconnected = s.onDisconnected
	tcpServer.OnMessage = s.onMessage

	s.Stats.Init()

	s.TCPHost = host
	s.TCPPort = port

	tcpServer.Listen(host, port)
	fmt.Printf(`
	 _               _
	| |             | |
	| | ____   _____| |_ ___  _ __ ___
	| |/ /\ \ / / __| __/ _ \| '__/ _ \
	|   <  \ V /\__ \ || (_) | | |  __/
	|_|\_\  \_/ |___/\__\___/|_|  \___|

	Started KVStore %s server
	  Port: %d
	  PID: %d

`, s.Version, port, os.Getpid())
	log.Info("Ready to accept connections.")

	// Graceful shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	signal := <-c

	log.Infof("received %s signal", signal)
	log.Info("Shutting down server...")
	tcpServer.Stop()
	log.Info("Goodbye 👋")
}

func (s *Server) onConnected(conn net.Conn) {
	// Increment connected clients
	s.ConnectedCount++
	s.TotalConnectionsCount++
}

func (s *Server) onDisconnected(conn net.Conn) {
	// Decrement connected clients
	s.ConnectedCount--
}

func (s *Server) onMessage(conn net.Conn, msg []byte) {
	buffer := bytes.NewBuffer(msg)
	packet := new(packet.Packet)

	err := packet.Decode(buffer)
	if err != nil {
		log.Error(err)
	}

	command := command.New(s.db, s.Stats, packet.Cmd)
	if command == nil {
		conn.Write([]byte(fmt.Sprintf("Command '%s' does not exist\n", packet.Cmd.String())))
	} else {
		result := command.Execute(strings.Split(string(packet.Args), " "))
		conn.Write([]byte(fmt.Sprintf("%s\n", result)))
	}
}
