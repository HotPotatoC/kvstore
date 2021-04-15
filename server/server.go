package server

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/HotPotatoC/kvstore/command"
	"github.com/HotPotatoC/kvstore/database"
	"github.com/HotPotatoC/kvstore/packet"
	"github.com/HotPotatoC/kvstore/pkg/comm"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/HotPotatoC/kvstore/pkg/tcp"
	"github.com/HotPotatoC/kvstore/pkg/utils"
	"github.com/HotPotatoC/kvstore/server/stats"
	"github.com/fatih/color"
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

	s.server.OnConnected = s.onConnected
	s.server.OnDisconnected = s.onDisconnected
	s.server.OnMessage = s.onMessage

	s.Stats.Init()

	s.TCPHost = host
	s.TCPPort = port

	s.server.Listen(host, port)
	s.printLogo()
	s.log.Info("Ready to accept connections.")

	rcvSignal := <-utils.WaitForSignals(os.Interrupt, syscall.SIGTERM)

	s.shutdown(rcvSignal)
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

	comm := comm.NewWithConn(conn)

	err := packet.Decode(buffer)
	if err != nil {
		s.log.Error(err)
	}

	command := command.New(s.db, s.Stats, packet.Cmd)
	if command == nil {
		err := comm.Send([]byte(fmt.Sprintf("Command '%s' does not exist\n", packet.Cmd.String())))
		if err != nil {
			s.log.Error(err)
		}
	} else {
		result := command.Execute(strings.Split(string(packet.Args), " "))
		err := comm.Send([]byte(fmt.Sprintf("%s\n", result)))
		if err != nil {
			s.log.Error(err)
		}
	}
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
	yellow := color.New(color.FgYellow).SprintFunc()
	white := color.New(color.FgWhite, color.Bold).SprintFunc()
	var logo string
	logo += "\n"
	logo += yellow("       .\n")
	logo += yellow("   .-\"   \"-\n")
	logo += yellow(" .n         \"w\n")
	logo += yellow(" |  ^~   ⌐\"  |\n")
	logo += yellow(" |     ╠     |        .\n")
	logo += yellow(" |     ╡    ⌐|    .-\"   \"-\n")
	logo += yellow(" .╜\"-. ╡ .─\"  . #¬        .┴|\n")
	logo += yellow(" |  ^~ \".⌐'.-\"  ╫   ^¬.-\"   |\n")
	logo += yellow(" |     | #¬     |     |     |\n")

	logo += yellow(" |     | |  ^¬ .╝.    |    ⌐\"\n")
	logo += fmt.Sprintln(yellow(" .╜\"-. | |    |    \"-.|,^"), white("        Started kvstore %s server"))
	logo += fmt.Sprintln(yellow(" |  ^¬ \" ╜    |     ,"), white("              Port: %d"))
	logo += fmt.Sprintln(yellow(" |     | m\"\"-.| ,─\".X ."), white("            PID: %d"))
	logo += yellow(" |     | |  ^¬  ⌐'.⌐\"   \"─\n")

	logo += yellow("  \" ─. | |    | ╡╜        .╜|\n")
	logo += yellow("       \" |    | |   ^¬.-\"   |\n")
	logo += yellow("          \" ─.| |     |     |\n")
	logo += yellow("                ╙.    |    ⌐*\n")
	logo += yellow("                   \"─.|,^\n\n")

	fmt.Printf(logo, s.Version, s.TCPPort, os.Getpid())
}
