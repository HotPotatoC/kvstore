package tcp_test

import (
	"net"
	"testing"

	"github.com/HotPotatoC/kvstore/pkg/tcp"
)

func TestTCP(t *testing.T) {
	server := tcp.New()

	server.Listen("127.0.0.1", 7223)
	go func() {
		_, err := net.Dial("tcp", "127.0.0.1:7223")
		if err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	}()
}
