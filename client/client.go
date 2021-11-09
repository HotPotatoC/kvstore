package client

import (
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/disk"
	"github.com/panjf2000/gnet"
)

// Client is a client that connects to a server.
type Client struct {
	// Conn is the underlying connection.
	Conn gnet.Conn
	// DB is the database client.
	DB *datastructure.Map
	// kvsDB is the file used to persist the data structure.
	KVSDB *disk.KVSDB
	// Argc is the number of arguments excluding the command.
	Argc int
	// Argv is the arguments excluding the command.
	Argv [][]byte
}
