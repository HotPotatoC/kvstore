package server

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/packet"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

// Server represents the database server
type Server struct {
	version string
	build   string
	db      *hashtable.HashTable
	clients uint64
}

func init() {
	log = logger.NewLogger()
}

// New creates a new kvstore server
func New(version, build string) *Server {
	return &Server{
		version: version,
		build:   build,
		db:      hashtable.New(),
		clients: 0,
	}
}

// Start runs the server
func (s *Server) Start(host string, port int) {
	log.Info("KVStore is starting...")
	log.Infof("version=%s build=%s pid=%d", s.version, s.build, os.Getpid())
	log.Info("starting tcp server...")
	tcpServer := tcp.New()

	tcpServer.OnConnected = s.onConnected
	tcpServer.OnDisconnected = s.onDisconnected
	tcpServer.OnMessage = s.onMessage

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

`, s.version, port, os.Getpid())
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
	atomic.AddUint64(&s.clients, 1)
}

func (s *Server) onDisconnected(conn net.Conn) {
	// Decrement connected clients
	atomic.AddUint64(&s.clients, ^uint64(0))
}

func (s *Server) onMessage(conn net.Conn, msg []byte) {
	buffer := bytes.NewBuffer(msg)
	packet := new(packet.Packet)

	err := packet.Decode(buffer)
	if err != nil {
		log.Error(err)
	}

	command := command.New(s.db, packet.Cmd)
	if command == nil {
		conn.Write([]byte(fmt.Sprintf("Command '%s' does not exist\n", packet.Cmd.String())))
	} else {
		result := command.Execute(strings.Split(string(packet.Args), " "))
		conn.Write([]byte(fmt.Sprintf("%s\n", result)))
	}
}
