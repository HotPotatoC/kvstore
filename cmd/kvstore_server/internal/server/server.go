package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/packet"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

type server struct {
	version string
	build   string
	db      *hashtable.HashTable
}

var (
	ErrConnectionRefused = errors.New("Connection refused")
)

func init() {
	log = logger.NewLogger()
}

func New(version, build string) *server {
	return &server{
		version: version,
		build:   build,
		db:      hashtable.NewHashTable(),
	}
}

func (s *server) Start(host string, port int) {
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

func (s *server) onConnected(conn net.Conn) {
	log.Infof("Connected: %s", conn.RemoteAddr().String())
}

func (s *server) onDisconnected(conn net.Conn) {
	log.Infof("%s disconnected", conn.RemoteAddr().String())
}

func (s *server) onMessage(conn net.Conn, msg []byte) {
	buffer := bytes.NewBuffer(msg)
	packet := new(packet.Packet)

	err := packet.Decode(buffer)
	if err != nil {
		fmt.Println(err)
	}

	command := command.GetCommand(s.db, packet.Cmd)
	if command == nil {
		conn.Write([]byte(fmt.Sprintf("Command '%s' does not exist\n", packet.Cmd.String())))
	} else {
		result := command.Execute(strings.Split(string(packet.Args), " "))
		conn.Write([]byte(fmt.Sprintf("%s\n", result)))
	}
}
