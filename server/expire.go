package server

import (
	"bytes"
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

func expireGenericCommand(c *client.Client, res *bytes.Buffer, u unit) {
	if c.Argc < 2 {
		res.Write(NewGenericError("wrong number of arguments for '" + c.Command + "' command"))
		return
	}

	key := string(c.Argv[0])
	n, err := common.ByteToInt(c.Argv[1])
	if err != nil {
		res.Write(NewGenericError("invalid expire time"))
		return
	}

	var result int64
	if u == unitSeconds {
		result = c.DB.Expire(key, time.Duration(n)*time.Second)
	}
	if u == unitMilliseconds {
		result = c.DB.Expire(key, time.Duration(n)*time.Millisecond)
	}

	if result == 0 {
		res.Write(protocol.MakeInteger(0))
	} else {
		res.Write(protocol.MakeInteger(1))
	}
}

func expireCommand(c *client.Client, res *bytes.Buffer) {
	expireGenericCommand(c, res, unitSeconds)
}

func pexpireCommand(c *client.Client, res *bytes.Buffer) {
	expireGenericCommand(c, res, unitMilliseconds)
}

func ttlGenericCommand(c *client.Client, res *bytes.Buffer, u unit) {
	if c.Argc != 1 {
		res.Write(NewGenericError("wrong number of arguments for 'ttl' command"))
		return
	}

	key := string(c.Argv[0])

	item, ok := c.DB.Get(key)
	if !ok {
		res.Write(protocol.MakeInteger(-2))
		return
	}

	// If the item does not expire, -1 is returned.
	if item.HasFlag(datastructure.ItemFlagExpireNX) {
		res.Write(protocol.MakeInteger(-1))
		return
	}

	leftToLive := time.Until(item.ExpiresAt)
	if u == unitSeconds {
		res.Write(protocol.MakeInteger(int64(leftToLive / time.Second)))
	}
	if u == unitMilliseconds {
		res.Write(protocol.MakeInteger(int64(leftToLive / time.Millisecond)))
	}
}

func ttlCommand(c *client.Client, res *bytes.Buffer) {
	ttlGenericCommand(c, res, unitSeconds)
}

func pttlCommand(c *client.Client, res *bytes.Buffer) {
	ttlGenericCommand(c, res, unitMilliseconds)
}
