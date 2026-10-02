package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
	"github.com/panjf2000/gnet/v2"
	"github.com/spf13/viper"
)

type connectionState struct {
	client  *client.Client
	argv    [][]byte
	running bool       // owned by the connection's event loop
	mu      sync.Mutex // protects the worker result and closed state
	result  *bytes.Buffer
	closed  bool
}

type limits struct{ loops, commands, outputBudget, input, output, workers, queue, maxClients int }

func positiveConfig(key string, fallback int) int {
	if n := viper.GetInt(key); n > 0 {
		return n
	}
	return fallback
}
func serverLimits() limits {
	maxClients := 1000
	if viper.IsSet("server.maxclients") {
		maxClients = viper.GetInt("server.maxclients")
	}
	return limits{
		loops: positiveConfig("server.loops", 4), commands: positiveConfig("server.command_budget", 64), outputBudget: positiveConfig("server.output_budget", 64<<10),
		input: positiveConfig("server.max_pending_input", 16<<20), output: positiveConfig("server.max_pending_output", 16<<20),
		workers: positiveConfig("server.workers", 4), queue: positiveConfig("server.worker_queue", 16), maxClients: maxClients,
	}
}

var (
	respBufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
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
	// jobs bounds slow command concurrency and queue depth.
	jobs         chan func()
	workerMu     sync.Mutex
	stopping     atomic.Bool
	workers      sync.WaitGroup
	drainOnce    sync.Once
	shutdownOnce sync.Once
	limits       limits
	// nextClientID is the next monotonically increasing client ID.
	nextClientID  int64
	activeClients atomic.Int64

	*gnet.BuiltinEventEngine
	eng   gnet.Engine
	engMu sync.Mutex
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
	"info": {
		Name:        "info",
		Description: "Gets server info",
		Type:        command.Read,
		Proc:        infoCommand},
	"echo": {Name: "echo", Type: command.Read, Proc: echoCommand},
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
		Proc:        nil,
	},
	"info": {
		Name:        "info",
		Description: "Returns the info of the current connection",
		Type:        command.Read,
		Proc:        nil,
	},
	"list": {
		Name:        "list",
		Description: "Lists all connected clients",
		Type:        command.Read,
		Proc:        nil,
	},
	"kill": {
		Name:        "kill",
		Description: "Closes a given connection",
		Type:        command.Write,
		Proc:        nil,
	},
	"setname": {
		Name:        "setname",
		Description: "Sets the name of the current connection",
		Type:        command.Write,
		Proc:        nil,
	},
	"getname": {
		Name:        "getname",
		Description: "Gets the name of the current connection",
		Type:        command.Read,
		Proc:        nil,
	},
}

// New creates a new server.
func New() (*Server, error) {
	maxMemory := int64(256 << 20)
	if viper.IsSet("database.maxmemory") {
		maxMemory = viper.GetInt64("database.maxmemory")
	}
	if maxMemory < 0 || viper.GetInt("server.maxclients") < 0 {
		return nil, fmt.Errorf("maxmemory and maxclients must be non-negative")
	}
	kvsDB, err := disk.OpenKVSDB(viper.GetString("database.path"))
	if err != nil {
		return nil, err
	}

	db, err := kvsDB.ReadWithLimit(maxMemory)
	if err != nil {
		kvsDB.Close()
		return nil, err
	}

	s := &Server{
		PID:    os.Getpid(),
		DB:     db,
		kvsDB:  kvsDB,
		limits: serverLimits(),
	}

	s.startWorkers()

	return s, nil
}

// Run starts the server.
func (s *Server) Run() error {
	defer s.shutdown()
	addrs := append([]string(nil), viper.GetStringSlice("server.addrs")...)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("%s:%d", addrs[i], viper.GetInt("server.port"))
	}
	if len(addrs) == 0 {
		return fmt.Errorf("no server bind addresses configured")
	}
	return gnet.Rotate(s, addrs, gnet.WithNumEventLoop(s.limits.loops))
}

