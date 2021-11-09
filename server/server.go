package server

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/command"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/HotPotatoC/kvstore-rewrite/logger"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
	"github.com/spf13/viper"
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
	// kvsDB is the file used to persist the data structure.
	kvsDB *disk.KVSDB
	// clients is a map of all the clients connected to the server.
	clients sync.Map
	// pool is the pool of goroutines that the server uses to handle incoming
	// connections.
	pool *goroutine.Pool

	reader *protocol.Reader

	// Stats is the statistics of the server.
	Stats

	*gnet.EventServer
	wg sync.WaitGroup
}

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
	"values": {
		Name:        "values",
		Description: "Gets all values",
		Type:        command.Read,
		Proc:        valuesCommand},
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

	return &Server{
		PID:   os.Getpid(),
		DB:    db,
		kvsDB: kvsDB,
		pool:  goroutine.Default(),
	}, nil
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

// React (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.React)
func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	data := append([]byte{}, frame...)
	c.ResetBuffer()

	err := s.pool.Submit(func() {
		s.reader = protocol.NewReader(bytes.NewReader(data))
		s.handle(data, c)
	})
	if err != nil {
		logger.S().Error(err)
	}
	return
}

// OnOpened (see gnet docs: https://pkg.go.dev/github.com/panjf2000/gnet#EventServer.OnOpened)
func (s *Server) OnOpened(conn gnet.Conn) (out []byte, action gnet.Action) {
	logger.S().Debugf("a new connection to the server has been opened [%s]", conn.RemoteAddr().String())

	s.clients.Store(conn.RemoteAddr().String(), conn)
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
		logger.S().Error("failed saving db: ", err)
	}

	if err := s.kvsDB.Close(); err != nil {
		logger.S().Error("failed closing db: ", err)
	}

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
func (s *Server) handle(data []byte, c gnet.Conn) {
	v, err := s.reader.ReadObject()
	if err != nil {
		logger.S().Error(err)
		return
	}

	recv := v.([]interface{})
	recvCmd, rawRecvArgv := bytes.ToLower(recv[0].([]byte)), recv[1:]

	var recvArgv [][]byte
	for _, v := range rawRecvArgv {
		recvArgv = append(recvArgv, v.([]byte))
	}

	cmd, ok := CommandTable[string(recvCmd)]
	if !ok {
		c.AsyncWrite(NewGenericError("unknown command '" + string(recvCmd) + "'"))
		return
	}

	cmd.Proc(&client.Client{
		Conn:  c,
		DB:    s.DB,
		KVSDB: s.kvsDB,
		Argv:  recvArgv,
		Argc:  len(recvArgv),
	})
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
