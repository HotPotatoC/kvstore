package server

import (
	"bytes"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

// getCommand gets the value of a key in the database
func getCommand(c *client.Client) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'get' command"))
		return
	}

	v, ok := c.DB.Get(string(c.Argv[0]))
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

	c.DB.Store(datastructure.NewItem(string(c.Argv[0]), string(c.Argv[1]), 0))

	c.Conn.AsyncWrite(protocol.MakeSimpleString("OK"))
}

// delCommand deletes a key from the database
func delCommand(c *client.Client) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'del' command"))
		return
	}

	n := c.DB.Delete(string(c.Argv[0]))

	c.Conn.AsyncWrite(protocol.MakeInteger(n))
}

// keysCommand returns all keys in the database
func keysCommand(c *client.Client) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'keys' command"))
		return
	}

	var dbKeys []string

	if bytes.Equal(c.Argv[0], []byte("*")) {
		dbKeys = c.DB.Keys()
	} else {
		dbKeys = c.DB.KeysWithPattern(string(c.Argv[0]))
	}

	var keys [][]byte

	for _, k := range dbKeys {
		keys = append(keys, protocol.MakeBulkString(k))
	}

	c.Conn.AsyncWrite(protocol.MakeArray(keys...))
}

// valuesCommand returns all values in the database
func valuesCommand(c *client.Client) {
	if c.Argc > 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'values' command"))
		return
	}

	var values [][]byte
	for _, v := range c.DB.Values() {
		values = append(values, protocol.MakeBulkString(v.Data.(string)))
	}

	c.Conn.AsyncWrite(protocol.MakeArray(values...))
}
