package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"github.com/HotPotatoC/kvstore/pkg/comm"
	"github.com/HotPotatoC/kvstore/pkg/utils"
)

// CLI represents the cli client
type CLI struct {
	comm   *comm.Comm
	reader *bufio.Reader
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
	go func() {
	start:
		for {
			fmt.Printf("%s> ", c.comm.Connection().RemoteAddr().String())

			input, err := c.reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}

			raw := bytes.Split(input, []byte(" "))[0]
			cmd := bytes.ToLower(
				bytes.TrimSpace(raw))
			args := bytes.TrimSpace(
				bytes.TrimPrefix(input, raw))

			preprocessed, err := preprocess(cmd, args)
			if err != nil {
				log.Println(err)
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

			fmt.Print(string(msg))
		}
	}()

	<-utils.WaitForSignals(os.Interrupt, syscall.SIGTERM)
	c.comm.Connection().Close()
	os.Exit(0)
}
