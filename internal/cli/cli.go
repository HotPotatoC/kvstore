package cli

import (
	"bufio"
	"bytes"
	"errors"
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

// CLI represents the cli client
type CLI struct {
	comm   *comm.Comm
	reader *bufio.Reader
}

func init() {
	log = logger.NewLogger()
}

// New creates a new CLI client
func New(addr string) *CLI {
	comm, err := comm.New(addr)
	if err != nil {
		log.Fatal(err)
	}

	return &CLI{
		comm:   comm,
		reader: bufio.NewReader(os.Stdin),
	}
}

// Start runs the CLI client
func (c *CLI) Start() {
start:
	for {
		fmt.Printf("%s> ", c.comm.Connection().RemoteAddr().String())

		input, err := c.reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		t1 := time.Now()
		preprocessed, err := c.preprocess(input)
		if err != nil {
			log.Error(err)
			continue start
		}

		err = c.comm.Send(preprocessed.Bytes())
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

func (c *CLI) preprocess(data []byte) (*bytes.Buffer, error) {
	var packet *packet.Packet
	var err error

	rawCmd := bytes.Split(data, []byte(" "))[0]
	cmd := bytes.ToLower(
		bytes.TrimSpace(rawCmd))
	args := bytes.TrimSpace(
		bytes.TrimPrefix(data, rawCmd))

	switch string(cmd) {
	case command.SET.String():
		if packet, err = c.set(args); err != nil {
			return nil, err
		}
	case command.GET.String():
		if packet, err = c.get(args); err != nil {
			return nil, err
		}
	case command.DEL.String():
		if packet, err = c.del(args); err != nil {
			return nil, err
		}
	case command.LIST.String():
		if packet, err = c.list(args); err != nil {
			return nil, err
		}
	case command.KEYS.String():
		if packet, err = c.keys(args); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Command '%s' does not exists", cmd)
	}

	buffer, err := packet.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed processing input: %v", err)
	}

	return buffer, nil
}

func (c *CLI) set(args []byte) (*packet.Packet, error) {
	if len(bytes.Split(args, []byte(" "))) < 2 {
		return nil, errors.New("Missing key/value arguments")
	}
	return packet.NewPacket(command.SET, args), nil
}

func (c *CLI) get(args []byte) (*packet.Packet, error) {
	if bytes.Compare(args, []byte("")) == 0 {
		return nil, errors.New("Missing key argument")
	}
	return packet.NewPacket(command.GET, args), nil
}

func (c *CLI) del(args []byte) (*packet.Packet, error) {
	if bytes.Compare(args, []byte("")) == 0 {
		return nil, errors.New("Missing key argument")
	}
	return packet.NewPacket(command.DEL, args), nil
}

func (c *CLI) list(args []byte) (*packet.Packet, error) {
	return packet.NewPacket(command.LIST, args), nil
}

func (c *CLI) keys(args []byte) (*packet.Packet, error) {
	return packet.NewPacket(command.KEYS, args), nil
}
