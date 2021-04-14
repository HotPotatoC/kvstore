package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/HotPotatoC/kvstore/command"
	"github.com/HotPotatoC/kvstore/packet"
	"github.com/HotPotatoC/kvstore/pkg/comm"
	"github.com/HotPotatoC/kvstore/pkg/utils"
	"github.com/HotPotatoC/kvstore/server/stats"
	"github.com/fatih/color"
	"github.com/peterh/liner"
)

// CLI represents the cli client
type CLI struct {
	comm     *comm.Comm
	terminal *liner.State
}

// New creates a new CLI client
func New(addr string) *CLI {
	comm, err := comm.New(addr)
	if err != nil {
		log.Fatal(err)
	}

	return &CLI{
		comm:     comm,
		terminal: newTerminal(),
	}
}

// Start runs the CLI client
func (c *CLI) Start() {
	defer c.terminal.Close()
	go func() {
		// Get server information on initial startup
		stats := c.getServerInformation()

		c.printLogo()
		yellow := color.New(color.FgHiYellow).SprintFunc()
		fmt.Printf("🚀 Connected to kvstore %s:%s server!\n\n", yellow(stats.Version), yellow(stats.Build))
	start:
		for {
			input, err := c.terminal.Prompt(fmt.Sprintf("%s> ", c.comm.Connection().RemoteAddr().String()))
			if err != nil {
				if err == io.EOF {
					c.comm.Conn.Close()
					os.Exit(1)
				}
				log.Fatal(err)
			}

			c.terminal.AppendHistory(input)
			cmd, args := c.parseCommand(input)

			switch cmd {
			// Displays all available commands with their args and description
			case "help":
				commands := []command.Op{
					command.SET,
					command.SETEX,
					command.GET,
					command.DEL,
					command.LIST,
					command.KEYS,
					command.FLUSH,
					command.INFO,
				}
				color.Set(color.FgHiYellow)
				fmt.Println("NOTE: All commands are case-insensitive")
				color.Unset()

				commandColorize := color.New(color.FgBlue, color.Bold).SprintFunc()
				argsColorize := color.New(color.FgWhite, color.Faint).SprintFunc()
				for _, cmd := range commands {
					fmt.Printf("- %s %s \n%s\n\n",
						commandColorize(strings.ToUpper(cmd.String())),
						argsColorize(cmd.Args()),
						cmd.Description())
				}
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

				msg, n, err := c.comm.Read()
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}

				fmt.Print(string(msg[:n]))
			}

			// Write history into tmp direcotry
			f, err := os.Create(filepath.Join(os.TempDir(), ".kvstore-cli-history"))
			if err != nil {
				log.Printf("Failed creating history file %s\n", filepath.Join(os.TempDir(), ".kvstore-cli-history"))
			}
			c.terminal.WriteHistory(f)
			_ = f.Close()
		}
	}()

	<-utils.WaitForSignals(os.Interrupt, syscall.SIGTERM)
	c.comm.Connection().Close()
	os.Exit(0)
}

func (c *CLI) printLogo() {
	color.Set(color.FgHiBlue)
	fmt.Println(" _               _                            _ _\n" +
		"| |             | |                          | (_)\n" +
		"| | ____   _____| |_ ___  _ __ ___ ______ ___| |_\n" +
		"| |/ /\\ \\ / / __| __/ _ \\| '__/ _ \\______/ __| | |\n" +
		"|   <  \\ V /\\__ \\ || (_) | | |  __/     | (__| | |\n" +
		"|_|\\_\\  \\_/ |___/\\__\\___/|_|  \\___|      \\___|_|_|\n\n")
	color.Unset()
}

func (c *CLI) getServerInformation() *stats.Stats {
	serverStats := new(stats.Stats)
	infoPacket := packet.NewPacket(command.INFO, []byte(""))
	infoBuffer, err := infoPacket.Encode()
	if err != nil {
		log.Fatal(err)
	}

	err = c.comm.Send(infoBuffer.Bytes())
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	infoMessage, n, err := c.comm.Read()
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	err = json.Unmarshal(infoMessage[:n], &serverStats)
	if err != nil {
		log.Fatal(err)
	}

	return serverStats
}

func (c *CLI) parseCommand(input string) (string, string) {
	raw := strings.Fields(input)[0]
	cmd := strings.ToLower(
		strings.TrimSpace(raw))
	args := strings.TrimSpace(
		strings.TrimPrefix(input, raw))

	return cmd, args
}
