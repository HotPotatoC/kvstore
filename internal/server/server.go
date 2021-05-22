package server

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/framecodec"
	"github.com/HotPotatoC/kvstore/internal/logger"
	"github.com/HotPotatoC/kvstore/internal/server/stats"
	"github.com/HotPotatoC/kvstore/internal/storage"
	"github.com/fatih/color"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
	"go.uber.org/zap"
)

type Server struct {
	*gnet.EventServer
	*stats.Stats

	log  *zap.SugaredLogger
	pool *goroutine.Pool

	connectedClients sync.Map

	storage             storage.Store
	storageAOFPersistor *storage.AOFPersistor

	tick         time.Duration
	commandQueue chan string
}

func New(version string, build string) *Server {
	storageAOFPersistor, err := storage.NewAOFPersistor()
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		Stats: &stats.Stats{
			Version: version,
			Build:   build,
		},
		log:                 logger.New(),
		pool:                goroutine.Default(),
		storage:             storage.New(),
		storageAOFPersistor: storageAOFPersistor,
		tick:                60 * time.Second,
		commandQueue:        make(chan string),
	}
}

func (s *Server) Start(host string, port int) error {
	s.TCPHost = host
	s.TCPPort = port

	s.startupMessage()

	go s.storageAOFPersistor.Run(s.tick)

	codec := framecodec.NewLengthFieldBasedFrameCodec(
		framecodec.NewGNETDefaultLengthFieldBasedFrameEncoderConfig(),
		framecodec.NewGNETDefaultLengthFieldBasedFrameDecoderConfig())

	err := gnet.Serve(s, fmt.Sprintf("tcp://%s:%d", host, port),
		gnet.WithTicker(true),
		gnet.WithTCPKeepAlive(10*time.Minute),
		gnet.WithCodec(codec))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) React(frame []byte, conn gnet.Conn) (out []byte, action gnet.Action) {
	data := append([]byte{}, frame...)
	err := s.pool.Submit(func() {
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

		op, err := s.parseCommand(cmd, args)
		if err != nil {
			err = conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", cmd)))
			if err != nil {
				s.log.Error(err)
			}
		}

		command := command.New(s.storage, s.Stats, op)
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

func (s *Server) OnShutdown(svr gnet.Server) {
	s.log.Info("Shutting down server...")

	s.connectedClients.Range(func(key, value interface{}) bool {
		c := value.(gnet.Conn)
		c.Close()
		s.connectedClients.Delete(key)
		return true
	})

	s.storageAOFPersistor.Close()

	s.log.Info("Goodbye 👋")
}

func (s *Server) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
	s.ConnectedCount++
	s.TotalConnectionsCount++

	s.connectedClients.Store(conn.RemoteAddr().String(), conn)
	return
}

func (s *Server) OnClosed(conn gnet.Conn, err error) (action gnet.Action) {
	s.ConnectedCount--
	s.connectedClients.Delete(conn.RemoteAddr().String())
	return
}

func (s *Server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	s.printLogo(srv)
	return
}

func (s *Server) parseCommand(cmd, args string) (command.Op, error) {
	writeToAOF := func(cmd, args string) {
		s.storageAOFPersistor.Add(fmt.Sprintf("%s %s", cmd, args))
	}

	switch string(cmd) {
	case command.SET.String():
		writeToAOF(cmd, args)
		return command.SET, nil
	case command.SETEX.String():
		writeToAOF(cmd, args)
		return command.SETEX, nil
	case command.DEL.String():
		writeToAOF(cmd, args)
		return command.DEL, nil
	case command.FLUSHALL.String():
		err := s.storageAOFPersistor.Truncate()
		if err != nil {
			return -1, err
		}

		return command.FLUSHALL, nil
	case command.GET.String():
		return command.GET, nil
	case command.KEYS.String():
		return command.KEYS, nil
	case command.INFO.String():
		return command.INFO, nil
	}
	return -1, command.ErrCommandDoesNotExist
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