func (s *Server) Stop() {
	s.workerMu.Lock()
	s.stopping.Store(true)
	s.workerMu.Unlock()
	s.engMu.Lock()
	eng := s.eng
	s.engMu.Unlock()
	if err := eng.Stop(context.Background()); err != nil {
		logger.S().Error("failed to stop server: ", err)
	}
	s.shutdown()
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	s.workerMu.Lock()
	if s.stopping.Load() {
		s.workerMu.Unlock()
		return gnet.Shutdown
	}
	// Publish only engines that will start. gnet never marks an engine shut down
	// when OnBoot returns Shutdown, so Stop must not wait on such an engine.
	s.engMu.Lock()
	s.eng = eng
	s.engMu.Unlock()
	s.workerMu.Unlock()

	fmt.Println()
	fmt.Printf("kvstore %s (%d-Bit)\n", build.Version, 8*int(unsafe.Sizeof(int(0))))
	fmt.Printf("Port: %d\n", viper.GetInt("server.port"))
	fmt.Printf("PID: %d\n", s.PID)
	fmt.Println()
	logger.S().Info("🚀 Ready to accept connections")
	return
}

func (s *Server) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	if n := s.activeClients.Add(1); s.limits.maxClients > 0 && n > int64(s.limits.maxClients) {
		s.activeClients.Add(-1)
		return []byte("-ERR max number of clients reached\r\n"), gnet.Close
	}
	c := &client.Client{
		ID:          atomic.AddInt64(&s.nextClientID, 1),
		RemoteAddr:  conn.RemoteAddr().String(),
		OutputLimit: s.limits.output,
		Conn:        conn,
		DB:          s.DB,
		KVSDB:       s.kvsDB,
		CreateTime:  time.Now(),
	}

	c.AddFlag(client.FlagNone)
	conn.SetContext(&connectionState{client: c, argv: make([][]byte, 0, 8)})
	s.clients.Store(c.ID, c)
	logger.S().Debugf("a new connection to the server has been opened [%s]", conn.RemoteAddr().String())
	return
}

func (s *Server) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	logger.S().Debugf("client closed the connection [%s]", conn.RemoteAddr().String())

	if state, ok := conn.Context().(*connectionState); ok {
		state.mu.Lock()
		state.closed = true
		result := state.result
		state.result = nil
		state.mu.Unlock()
		if result != nil {
			respBufPool.Put(result)
		}
		clear(state.argv[:cap(state.argv)])
		if _, loaded := s.clients.LoadAndDelete(state.client.ID); loaded {
			s.activeClients.Add(-1)
		}
	}
	return
}

// gnet calls OnShutdown before its event loops exit. Drain slow work here;
// persist only after Run returns or Engine.Stop has waited for every loop.
func (s *Server) OnShutdown(_ gnet.Engine) { s.drainWorkers() }

func (s *Server) startWorkers() {
	s.jobs = make(chan func(), s.limits.queue)
	for i := 0; i < s.limits.workers; i++ {
		s.workers.Add(1)
		go func() {
			defer s.workers.Done()
			for job := range s.jobs {
				job()
			}
		}()
	}
}

func (s *Server) drainWorkers() {
	s.drainOnce.Do(func() {
		s.workerMu.Lock()
		s.stopping.Store(true)
		if s.jobs != nil {
			close(s.jobs)
		}
		s.workerMu.Unlock()
		s.workers.Wait()
	})
}

// shutdown finalizes persistence after all event loops and workers have stopped.
func (s *Server) shutdown() {
	s.drainWorkers()
	s.shutdownOnce.Do(func() {
		if s.DB != nil {
			defer s.DB.Close()
		}
		if s.kvsDB != nil {
			if err := s.kvsDB.Write(s.DB); err != nil {
				logger.S().Warn("failed saving db: ", err)
			}
			if err := s.kvsDB.Close(); err != nil {
				logger.S().Warn("failed closing db: ", err)
			}
		}
	})
}

func (s *Server) writeResponse(conn gnet.Conn, res *bytes.Buffer) bool {
	if res.Len() == 0 {
		return true
	}
	if res.Len() > s.limits.output-conn.OutboundBuffered() {
		return false
	}
	_, err := conn.Write(res.Bytes())
	return err == nil
}

type commandKind uint8

const (
	commandOther commandKind = iota
	commandGet
	commandSet
	commandDel
	commandPing
)

