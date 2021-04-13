package server

import (
	"fmt"
	"os"
	"syscall"

	"github.com/HotPotatoC/kvstore/database"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"github.com/HotPotatoC/kvstore/pkg/utils"
	"github.com/HotPotatoC/kvstore/server/stats"
	"go.uber.org/zap"
)

// Server represents the database server
type Server struct {
	db     database.Store
	log    *zap.SugaredLogger
	server *tcp.Server
	// Info
	*stats.Stats
}

// New creates a new kvstore server
func New(version, build string) *Server {
	return &Server{
		db:  database.New(),
		log: logger.NewLogger(),
		Stats: &stats.Stats{
			Version: version,
			Build:   build,
		},
	}
}

// Start runs the server
func (s *Server) Start(host string, port int) {
	s.startupMessage()
	s.server = tcp.New()

	s.attachHooks()
	s.Stats.Init()

	s.TCPHost = host
	s.TCPPort = port

	s.server.Listen(host, port)
	s.printLogo()
	s.log.Info("Ready to accept connections.")

	rcvSignal := <-utils.WaitForSignals(os.Interrupt, syscall.SIGTERM)

	s.shutdown(rcvSignal)
}

func (s *Server) shutdown(signal os.Signal) {
	s.log.Infof("received %s signal", signal)
	s.log.Info("Shutting down server...")
	s.server.Stop()
	s.log.Info("Goodbye 👋")
}

func (s *Server) startupMessage() {
	s.log.Info("KVStore is starting...")
	s.log.Infof("version=%s build=%s pid=%d", s.Version, s.Build, os.Getpid())
	s.log.Info("starting tcp server...")
}

func (s *Server) printLogo() {
	logo := "\t _               _\n" +
		"\t| |             | |\n" +
		"\t| | ____   _____| |_ ___  _ __ ___\n" +
		"\t| |/ /\\ \\ / / __| __/ _ \\| '__/ _ \\ \n" +
		"\t|   <  \\ V /\\_  \\ || (_) | | |  __/\n" +
		"\t|_|\\_\\  \\_/ |___/\\__\\___/|_|  \\___|\n\n"

	details := "\tStarted KVStore %s server\n" +
		"\t    Port: %d\n" +
		"\t    PID: %d\n\n"

	fmt.Printf(logo+details, s.Version, s.TCPPort, os.Getpid())
}
