package server

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
	"github.com/panjf2000/gnet/v2"
)

func BenchmarkOnTraffic(b *testing.B) {
	for _, name := range []string{"PING", "GET", "SET"} {
		for _, pipeline := range []int{1, 16} {
			b.Run(fmt.Sprintf("%s/P%d", name, pipeline), func(b *testing.B) {
				s := newTestServer(b)
				conn := openTestConn(s)
				conn.capture = false
				s.DB.Store(datastructure.NewItem("key", "value", 0))
				args := []string{name}
				if name != "PING" {
					args = append(args, "key")
				}
				if name == "SET" {
					args = append(args, "value")
				}
				data := bytes.Repeat(protocol.MakeCommand(args...), pipeline)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					conn.input = data
					conn.discards = conn.discards[:0]
					if action := s.OnTraffic(conn); action != gnet.None {
						b.Fatalf("action=%v", action)
					}
				}
				if conn.writeBytes == 0 {
					b.Fatal("no response")
				}
			})
		}
	}
}