// Classify the common commands once for both worker selection and dispatch.
func classifyCommand(args [][]byte) commandKind {
	if len(args) == 0 {
		return commandOther
	}
	name := args[0]
	switch len(name) {
	case 3:
		switch name[0] | 0x20 {
		case 'g':
			if name[1]|0x20 == 'e' && name[2]|0x20 == 't' {
				return commandGet
			}
		case 's':
			if name[1]|0x20 == 'e' && name[2]|0x20 == 't' {
				return commandSet
			}
		case 'd':
			if name[1]|0x20 == 'e' && name[2]|0x20 == 'l' {
				return commandDel
			}
		}
	case 4:
		if name[0]|0x20 == 'p' && name[1]|0x20 == 'i' && name[2]|0x20 == 'n' && name[3]|0x20 == 'g' {
			return commandPing
		}
	}
	return commandOther
}

func slowCommand(args [][]byte, kind commandKind) bool {
	switch kind {
	case commandGet, commandSet, commandPing, commandDel:
		return false
	}
	if len(args) == 0 {
		return false
	}
	switch {
	case bytes.EqualFold(args[0], []byte("KEYS")), bytes.EqualFold(args[0], []byte("FLUSHALL")):
		return true
	case bytes.EqualFold(args[0], []byte("CLIENT")):
		return len(args) > 1 && (bytes.EqualFold(args[1], []byte("LIST")) || bytes.EqualFold(args[1], []byte("KILL")))
	}
	return false
}

func (s *Server) submitSlow(state *connectionState, args [][]byte, kind commandKind) bool {
	s.workerMu.Lock()
	defer s.workerMu.Unlock()
	if s.stopping.Load() {
		return false
	}
	// args is an owned copy; the connection input may be reused before execution.
	select {
	case s.jobs <- func() {
		state.mu.Lock()
		closed := state.closed
		state.mu.Unlock()
		if closed {
			return
		}
		res := respBufPool.Get().(*bytes.Buffer)
		res.Reset()
		s.processCommand(args, kind, state.client, res)
		state.mu.Lock()
		if state.closed {
			state.mu.Unlock()
			respBufPool.Put(res)
			return
		}
		state.result = res
		state.mu.Unlock()
		if err := state.client.Conn.Wake(nil); err != nil {
			state.client.Conn.Close()
		}
	}:
		return true
	default:
		return false
	}
}

func (s *Server) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	state, ok := conn.Context().(*connectionState)
	if !ok {
		return gnet.Close
	}
	if conn.InboundBuffered() > s.limits.input || conn.OutboundBuffered() > s.limits.output {
		return gnet.Close
	}
	if s.stopping.Load() {
		return gnet.Close
	}
	if state.running {
		state.mu.Lock()
		res := state.result
		state.result = nil
		state.mu.Unlock()
		if res == nil {
			return gnet.None
		}
		state.running = false
		wrote := s.writeResponse(conn, res)
		respBufPool.Put(res)
		if !wrote || state.client.HasFlag(client.FlagCloseASAP) {
			return gnet.Close
		}
	}
	size := conn.InboundBuffered()
	if size == 0 {
		return gnet.None
	}
	data, err := conn.Peek(size)
	if err != nil {
		return gnet.Close
	}
	res := respBufPool.Get().(*bytes.Buffer)
	res.Reset()
	defer respBufPool.Put(res)
	consumed, count := 0, 0
	yielded := false
	for consumed < len(data) {
		args, n, err := protocol.ParseCommand(data[consumed:], state.argv)
		if errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			protocol.WriteError(res, "ERR protocol error: "+err.Error())
			action = gnet.Close
			break
		}
		state.argv = args
		kind := classifyCommand(args)
		if slowCommand(args, kind) {
			owned := make([][]byte, len(args))
			for i := range args {
				owned[i] = bytes.Clone(args[i])
			}
			state.running = true
			state.client.SetBusy(true)
			if !s.submitSlow(state, owned, kind) {
				state.running = false
				state.client.SetBusy(false)
				protocol.WriteError(res, "ERR server overloaded")
			}
			consumed += n
			count++
			if state.running {
				break
			}
		} else {
			s.processCommand(args, kind, state.client, res)
			consumed += n
			count++
		}
		if state.client.HasFlag(client.FlagCloseASAP) {
			action = gnet.Close
			break
		}
		if count >= s.limits.commands || res.Len() >= s.limits.outputBudget {
			yielded = consumed < len(data)
			break
		}
	}
	clear(state.argv[:cap(state.argv)])
	if !s.writeResponse(conn, res) {
		return gnet.Close
	}
	if consumed > 0 {
		if _, err := conn.Discard(consumed); err != nil {
			return gnet.Close
		}
	}
	if yielded && action != gnet.Close && !state.running {
		if err := conn.Wake(nil); err != nil {
			return gnet.Close
		}
	}
	return action
}

