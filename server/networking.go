package server

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/logger"
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
func (s *Server) killClient(c *client.Client, kct KillClientType, target interface{}) {
	s.pool.Submit(func() {
		delNum := 0
		switch kct {
		// Kill by the client ID
		case KillClientByID:
			logger.S().Debug("Killing client with ID: ", target)
			s.clients.Range(func(key, value interface{}) bool {
				if value.(*client.Client).ID == target {
					if value.(*client.Client).Flags&client.FlagBusy != 0 {
						// If the client is busy, mark it for close
						value.(*client.Client).Flags |= client.FlagCloseASAP
						delNum++
					} else {
						// If the client is not busy, kill it immediately
						value.(*client.Client).Conn.Close()
						s.clients.Delete(key)
						delNum++
					}
					return false
				}
				return true
			})

		// Kill by the client remote address (addr:port) or the client name
		case KillClientByAddr:
			logger.S().Debug("Killing client with IP address: ", target)
			targetClient, ok := s.clients.Load(target)
			if ok {
				targetClient.(*client.Client).Conn.Close()
				s.clients.Delete(target)
				delNum++
				c.Conn.AsyncWrite(protocol.MakeBool(true))
				return
			}
		// Kill by the client name
		case KillClientByName:
			logger.S().Debug("Killing client with name: ", target.(string))
			s.clients.Range(func(key, value interface{}) bool {
				if value.(*client.Client).Name == target {
					value.(*client.Client).Conn.Close()
					s.clients.Delete(key)
					delNum++
					return false
				}
				return true
			})
		}

		c.Conn.AsyncWrite(protocol.MakeBool(delNum > 0))
	})
}

// afterCommand is called after a command is executed.
// It reads the client flags and returns the client to the free state.
// It also checks if the client is in the closing state and if so, it closes the connection.
func (s *Server) afterCommand(c *client.Client) {
	// Clear the busy flag
	c.Flags &= ^client.FlagBusy
	// Set the FlagNone flag
	c.Flags |= client.FlagNone

	if c.Flags&client.FlagCloseASAP != 0 { // If the client is marked for close, close the connection
		c.Flags &= ^client.FlagCloseASAP
		s.killClient(c, KillClientByID, c.ID)
	}
}

// clientCommand is a command that handles client commands.
func clientCommand(c *client.Client) {
	subCmd := bytes.ToLower(c.Argv[0])

	// id sub-command
	if bytes.Equal(subCmd, []byte("id")) {
		clientIDSubCommand(c)
		return
	}

	// info sub-command
	if bytes.Equal(subCmd, []byte("info")) {
		clientInfoSubCommand(c)
		return
	}

	// list sub-command
	if bytes.Equal(subCmd, []byte("list")) {
		clientListSubCommand(c)
		return
	}

	// kill sub-command
	if bytes.Equal(subCmd, []byte("kill")) {
		clientKillSubCommand(c)
		return
	}

	// setname sub-command
	if bytes.Equal(subCmd, []byte("setname")) {
		clientSetNameSubCommand(c)
		return
	}

	// getname sub-command
	if bytes.Equal(subCmd, []byte("getname")) {
		clientGetNameSubCommand(c)
		return
	}
}

// clientIDSubCommand Returns the id of the current connection.
func clientIDSubCommand(c *client.Client) {
	c.Conn.AsyncWrite(protocol.MakeInteger(c.ID))
}

// clientInfoSubCommand Returns information and statistics about the server.
func clientInfoSubCommand(c *client.Client) {
	var s string

	s += "id=" + strconv.FormatInt(c.ID, 10)
	s += " addr=" + c.Conn.RemoteAddr().String()
	s += " name=" + c.Name
	s += " age=" + strconv.FormatInt(time.Now().Unix()-c.CreateTime.Unix(), 10)
	s += " flags=" + c.Flags.String()

	c.Conn.AsyncWrite(protocol.MakeBulkString(s))
}

// clientListSubCommand Returns the list of client connections.
func clientListSubCommand(c *client.Client) {
	var clientss []string

	server.clients.Range(func(key, value interface{}) bool {
		client := value.(*client.Client)
		var ss string
		ss += "id=" + strconv.FormatInt(client.ID, 10)
		ss += " addr=" + client.Conn.RemoteAddr().String()
		ss += " name=" + client.Name
		ss += " age=" + strconv.FormatInt(time.Now().Unix()-client.CreateTime.Unix(), 10)
		ss += " flags=" + client.Flags.String()
		ss += string(protocol.CRLF)

		clientss = append(clientss, ss)
		return true
	})

	c.Conn.AsyncWrite(protocol.MakeBulkString(strings.Join(clientss, "")))
}

// clientKillSubCommand Kills the connection of a client.
func clientKillSubCommand(c *client.Client) {

	if c.Argc < 2 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'kill' subcommand for 'client' command"))
		return
	}

	filter := bytes.ToLower(c.Argv[1])

	// Kill by the client ID
	if bytes.Equal(filter, []byte("id")) {
		id, err := strconv.ParseInt(string(c.Argv[2]), 10, 64)
		if err != nil {
			c.Conn.AsyncWrite(NewGenericError("invalid argument for 'kill' subcommand for 'client' command"))
			return
		}

		if id == c.ID {
			c.Conn.AsyncWrite(protocol.MakeBool(false))
			return
		}

		server.killClient(c, KillClientByID, id)
	}

	// Kill by the client remote address (addr:port)
	if bytes.Equal(filter, []byte("address")) {
		if bytes.Equal(c.Argv[2], []byte(c.Conn.RemoteAddr().String())) {
			c.Conn.AsyncWrite(protocol.MakeBool(false))
			return
		}

		server.killClient(c, KillClientByAddr, string(c.Argv[2]))
	}

	// Kill by the client name
	if bytes.Equal(filter, []byte("user")) {
		if bytes.Equal(c.Argv[2], []byte(c.Name)) {
			c.Conn.AsyncWrite(protocol.MakeBool(false))
			return
		}

		server.killClient(c, KillClientByName, string(c.Argv[2]))
	}
}

// clientSetNameSubCommand Sets the name of the client.
func clientSetNameSubCommand(c *client.Client) {
	if c.Argc < 2 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'setname' subcommand for 'client' command"))
		return
	}

	c.Name = string(c.Argv[1])
	c.Conn.AsyncWrite(protocol.MakeSimpleString("OK"))
}

// clientGetNameSubCommand Returns the name of the client.
func clientGetNameSubCommand(c *client.Client) {
	c.Conn.AsyncWrite(protocol.MakeBulkString(c.Name))
}
