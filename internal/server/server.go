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
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
	*gnet.EventServer
	stats *stats.Stats

	pool *goroutine.Pool

	connectedClients sync.Map

	storage storage.Store
	aof     storage.Persistor
}

func New(version string, build string) *Server {
	return &Server{
		stats: &stats.Stats{
			Version: version,
			Build:   build,
		},
		pool:    goroutine.Default(),
		storage: storage.New(),
	}
}

func (s *Server) Start() error {
	s.stats.TCPHost = viper.GetString("server.host")
	s.stats.TCPPort = viper.GetInt("server.port")

	s.stats.Init()

	codec := framecodec.NewLengthFieldBasedFrameCodec(
		framecodec.NewGNETDefaultLengthFieldBasedFrameEncoderConfig(),
		framecodec.NewGNETDefaultLengthFieldBasedFrameDecoderConfig())

	addr := fmt.Sprintf("%s://%s:%d",
		viper.GetString("server.protocol"),
		viper.GetString("server.host"),
		viper.GetInt("server.port"))

	opts := []gnet.Option{
		gnet.WithMulticore(viper.GetBool("server.multicore")),
		gnet.WithReusePort(viper.GetBool("server.reuse_port")),
		gnet.WithReadBufferCap(viper.GetInt("server.read_buffer_cap")),
		gnet.WithLogger(zap.L().Sugar()),
		gnet.WithCodec(codec),
	}

	if viper.GetInt("server.tcp_keep_alive") > 0 {
		opts = append(opts, gnet.WithTCPKeepAlive(
			viper.GetDuration("server.tcp_keep_alive")*time.Second))
	}

	if viper.GetBool("aof.enabled") {
		aof, err := storage.NewAOFPersistor(viper.GetString("aof.path"))
		if err != nil {
			log.Fatal(err)
		}

		s.aof = aof

		logger.S().Info("initializing AOF persistor service...")
		logger.S().Info("reading AOF log...")

		for cmdStr := range s.aof.Read() {
			opts, err := command.Parse(cmdStr)
			if err != nil {
				logger.S().Errorf("err parsing command: %v", err)
				continue
			}

			command.New(s.storage, s.stats, opts.Op).Execute(opts.Args)
		}

		go s.aof.Run(viper.GetDuration("aof.persist_after") * time.Second)
	} else {
		logger.S().Info("AOF Persistor service disabled")
		// Use a mock version of AOF Persistor
		aof, _ := storage.NewMockAOFPersistor()

		s.aof = aof
	}

	s.startupMessage()
	err := gnet.Serve(s, addr, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) React(frame []byte, conn gnet.Conn) (out []byte, action gnet.Action) {
	data := append([]byte{}, frame...)
	conn.ResetBuffer()

	logger.S().Debugf("received data: %s [%s]", data, conn.RemoteAddr().String())
	logger.S().Debugf("submitting a task to the worker pool for the given data [%s]", conn.RemoteAddr().String())
	err := s.pool.Submit(func() {
		defer logger.S().Debugf("task done [%s]", conn.RemoteAddr().String())

		raw := strings.Fields(string(data))
		if len(raw) < 1 {
			logger.S().Debugf("client provided a missing command [%s]", conn.RemoteAddr().String())
			err := conn.AsyncWrite([]byte("missing command\n"))
			if err != nil {
				logger.S().Error(err)
			}
			return
		}

		logger.S().Debugf("parsing command [%s]", conn.RemoteAddr().String())
		opts, err := command.Parse(string(data))
		if err != nil {
			logger.S().Debugf("the command does not exists [%s]", conn.RemoteAddr().String())
			logger.S().Debugf("sending notice to the client [%s]", conn.RemoteAddr().String())
			err = conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", opts.Command)))
			if err != nil {
				logger.S().Error(err)
			}
		}

		cmd := command.New(s.storage, s.stats, opts.Op)
		if cmd == nil {
			logger.S().Debugf("the command does not exists [%s]", conn.RemoteAddr().String())
			logger.S().Debugf("sending notice to the client [%s]", conn.RemoteAddr().String())
			err := conn.AsyncWrite([]byte(fmt.Sprintf("Command '%s' does not exist\n", opts.Command)))
			if err != nil {
				logger.S().Error(err)
			}
		} else {
			result := cmd.Execute(opts.Args)
			logger.S().Debugf("sending data to the client [%s]", conn.RemoteAddr().String())
			err := conn.AsyncWrite([]byte(fmt.Sprintf("%s\n", result)))
			if err != nil {
				logger.S().Error(err)
			}

			if opts.Mode[0] == command.PersistMode {
				if opts.Op == command.FLUSHALL {
					logger.S().Debugf("received a flushall command")
					err := s.aof.Truncate()
					if err != nil {
						logger.S().Error(err)
					}
				} else {
					logger.S().Debugf("received a persist-enabled command: %s", command.SETEX.String())
					s.aof.Write(opts.Full)
					logger.S().Debug("enqueued the command into the persistor queue")
				}
			}
		}
	})
	if err != nil {
		logger.S().Error(err)
	}

	return
}

func (s *Server) OnShutdown(svr gnet.Server) {
	logger.S().Info("shutting down server...")

	s.connectedClients.Range(func(key, value interface{}) bool {
		c := value.(gnet.Conn)
		c.Close()
		s.connectedClients.Delete(key)
		return true
	})

	err := s.aof.Flush()
	if err != nil {
		logger.S().Error(err)
	}

	s.aof.Close()
}

func (s *Server) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
	logger.S().Debugf("a new connection to the server has been opened [%s]", conn.RemoteAddr().String())
	s.stats.ConnectedCount++
	s.stats.TotalConnectionsCount++

	s.connectedClients.Store(conn.RemoteAddr().String(), conn)
	return
}

func (s *Server) OnClosed(conn gnet.Conn, err error) (action gnet.Action) {
	logger.S().Debugf("client closed the connection [%s]", conn.RemoteAddr().String())
	s.stats.ConnectedCount--
	s.connectedClients.Delete(conn.RemoteAddr().String())
	return
}

func (s *Server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	s.printLogo(srv)
	return
}

func (s *Server) startupMessage() {
	logger.S().Info("kvstore is starting...")
	logger.S().Infof("version=%s build=%s pid=%d", s.stats.Version, s.stats.Build, os.Getpid())
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

	fmt.Printf(logo, s.stats.Version, s.stats.TCPPort, os.Getpid())

	logger.S().Infof("listening on %s", srv.Addr)
	logger.S().Infof("started kvstore server")
	logger.S().Info("ready to accept connections.")
}