func (s *Server) processCommand(rawCmd [][]byte, kind commandKind, c *client.Client, responseBuffer *bytes.Buffer) {
	if len(rawCmd) == 0 {
		responseBuffer.Write(protocol.MakeError("ERR malformed command"))
		return
	}
	recvCmdBytes := rawCmd[0]
	recvArgv := rawCmd[1:]
	defer func() { c.Argv = nil }()

	switch kind {
	case commandGet:
		c.Command = "get"
		c.Argv = recvArgv
		c.Argc = len(recvArgv)
		c.SetBusy(true)
		getCommand(c, responseBuffer)
		s.afterCommand(c, responseBuffer)
		return
	case commandSet:
		c.Command = "set"
		c.Argv = recvArgv
		c.Argc = len(recvArgv)
		c.SetBusy(true)
		setCommand(c, responseBuffer)
		s.afterCommand(c, responseBuffer)
		return
	case commandDel:
		c.Command = "del"
		c.Argv = recvArgv
		c.Argc = len(recvArgv)
		c.SetBusy(true)
		delCommand(c, responseBuffer)
		s.afterCommand(c, responseBuffer)
		return
	case commandPing:
		c.Command = "ping"
		c.Argv = recvArgv
		c.Argc = len(recvArgv)
		c.SetBusy(true)
		pingCommand(c, responseBuffer)
		s.afterCommand(c, responseBuffer)
		return
	}

	recvCmdBytes = bytes.Clone(recvCmdBytes)
	for i, ch := range recvCmdBytes {
		recvCmdBytes[i] = lowerCommandByte(ch)
	}
	cmd, ok := CommandTable[string(recvCmdBytes)]
	if !ok {
		responseBuffer.Write(protocol.MakeError(fmt.Sprintf("ERR unknown command '%s'", recvCmdBytes)))
		return
	}

	if cmd.SubCommands != nil {
		if len(recvArgv) == 0 {
			responseBuffer.Write(protocol.MakeError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", recvCmdBytes)))
			return
		}

		subCmdStr := string(bytes.ToLower(recvArgv[0]))
		subCmd, ok := cmd.SubCommands[subCmdStr]
		if !ok {
			responseBuffer.Write(protocol.MakeError(fmt.Sprintf("ERR unknown subcommand '%s' for '%s' command", subCmdStr, recvCmdBytes)))
			return
		}
		cmd = subCmd
	}

	c.Command = cmd.Name
	c.Argv = recvArgv
	c.Argc = len(recvArgv)

	c.SetBusy(true)

	if cmd.Proc == nil {
		s.clientCommand(c, responseBuffer)
	} else {
		cmd.Proc(c, responseBuffer)
	}
	s.afterCommand(c, responseBuffer)
}

func lowerCommandByte(ch byte) byte {
	if ch >= 'A' && ch <= 'Z' {
		return ch + ('a' - 'A')
	}
	return ch
}

// pingCommand handles ping command.
func pingCommand(c *client.Client, res *bytes.Buffer) {
	res.Write(protocol.RespPONG)
}

// flushallCommand clears all keys and values from the database.
// Also, it clears the database from disk.
func flushallCommand(c *client.Client, res *bytes.Buffer) {
	n := c.DB.Clear()
	if err := c.KVSDB.Clear(); err != nil {
		res.Write(NewGenericError(err.Error()))
		return
	}

	logger.S().Info("DB saved on disk")

	protocol.WriteInteger(res, n)
}

func commandCommand(c *client.Client, res *bytes.Buffer) {
	protocol.WriteError(res, "NOT_IMPLEMENTED")
}

func echoCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc != 1 {
		protocol.WriteError(res, "ERR wrong number of arguments for 'echo' command")
		return
	}
	protocol.WriteBulkString(res, string(c.Argv[0]))
}
