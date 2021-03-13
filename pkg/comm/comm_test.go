package comm_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/comm"
)

func TestComm(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:7222")
	if err != nil {
		t.Errorf("Expected nil but got an error: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()

			comm := comm.NewWithConn(conn)

			err = comm.Send([]byte("Comm1"))
			if err != nil {
				t.Errorf("Expected nil but got an error: %v", err)
			}

			data, _, err := comm.Read()
			if bytes.Equal([]byte("Comm2"), data) {
				t.Errorf("Expected 'Comm2' but got: %s", string(data))
			}
			return
		}
	}()

	time.Sleep(500 * time.Millisecond)
	comm, err := comm.New("127.0.0.1:7222")
	if err != nil {
		t.Errorf("Expected nil but got an error: %v", err)
	}

	data, _, err := comm.Read()
	if err != nil {
		t.Errorf("Expected nil but got an error: %v", err)
	}

	if bytes.Equal([]byte("Comm1"), data) {
		t.Errorf("Expected 'Comm1' but got: %s", string(data))
	}

	err = comm.Send([]byte("Comm2"))
	if err != nil {
		t.Errorf("Expected nil but got an error: %v", err)
	}
}
