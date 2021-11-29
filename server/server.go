package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/HotPotatoC/kvstore-rewrite/build"
	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/command"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/HotPotatoC/kvstore-rewrite/logger"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Server is the main server struct.
type Server struct {
	// PID of the server process.
	PID int
	// Port on which the server is listening.
	Port int
	// TLSPort on which the server is listening for TLS connections.
	TLSPort int
	// BindAddresses on which the server is listening.
	BindAddresses []string
	// DB is the data structure that the server uses to store key-value pairs.
	DB *datastructure.Map
	// Stats is the statistics of the server.
	Stats
	// kvsDB is the file used to persist the data structure.
	kvsDB *disk.KVSDB
	// clients is a map of all the clients connected to the server.
	clients sync.Map
	// pool is the pool of goroutines that the server uses to handle incoming
	// connections.
	pool *goroutine.Pool
	// nextClientID is the next monotonically increasing client ID.
	nextClientID int64

	*gnet.EventServer
	wg sync.WaitGroup
}

// server is the global server variable.
var server *Server

// CommandTable is the table of commands that the server supports.
var CommandTable = map[string]command.Command{
	"get": {
		Name:        "get",
		Description: "Gets a key's value",
		Type:        command.Read,
		Proc:        getCommand},
	"set": {
		Name:        "set",
		Description: "Sets a new key",
		Type:        command.Write,
		Proc:        setCommand},
	"del": {
		Name:        "del",
		Description: "Gets a key's value",
		Type:        command.Write,
		Proc:        delCommand},
	"keys": {
		Name:        "keys",
		Description: "Gets all keys",
		Type:        command.Read,
		Proc:        keysCommand},
	"info": {
		Name:        "info",
		Description: "Gets server info",
		Type:        command.Read,
		Proc:        infoCommand},
	"ping": {
		Name:        "ping",
		Description: "Pings the server",
		Type:        command.Read,
		Proc:        pingCommand},
	"flushall": {
		Name:        "flushall",
		Description: "Flushes all keys",
		Type:        command.Write,
		Proc:        flushallCommand},
	"command": {
		Name:        "command",
		Description: "Gets all commands",
		Type:        command.Read,
		Proc:        commandCommand},
	"expire": {
		Name:        "expire",
		Description: "Sets a key's expiration by seconds",
		Type:        command.Write,
		Proc:        expireCommand},
	"pexpire": {
		Name:        "pexpire",
		Description: "Sets a key's expiration by milliseconds",
		Type:        command.Write,
		Proc:        pexpireCommand},
	"ttl": {
		Name:        "ttl",
		Description: "Gets a key's expiration in seconds",
		Type:        command.Read,
		Proc:        ttlCommand},
	"pttl": {
		Name:        "pttl",
		Description: "Gets a key's expiration in milliseconds",
		Type:        command.Read,
		Proc:        pttlCommand},
	"client": {
		Name:        "client",
		SubCommands: clientSubCommands,
	},
}

var clientSubCommands = map[string]command.Command{
	"id": {
		Name:        "id",
		Description: "Returns the id of the current connection",
		Type:        command.Read,
		Proc:        clientCommand,
	},
	"info": {
		Name:        "info",
		Description: "Returns the info of the current connection",
		Type:        command.Read,
		Proc:        clientCommand,
	},
	"list": {
		Name:        "list",
		Description: "Lists all connected clients",
		Type:        command.Read,
		Proc:        clientCommand,
	},
	"kill": {
		Name:        "kill",
		Description: "Closes a given connection",
		Type:        command.Write,
		Proc:        clientCommand,
	},
	"setname": {
		Name:        "setname",
		Description: "Sets the name of the current connection",
		Type:        command.Write,
		Proc:        clientCommand,
	},
	"getname": {
		Name:        "getname",
		Description: "Gets the name of the current connection",
		Type:        command.Read,
		Proc:        clientCommand,
	},
}

// New creates a new server.
func New() (*Server, error) {
	kvsDB, err := disk.OpenKVSDB(viper.GetString("database.path"))
	if err != nil {
		return nil, err
	}

	db, err := kvsDB.Read()
	if err != nil {
		return nil, err
	}

	server = &Server{
		PID:   os.Getpid(),
		DB:    db,
		kvsDB: kvsDB,
		pool:  goroutine.Default(),
	}

	return server, nil
}

// Run starts the server.
func (s *Server) Run() error {
	for _, addr := range viper.GetStringSlice("server.addrs") {
		s.wg.Add(1)
		s.bindToAddress(addr)
	}

	s.wg.Wait()
	return nil
}

// Stop stops the server.
func (s *Server) Stop() {
	s.clients.Range(func(key, value interface{}) bool {
		c := value.(*client.Client)
		c.Conn.Close()
		s.clients.Delete(key)
		return true
	})

	s.pool.Release()

	for _, addr := range viper.GetStringSlice("server.addrs") {
		if err := gnet.Stop(context.Background(), fmt.Sprintf("%s:%d", addr, viper.GetInt("server.port"))); err != nil {
			logger.S().Error("failed to stop server", zap.String("addr", addr), err)
		}
	}
}

// React (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.React)
func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	data := append([]byte{}, frame...)
	c.ResetBuffer()

	err := s.pool.Submit(func() {
		s.handle(data, c)
	})
	if err != nil {
		logger.S().Error(err)
	}
	return
}

