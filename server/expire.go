package server

import (
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/common"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

type unit uint8

const (
	unitSeconds unit = iota
	unitMilliseconds
)

func expireGenericCommand(c *client.Client, u unit) {
	if c.Argc < 2 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for '" + c.Command + "' command"))
		return
	}

	key := string(c.Argv[0])
	n, err := common.ByteToInt(c.Argv[1])
	if err != nil {
		c.Conn.AsyncWrite(NewGenericError("invalid expire time"))
		return
	}

	var res int64
	if u == unitSeconds {
		res = c.DB.Expire(key, time.Duration(n)*time.Second)
	}
	if u == unitMilliseconds {
		res = c.DB.Expire(key, time.Duration(n)*time.Millisecond)
	}

	if res == 0 {
		c.Conn.AsyncWrite(protocol.MakeInteger(0))
	} else {
		c.Conn.AsyncWrite(protocol.MakeInteger(1))
	}
}

func expireCommand(c *client.Client) {
	expireGenericCommand(c, unitSeconds)
}

func pexpireCommand(c *client.Client) {
	expireGenericCommand(c, unitMilliseconds)
}

func ttlGenericCommand(c *client.Client, u unit) {
	if c.Argc != 1 {
		c.Conn.AsyncWrite(NewGenericError("wrong number of arguments for 'ttl' command"))
		return
	}

	key := string(c.Argv[0])

	item, ok := c.DB.Get(key)
	if !ok {
		c.Conn.AsyncWrite(protocol.MakeInteger(-2))
		return
	}

	// If the item does not expire, -1 is returned.
	if item.HasFlag(datastructure.ItemFlagExpireNX) {
		c.Conn.AsyncWrite(protocol.MakeInteger(-1))
		return
	}

	leftToLive := time.Until(item.ExpiresAt)
	if u == unitSeconds {
		c.Conn.AsyncWrite(protocol.MakeInteger(int64(leftToLive / time.Second)))
	}
	if u == unitMilliseconds {
		c.Conn.AsyncWrite(protocol.MakeInteger(int64(leftToLive / time.Millisecond)))
	}
}

func ttlCommand(c *client.Client) {
	ttlGenericCommand(c, unitSeconds)
}
func pttlCommand(c *client.Client) {
	ttlGenericCommand(c, unitMilliseconds)
}
