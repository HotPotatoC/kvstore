package main

import (
	"github.com/HotPotatoC/kvstore/internal/server"
)

func main() {
	server := server.New()

	server.Start("0.0.0.0", 7275)
}