// OnInitComplete (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.OnInitComplete)
func (s *Server) OnInitComplete(svr gnet.Server) (action gnet.Action) {
	fmt.Println()
	fmt.Printf("kvstore %s (%d-Bit)\n", build.Version, 8*int(unsafe.Sizeof(int(0))))
	fmt.Printf("Port: %d\n", viper.GetInt("server.port"))
	fmt.Printf("PID: %d\n", s.PID)
	fmt.Println()
	logger.S().Info("🚀 Ready to accept connections")
	return
}

// OnOpened (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.OnOpened)
func (s *Server) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
	logger.S().Debugf("a new connection to the server has been opened [%s]", conn.RemoteAddr().String())

	s.clients.Store(conn.RemoteAddr().String(), &client.Client{
		ID:         atomic.AddInt64(&s.nextClientID, 1),
		Flags:      client.FlagNone,
		Conn:       conn,
		DB:         s.DB,
		KVSDB:      s.kvsDB,
		CreateTime: time.Now(),
	})

	return
}

// OnClosed (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.OnClosed)
func (s *Server) OnClosed(conn gnet.Conn, err error) (action gnet.Action) {
	logger.S().Debugf("client closed the connection [%s]", conn.RemoteAddr().String())

	s.clients.Delete(conn.RemoteAddr().String())
	return
}

// OnShutdown (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.OnShutdown)
func (s *Server) OnShutdown(svr gnet.Server) {
	if err := s.kvsDB.Write(s.DB); err != nil {
		logger.S().Warn("failed saving db: ", err)
	}

	if err := s.kvsDB.Close(); err != nil {
		logger.S().Warn("failed closing db: ", err)
	}

	logger.S().Info("DB saved on disk")
	logger.S().Info("server has been shut down")
}

// bindToAddress binds the server to the given address.
func (s *Server) bindToAddress(addr string) {
	logger.S().Debug("Binding to address: ", fmt.Sprintf("%s:%d", addr, viper.GetInt("server.port")))
	go func(addr string) {
		if err := gnet.Serve(s, fmt.Sprintf("%s:%d", addr, viper.GetInt("server.port"))); err != nil {
			logger.S().Errorf("Failed to bind to address %s: %s", addr, err)
			s.wg.Done()
			os.Exit(1)
		}
		s.wg.Done()
	}(addr)
}

// handle handles client requests.
func (s *Server) handle(data []byte, conn gnet.Conn) {
	recvCmd, recvArgv := s.parseObject(data)

	cmd, ok := CommandTable[string(recvCmd)]
	if !ok {
		conn.AsyncWrite(NewGenericError("unknown command '" + string(recvCmd) + "'"))
		return
	}

	if cmd.SubCommands != nil {
		if len(recvArgv) == 0 {
			conn.AsyncWrite(NewGenericError("wrong number of arguments for '" + string(recvCmd) + "' command"))
			return
		}

		subCmd, ok := cmd.SubCommands[string(recvArgv[0])]
		if !ok {
			conn.AsyncWrite(NewGenericError("unknown subcommand '" + string(recvArgv[0]) + "' for '" + string(recvCmd) + "' command"))
			return
		}

		cmd = subCmd
	}

	v, _ := s.clients.Load(conn.RemoteAddr().String())

	c := v.(*client.Client)
	c.Command = cmd.Name
	c.Argv = recvArgv
	c.Argc = len(recvArgv)

	// mark the client as busy
	c.RemoveFlag(client.FlagNone)
	c.AddFlag(client.FlagBusy)

	cmd.Proc(c)
	s.afterCommand(c)
}

// parseObject parses the resp3 object sent by the client.
// returns the command and the arguments.
func (s *Server) parseObject(data []byte) ([]byte, [][]byte) {
	reader := protocol.NewReader(bytes.NewReader(data))
	// TODO: Once generics are released, we should use it here.
	obj, err := reader.ReadObject()
	if err != nil {
		logger.S().Error(err)
		return nil, nil
	}

	recv := obj.([]interface{})

	recvCmd, rawRecvArgv := bytes.ToLower(recv[0].([]byte)), recv[1:]

	var recvArgv [][]byte
	for _, v := range rawRecvArgv {
		recvArgv = append(recvArgv, v.([]byte))
	}

	// Wrap args if it starts with a quote
	if len(recvArgv) > 0 && recvArgv[0][0] == '"' {
		recvArgv = command.WrapArgsFromQuotes(recvArgv)
	}

	return recvCmd, recvArgv
}

// pingCommand handles ping command.
func pingCommand(c *client.Client) {
	c.Conn.AsyncWrite(protocol.MakeSimpleString("PONG"))
}

// flushallCommand clears all keys and values from the database.
// Also, it clears the database from disk.
func flushallCommand(c *client.Client) {
	n := c.DB.Clear()
	if err := c.KVSDB.Clear(); err != nil {
		c.Conn.AsyncWrite(NewGenericError(err.Error()))
	}

	logger.S().Info("DB saved on disk")

	c.Conn.AsyncWrite(protocol.MakeInteger(n))
}

// commandCommand sends all registered commands to the client.
// TODO: implement this.
func commandCommand(c *client.Client) {
	c.Conn.AsyncWrite(protocol.MakeError("NOT_IMPLEMENTED"))
}
