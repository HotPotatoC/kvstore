package server

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/database"
	"github.com/HotPotatoC/kvstore/internal/server/stats"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"github.com/fatih/color"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
	"go.uber.org/zap"
)

type Server struct {
	*gnet.EventServer
	*stats.Stats

	db  database.Store
	log *zap.SugaredLogger

	workerPool *goroutine.Pool
}

func New(version string, build string) *Server {
	return &Server{
		db:         database.New(),
		log:        logger.New(),
		workerPool: goroutine.Default(),
		Stats: &stats.Stats{
			Version: version,
			Build:   build,
		},
	}
}

func (s *Server) Start(host string, port int) error {
	s.TCPHost = host
	s.TCPPort = port

	s.startupMessage()

	codec := gnet.NewLengthFieldBasedFrameCodec(
		gnet.EncoderConfig{
			ByteOrder:                       binary.BigEndian,
			LengthFieldLength:               4,
			LengthAdjustment:                0,
			LengthIncludesLengthFieldLength: false,
		},
		gnet.DecoderConfig{
			ByteOrder:           binary.BigEndian,
			LengthFieldOffset:   0,
			LengthFieldLength:   4,
			LengthAdjustment:    0,
			InitialBytesToStrip: 4,
		})

	err := gnet.Serve(s, fmt.Sprintf("tcp://%s:%d", host, port),
		gnet.WithTCPKeepAlive(10*time.Minute),
		gnet.WithCodec(codec))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	s.printLogo(srv)
	return
}

func (s *Server) OnShutdown(svr gnet.Server) {
	s.log.Info("Shutting down server...")
	s.log.Info("Goodbye 👋")
}

func (s *Server) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
	s.ConnectedCount++
	s.TotalConnectionsCount++
	return
}

func (s *Server) OnClosed(conn gnet.Conn, err error) (action gnet.Action) {
	s.ConnectedCount--
	return
}

func (s *Server) React(frame []byte, conn gnet.Conn) (out []byte, action gnet.Action) {
	data := append([]byte{}, frame...)
	err := s.workerPool.Submit(func() {
		raw := strings.Fields(string(data))
		if len(raw) < 1 {
			err := conn.AsyncWrite([]byte("missing command\n"))
			if err != nil {
				s.log.Error(err)
			}
			return
		}

		cmd := strings.ToLower(
			strings.TrimSpace(raw[0]))
		args := strings.TrimSpace(
			strings.TrimPrefix(string(data), raw[0]))

		var op command.Op
		switch string(cmd) {
		case command.SET.String():
			op = command.SET
		case command.SETEX.String():
			op = command.SETEX
		case command.GET.String():
			op = command.GET
		case command.DEL.String():
			op = command.DEL
		case command.LIST.String():
			op = command.LIST
		case command.KEYS.String():
			op = command.KEYS
		case command.FLUSH.String():
			op = command.FLUSH
		case command.INFO.String():
			op = command.INFO
		default:
			err := conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", cmd)))
			if err != nil {
				s.log.Error(err)
			}
		}

		command := command.New(s.db, s.Stats, op)
		if command == nil {
			err := conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", cmd)))
			if err != nil {
				s.log.Error(err)
			}
		} else {
			result := command.Execute(strings.Split(string(args), " "))
			err := conn.AsyncWrite([]byte(fmt.Sprintf("%s\n", result)))
			if err != nil {
				s.log.Error(err)
			}
		}
	})

	if err != nil {
		s.log.Error(err)
	}
	return
}

func (s *Server) startupMessage() {
	s.log.Info("kvstore is starting...")
	s.log.Infof("version=%s build=%s pid=%d", s.Version, s.Build, os.Getpid())
	s.log.Info("starting gnet event server...")
}

func (s *Server) printLogo(srv gnet.Server) {
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
	s.log.Infof("Started kvstore server")
	s.log.Info("Ready to accept connections.")
}
