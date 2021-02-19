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
	"time"

	"github.com/HotPotatoC/kvstore/internal/cli"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

type server struct {
	conn net.Conn
}

var (
	ErrConnectionRefused = errors.New("Connection refused")
)

func init() {
	log = logger.NewLogger()
}

func New() *server {
	return &server{}
}

func (s *server) Start(host string, port int) {
	log.Info("KVStore is starting...")
	log.Infof("starting tcp server... pid=%d", os.Getpid())
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

	Started KVStore server
	  Port: %d
	  PID: %d

`, port, os.Getpid())
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
	cmd := bytes.ToLower(
		bytes.TrimSpace(bytes.Split(msg, []byte(" "))[0]))
	args := bytes.TrimSpace(
		bytes.TrimPrefix(msg, cmd))

	command := cli.GetCommand(string(cmd))
	if command == nil {
		conn.Write([]byte(fmt.Sprintf("Command '%s' does not exist\n", cmd)))
	} else {
		t1 := time.Now()
		command.Execute(string(args))
		t2 := time.Now()
		diff := t2.Sub(t1)
		conn.Write([]byte(fmt.Sprintf("%s %fs\n", strings.ToUpper(string(cmd)), diff.Seconds())))
	}
}
