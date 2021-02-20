package main

import (
	"flag"

	"github.com/HotPotatoC/kvstore/internal/server"
)

var (
	host = flag.String("host", "0.0.0.0", "KVStore server host")
	port = flag.Int("port", 7275, "KVStore server port")
)

func init() {
	flag.StringVar(host, "h", "0.0.0.0", "KVStore server host")
	flag.IntVar(port, "p", 7275, "KVStore server port")
}

func main() {
	flag.Parse()

	server := server.New()

	server.Start(*host, *port)
}
