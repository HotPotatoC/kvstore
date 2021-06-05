package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/HotPotatoC/kvstore/internal/command"
	"github.com/HotPotatoC/kvstore/internal/framecodec"
	"github.com/HotPotatoC/kvstore/internal/logger"
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
		logger.S().Fatal(err)
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
			logger.S().Fatal(err)
		}

		if input == "" {
			continue
		}

		c.terminal.AppendHistory(input)
		raw := strings.Fields(input)[0]
		cmd := strings.ToLower(
			strings.TrimSpace(raw))

		opts, err := command.Parse(input)

		switch cmd {
		// Displays all available commands with their args and description
		case "help":
			commands := []command.Op{
				command.SET,
				command.SETEX,
				command.GET,
				command.DEL,
				command.KEYS,
				command.FLUSHALL,
				command.INFO,
				command.PING,
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
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = c.conn.WriteFrame([]byte(opts.Full))
			if err != nil && err != io.EOF {
				logger.S().Fatal(err)
			}

			msg, err := c.conn.ReadFrame()
			if err != nil && err != io.EOF {
				logger.S().Fatal(err)
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
		logger.S().Fatal(err)
	}

	infoMessage, err := c.conn.ReadFrame()
	if err != nil && err != io.EOF {
		logger.S().Fatal(err)
	}

	err = json.Unmarshal(infoMessage, &serverStats)
	if err != nil {
		logger.S().Fatal(err)
	}

	return &serverStats
}
