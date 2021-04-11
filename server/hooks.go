package server

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/HotPotatoC/kvstore/command"
	"github.com/HotPotatoC/kvstore/packet"
	"github.com/HotPotatoC/kvstore/pkg/comm"
)

func (s *Server) attachHooks() {
	s.server.OnConnected = s.onConnected
	s.server.OnDisconnected = s.onDisconnected
	s.server.OnMessage = s.onMessage
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
		comm.Send([]byte(fmt.Sprintf("Command '%s' does not exist\n", packet.Cmd.String())))
	} else {
		result := command.Execute(strings.Split(string(packet.Args), " "))
		comm.Send([]byte(fmt.Sprintf("%s\n", result)))
	}
}
