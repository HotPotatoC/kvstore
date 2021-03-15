package server

import (
	"fmt"
	"os"
	"syscall"

	"github.com/HotPotatoC/kvstore/internal/server/stats"
	"github.com/HotPotatoC/kvstore/pkg/hashtable"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"github.com/HotPotatoC/kvstore/pkg/utils"
	"go.uber.org/zap"
)

// Server represents the database server
type Server struct {
	db  *hashtable.HashTable
	log *zap.SugaredLogger
	// Info
	*stats.Stats
}

// New creates a new kvstore server
func New(version, build string) *Server {
	return &Server{
		db:  hashtable.New(),
		log: logger.NewLogger(),
		Stats: &stats.Stats{
			Version: version,
			Build:   build,
		},
	}
}

// Start runs the server
func (s *Server) Start(host string, port int) {
	s.log.Info("KVStore is starting...")
	s.log.Infof("version=%s build=%s pid=%d", s.Version, s.Build, os.Getpid())
	s.log.Info("starting tcp server...")
	tcpServer := tcp.New()

	s.attachHooks(tcpServer)
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
	s.log.Info("Ready to accept connections.")

	signal := <-utils.WaitForSignals(os.Interrupt, syscall.SIGTERM)

	s.log.Infof("received %s signal", signal)
	s.log.Info("Shutting down server...")
	tcpServer.Stop()
	s.log.Info("Goodbye 👋")
}
