package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/packet"
	"github.com/HotPotatoC/kvstore/pkg/comm"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

type client struct {
	comm   *comm.Comm
	reader *bufio.Reader
}

func init() {
	log = logger.NewLogger()
}

func New(addr string) *client {
	comm, err := comm.New(addr)
	if err != nil {
		log.Fatal(err)
	}

	return &client{
		comm:   comm,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *client) StartCLI() {
	for {
		fmt.Printf("%s> ", c.comm.Connection().RemoteAddr().String())

		input, err := c.reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		t1 := time.Now()
		err = c.comm.Send(c.handle(input))
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		msg, _, err := c.comm.Read()
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		t2 := time.Now()
		fmt.Printf("%fs\n", t2.Sub(t1).Seconds())
		fmt.Println(string(msg))
	}
}

func (c *client) handle(data []byte) []byte {
	cmd := bytes.ToLower(
		bytes.TrimSpace(bytes.Split(data, []byte(" "))[0]))
	args := bytes.TrimSpace(
		bytes.TrimPrefix(data, cmd))

	p := new(packet.Packet)

	switch string(cmd) {
	case "set":
		p = packet.NewPacket(command.SET, args)
	case "get":
		p = packet.NewPacket(command.GET, args)
	case "del":
		p = packet.NewPacket(command.DEL, args)
	case "list":
		p = packet.NewPacket(command.LIST, args)
	}

	buffer, err := p.Encode()
	if err != nil {
		log.Fatal(err)
	}

	return buffer.Bytes()
}
