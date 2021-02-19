package main

import "github.com/HotPotatoC/kvstore/internal/client"

func main() {
	client := client.New("0.0.0.0:7275")

	client.StartCLI()
}
