package main

import (
	"flag"

	"github.com/HotPotatoC/kvstore/internal/cli"
	"github.com/HotPotatoC/kvstore/internal/logger"
)

var (
	addr = flag.String("address", "127.0.0.1:7275", "KVStore target server address")
)

func init() {
	flag.StringVar(addr, "a", "127.0.0.1:7275", "KVStore target server address")
}

func main() {
	flag.Parse()

	logger.Init(false)

	cli := cli.New(*addr)

	cli.Start()
}
