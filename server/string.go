package server

import (
	"bytes"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/common"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

// getCommand gets the value of a key in the database
func getCommand(c *client.Client) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'get' command"))
		return
	}

	key := string(c.Argv[0])

	v, ok := c.DB.Get(key)
	if !ok {
		c.Conn.AsyncWrite(protocol.MakeNull())
		return
	}

	c.Conn.AsyncWrite(protocol.MakeBulkString(v.Data.(string)))
}

// setCommand sets the value of a key in the database
func setCommand(c *client.Client) {
	if c.Argc < 2 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'set' command"))
		return
	}

	key, value := string(c.Argv[0]), string(c.Argv[1])

	expiry := time.Duration(0)
	if c.Argc > 2 {
		option := string(bytes.ToLower(c.Argv[2]))
		switch {
		case (option == "ex" || option == "px"): // set expire time
			if c.Argc != 4 {
				c.Conn.AsyncWrite(NewGenericError("syntax error"))
				return
			}

			n, err := common.ByteToInt(c.Argv[3])
			if err != nil {
				c.Conn.AsyncWrite(NewGenericError("syntax error"))
				return
			}

			if option == "ex" { // set expire time in seconds
				expiry = time.Duration(n) * time.Second
			}

			if option == "px" { // set expire time in milliseconds
				expiry = time.Duration(n) * time.Millisecond
			}
		case option == "nx": // set only if key does not exist
			if c.Argc != 3 {
				c.Conn.AsyncWrite(NewGenericError("syntax error"))
				return
			}

			if c.DB.Exists(key) {
				c.Conn.AsyncWrite(protocol.MakeNull())
				return
			}
		case option == "xx": // set only if key exists
			if c.Argc != 3 {
				c.Conn.AsyncWrite(NewGenericError("syntax error"))
				return
			}

			if !c.DB.Exists(key) {
				c.Conn.AsyncWrite(protocol.MakeNull())
				return
			}
		}
	}

	c.DB.Store(datastructure.NewItem(key, value, expiry))

	c.Conn.AsyncWrite(protocol.MakeSimpleString("OK"))
}

// delCommand deletes a key from the database
func delCommand(c *client.Client) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'del' command"))
		return
	}

	key := string(c.Argv[0])

	n := c.DB.Delete(key)

	c.Conn.AsyncWrite(protocol.MakeInteger(n))
}

// keysCommand returns all keys in the database
func keysCommand(c *client.Client) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'keys' command"))
		return
	}

	var dbKeys []string

	pattern := string(c.Argv[0])

	if bytes.Equal(c.Argv[0], []byte("*")) {
		dbKeys = c.DB.Keys()
	} else {
		dbKeys = c.DB.KeysWithPattern(pattern)
	}

	var keys [][]byte

	for _, k := range dbKeys {
		keys = append(keys, protocol.MakeBulkString(k))
	}

	c.Conn.AsyncWrite(protocol.MakeArray(keys...))
}
