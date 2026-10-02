package server

import (
	"bytes"
	"strconv"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

// KillClientType is the type of the kill client command.
type KillClientType int8

const (
	// KillClientByID is the type of the kill client command by the client ID.
	KillClientByID KillClientType = iota
	// KillClientByAddr is the type of the kill client command by the client IP address.
	KillClientByAddr
	// KillClientByName is the type of the kill client command by the client name.
	KillClientByName
)

// killClient kills the client with the given target ID or remote address (addr:port) or the name of the client.
// After the client is killed, send either 0 (false) or 1 (true) to the client.
func (s *Server) killClient(c *client.Client, res *bytes.Buffer, kct KillClientType, target any) {
	killed := false
	s.clients.Range(func(_ any, value any) bool {
		targetClient := value.(*client.Client)
		match := false
		switch kct {
		case KillClientByID:
			match = targetClient.ID == target
		case KillClientByAddr:
			match = targetClient.RemoteAddr == target
		case KillClientByName:
			match = targetClient.Name() == target
		}
		if !match {
			return true
		}
		targetClient.AddFlag(client.FlagCloseASAP)
		if targetClient.HasFlag(client.FlagBusy) {
			targetClient.Conn.Wake(nil)
		} else {
			targetClient.Conn.Close()
		}
		killed = true
		return false
	})
	res.Write(protocol.MakeBool(killed))
}

func (s *Server) afterCommand(c *client.Client, _ *bytes.Buffer) {
	c.RemoveFlag(client.FlagBusy)
	c.AddFlag(client.FlagNone)
}

// clientCommand is a command that handles client commands.
func (s *Server) clientCommand(c *client.Client, res *bytes.Buffer) {
	subCmd := bytes.ToLower(c.Argv[0])

	// id sub-command
	if bytes.Equal(subCmd, []byte("id")) {
		clientIDSubCommand(c, res)
		return
	}

	// info sub-command
	if bytes.Equal(subCmd, []byte("info")) {
		clientInfoSubCommand(c, res)
		return
	}

	// list sub-command
	if bytes.Equal(subCmd, []byte("list")) {
		s.clientListSubCommand(c, res)
		return
	}

	// kill sub-command
	if bytes.Equal(subCmd, []byte("kill")) {
		s.clientKillSubCommand(c, res)
		return
	}

	// setname sub-command
	if bytes.Equal(subCmd, []byte("setname")) {
		clientSetNameSubCommand(c, res)
		return
	}

	// getname sub-command
	if bytes.Equal(subCmd, []byte("getname")) {
		clientGetNameSubCommand(c, res)
		return
	}
}

// clientIDSubCommand Returns the id of the current connection.
func clientIDSubCommand(c *client.Client, res *bytes.Buffer) {
	protocol.WriteInteger(res, c.ID)
}

// clientInfoSubCommand Returns information and statistics about the server.
func clientInfoSubCommand(c *client.Client, res *bytes.Buffer) {
	var s string

	s += "id=" + strconv.FormatInt(c.ID, 10)
	s += " addr=" + c.RemoteAddr
	s += " name=" + c.Name()
	s += " age=" + strconv.FormatInt(time.Now().Unix()-c.CreateTime.Unix(), 10)
	s += " flags=" + c.FlagsString()

	protocol.WriteBulkString(res, s)
}

// clientListSubCommand Returns the list of client connections.
func (s *Server) clientListSubCommand(c *client.Client, res *bytes.Buffer) {
	var list bytes.Buffer
	tooLarge := false

	s.clients.Range(func(key, value any) bool {
		client := value.(*client.Client)
		var ss string
		ss += "id=" + strconv.FormatInt(client.ID, 10)
		ss += " addr=" + client.RemoteAddr
		ss += " name=" + client.Name()
		ss += " age=" + strconv.FormatInt(time.Now().Unix()-client.CreateTime.Unix(), 10)
		ss += " flags=" + client.FlagsString()
		ss += string(protocol.CRLF)

		if c.OutputLimit > 0 && len(ss) > c.OutputLimit-list.Len()-32 {
			tooLarge = true
			return false
		}
		list.WriteString(ss)
		return true
	})

	if tooLarge {
		protocol.WriteError(res, "ERR response exceeds output limit")
		return
	}
	protocol.WriteBulkString(res, list.String())
}

// clientKillSubCommand Kills the connection of a client.
func (s *Server) clientKillSubCommand(c *client.Client, res *bytes.Buffer) {

	if c.Argc != 3 {
		res.Write(NewGenericError("wrong number of arguments for 'kill' subcommand for 'client' command"))
		return
	}

	filter := bytes.ToLower(c.Argv[1])

	// Kill by the client ID
	if bytes.Equal(filter, []byte("id")) {
		id, err := strconv.ParseInt(string(c.Argv[2]), 10, 64)
		if err != nil {
			res.Write(NewGenericError("invalid argument for 'kill' subcommand for 'client' command"))
			return
		}

		if id == c.ID {
			res.Write(protocol.MakeBool(false))
			return
		}

		s.killClient(c, res, KillClientByID, id)
		return
	}

	// Kill by the client remote address (addr:port)
	if bytes.Equal(filter, []byte("address")) {
		if bytes.Equal(c.Argv[2], []byte(c.RemoteAddr)) {
			res.Write(protocol.MakeBool(false))
			return
		}

		s.killClient(c, res, KillClientByAddr, string(c.Argv[2]))
		return
	}

	// Kill by the client name
	if bytes.Equal(filter, []byte("user")) {
		if bytes.Equal(c.Argv[2], []byte(c.Name())) {
			res.Write(protocol.MakeBool(false))
			return
		}

		s.killClient(c, res, KillClientByName, string(c.Argv[2]))
		return
	}
	protocol.WriteError(res, "ERR invalid CLIENT KILL filter")
}

// clientSetNameSubCommand Sets the name of the client.
func clientSetNameSubCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc < 2 {
		res.Write(NewGenericError("wrong number of arguments for 'setname' subcommand for 'client' command"))
		return
	}

	c.SetName(string(c.Argv[1]))
	res.Write(protocol.RespOK)
}

// clientGetNameSubCommand Returns the name of the client.
func clientGetNameSubCommand(c *client.Client, res *bytes.Buffer) {
	protocol.WriteBulkString(res, c.Name())
}
