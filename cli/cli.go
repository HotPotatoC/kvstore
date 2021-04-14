package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/HotPotatoC/kvstore/command"
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

			switch string(cmd) {
			// Displays all available commands with their args and description
			case "help":
				var commands = []command.Op{
					command.SET,
					command.SETEX,
					command.GET,
					command.DEL,
					command.LIST,
					command.KEYS,
					command.FLUSH,
					command.INFO,
				}

				fmt.Println("NOTE: All commands are case-insensitive")
				for _, cmd := range commands {
					fmt.Printf("- %s %s \n%s\n\n",
						yellow(strings.ToUpper(cmd.String())),
						dimmed(cmd.Args()),
						cmd.Description())
				}
				continue start
			// Exit out of the CLI
			case "exit":
				c.comm.Conn.Close()
				os.Exit(0)
			// This is where commands are parsed and processed inputs are sent to the server
			default:
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
		}
	}()

	<-utils.WaitForSignals(os.Interrupt, syscall.SIGTERM)
	c.comm.Connection().Close()
	os.Exit(0)
}
