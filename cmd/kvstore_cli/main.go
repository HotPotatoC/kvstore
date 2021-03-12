package main

import (
	"flag"

	"github.com/HotPotatoC/kvstore/cmd/kvstore_cli/internal/client"
)

var (
	addr = flag.String("address", "127.0.0.1:7275", "KVStore target server address")
)

func init() {
	flag.StringVar(addr, "a", "127.0.0.1:7275", "KVStore target server address")
}

func main() {
	flag.Parse()

	cli := client.New(*addr)

	cli.Start()
}
