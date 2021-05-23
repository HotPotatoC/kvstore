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
)

type Server struct {
	*gnet.EventServer
	*stats.Stats

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

	logger.L().Debug("initializing aof persistor service...")
	go s.storageAOFPersistor.Run(s.tick)

	codec := framecodec.NewLengthFieldBasedFrameCodec(
		framecodec.NewGNETDefaultLengthFieldBasedFrameEncoderConfig(),
		framecodec.NewGNETDefaultLengthFieldBasedFrameDecoderConfig())

	logger.L().Debug("server configurations:")
	logger.L().Debug("- SO_KEEPALIVE: 10 minutes")
	logger.L().Debug("- Codec: Length-Field-Based-Frame Codec")
	err := gnet.Serve(s, fmt.Sprintf("tcp://%s:%d", host, port),
		gnet.WithTCPKeepAlive(10*time.Minute),
		gnet.WithCodec(codec))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) React(frame []byte, conn gnet.Conn) (out []byte, action gnet.Action) {
	data := append([]byte{}, frame...)
	logger.L().Debugf("received data: %s [%s]", data, conn.RemoteAddr().String())
	logger.L().Debugf("submitting a task to the worker pool for the given data [%s]", conn.RemoteAddr().String())
	err := s.pool.Submit(func() {
		raw := strings.Fields(string(data))
		if len(raw) < 1 {
			logger.L().Debugf("client provided a missing command [%s]", conn.RemoteAddr().String())
			err := conn.AsyncWrite([]byte("missing command\n"))
			if err != nil {
				logger.L().Error(err)
			}
			return
		}

		logger.L().Debugf("parsing command [%s]", conn.RemoteAddr().String())
		cmd := strings.ToLower(
			strings.TrimSpace(raw[0]))
		args := strings.TrimSpace(
			strings.TrimPrefix(string(data), raw[0]))

		op, err := s.parseCommand(cmd, args)
		if err != nil {
			err = conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", cmd)))
			if err != nil {
				logger.L().Error(err)
			}
		}

		command := command.New(s.storage, s.Stats, op)
		if command == nil {
			err := conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", cmd)))
			if err != nil {
				logger.L().Error(err)
			}
		} else {
			result := command.Execute(strings.Split(string(args), " "))
			err := conn.AsyncWrite([]byte(fmt.Sprintf("%s\n", result)))
			if err != nil {
				logger.L().Error(err)
			}
		}
	})
	if err != nil {
		logger.L().Error(err)
	}

	return
}

func (s *Server) OnShutdown(svr gnet.Server) {
	logger.L().Info("Shutting down server...")

	s.connectedClients.Range(func(key, value interface{}) bool {
		c := value.(gnet.Conn)
		c.Close()
		s.connectedClients.Delete(key)
		return true
	})

	s.storageAOFPersistor.Close()

	logger.L().Info("Goodbye 👋")
}

func (s *Server) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
	logger.L().Debugf("received a client [%s]", conn.RemoteAddr().String())
	s.ConnectedCount++
	s.TotalConnectionsCount++

	s.connectedClients.Store(conn.RemoteAddr().String(), conn)
	return
}

func (s *Server) OnClosed(conn gnet.Conn, err error) (action gnet.Action) {
	logger.L().Debugf("client closed the connection [%s]", conn.RemoteAddr().String())
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
		logger.L().Debugf("received a write command: %s", command.SET.String())
		writeToAOF(cmd, args)
		logger.L().Debug("enqueued the command into the aof persistor queue")
		return command.SET, nil
	case command.SETEX.String():
		logger.L().Debugf("received a write command: %s", command.SETEX.String())
		writeToAOF(cmd, args)
		logger.L().Debug("enqueued the command into the aof persistor queue")
		return command.SETEX, nil
	case command.DEL.String():
		logger.L().Debugf("received a write command: %s", command.DEL.String())
		writeToAOF(cmd, args)
		logger.L().Debug("enqueued the command into the aof persistor queue")
		return command.DEL, nil
	case command.FLUSHALL.String():
		logger.L().Debugf("received a write command: %s", command.FLUSHALL.String())
		err := s.storageAOFPersistor.Truncate()
		if err != nil {
			return -1, err
		}

		return command.FLUSHALL, nil
	case command.GET.String():
		logger.L().Debugf("received a read command: %s", command.GET.String())
		return command.GET, nil
	case command.KEYS.String():
		logger.L().Debugf("received a read command: %s", command.KEYS.String())
		return command.KEYS, nil
	case command.INFO.String():
		logger.L().Debugf("received a read command: %s", command.INFO.String())
		return command.INFO, nil
	}
	return -1, command.ErrCommandDoesNotExist
}

func (s *Server) startupMessage() {
	logger.L().Info("kvstore is starting...")
	logger.L().Infof("version=%s build=%s pid=%d", s.Version, s.Build, os.Getpid())
	logger.L().Info("starting gnet event server...")
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
	logger.L().Infof("Started kvstore server")
	logger.L().Info("Ready to accept connections.")
}
