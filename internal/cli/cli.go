package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/framecodec"
	"github.com/HotPotatoC/kvstore/internal/server/stats"
	"github.com/fatih/color"
	"github.com/peterh/liner"
	"github.com/smallnest/goframe"
)

// CLI represents the cli client
type CLI struct {
	terminal *liner.State
	conn     goframe.FrameConn
}

// New creates a new CLI client
func New(addr string) *CLI {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	connWithCodec := framecodec.NewLengthFieldBasedFrameCodecConn(
		framecodec.NewDefaultLengthFieldBasedFrameEncoderConfig(),
		framecodec.NewDefaultLengthFieldBasedFrameDecoderConfig(),
		conn)

	return &CLI{
		terminal: newTerminal(),
		conn:     connWithCodec,
	}
}

// Start runs the CLI client
func (c *CLI) Start() {
	defer c.terminal.Close()
	// Get server information on initial startup
	stats := c.getServerInformation()

	yellow := color.New(color.FgHiYellow).SprintFunc()
	fmt.Printf("🚀 Connected to kvstore %s:%s server!\n\n", yellow(stats.Version), yellow(stats.Build))
	for {
		input, err := c.terminal.Prompt(fmt.Sprintf("%s> ", c.conn.Conn().RemoteAddr().String()))
		if err != nil && !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}

		if input == "" {
			continue
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
			c.conn.Conn().Close()
			os.Exit(0)
		// This is where commands are parsed and processed inputs are sent to the server
		default:
			var op command.Op

			switch cmd {
			case command.SET.String():
				op = command.SET
			case command.SETEX.String():
				op = command.SETEX
			case command.GET.String():
				op = command.GET
			case command.DEL.String():
				op = command.DEL
			case command.LIST.String():
				op = command.LIST
			case command.KEYS.String():
				op = command.KEYS
			case command.FLUSH.String():
				op = command.FLUSH
			case command.INFO.String():
				op = command.INFO
			default:
				fmt.Printf("Command '%s' does not exist\n", cmd)
				continue
			}

			err = c.conn.WriteFrame([]byte(fmt.Sprintf("%s %s", op.String(), args)))
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}

			msg, err := c.conn.ReadFrame()
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}

			fmt.Print(string(msg))
		}

		// Write history into tmp direcotry
		f, err := os.Create(filepath.Join(os.TempDir(), ".kvstore-cli-history"))
		if err != nil {
			fmt.Printf("Failed creating history file %s\n", filepath.Join(os.TempDir(), ".kvstore-cli-history"))
		}
		c.terminal.WriteHistory(f)
		_ = f.Close()
	}
}

func (c *CLI) getServerInformation() *stats.Stats {
	var serverStats stats.Stats

	err := c.conn.WriteFrame(command.INFO.Bytes())
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	infoMessage, err := c.conn.ReadFrame()
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	err = json.Unmarshal(infoMessage, &serverStats)
	if err != nil {
		log.Fatal(err)
	}

	return &serverStats
}

func (c *CLI) parseCommand(input string) (string, string) {
	raw := strings.Fields(input)[0]
	cmd := strings.ToLower(
		strings.TrimSpace(raw))
	args := strings.TrimSpace(
		strings.TrimPrefix(input, raw))

	return cmd, args
}
