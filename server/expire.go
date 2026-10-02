package server

import (
	"bytes"
	"math"
	"strconv"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

type unit uint8

const (
	unitSeconds unit = iota
	unitMilliseconds
)

func parseExpireInteger(arg []byte) (int64, error) {
	n, err := strconv.ParseInt(string(arg), 10, 64)
	if err == nil && strconv.FormatInt(n, 10) != string(arg) {
		err = strconv.ErrSyntax
	}
	return n, err
}

func expireGenericCommand(c *client.Client, res *bytes.Buffer, u unit) {
	if c.Argc != 2 {
		res.Write(NewGenericError("wrong number of arguments for '" + c.Command + "' command"))
		return
	}

	key := string(c.Argv[0])
	n, err := parseExpireInteger(c.Argv[1])
	if err != nil {
		res.Write(NewGenericError("value is not an integer or out of range"))
		return
	}

	multiplier := time.Second
	if u == unitMilliseconds {
		multiplier = time.Millisecond
	}
	if n > math.MaxInt64/int64(multiplier) || (u == unitSeconds && n < math.MinInt64/1000) {
		res.Write(NewGenericError("invalid expire time in '" + c.Command + "' command"))
		return
	}
	var expiry time.Duration
	if n > 0 {
		expiry = time.Duration(n) * multiplier
	}
	protocol.WriteInteger(res, c.DB.Expire(key, expiry))
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
		protocol.WriteInteger(res, -2)
		return
	}

	if item.HasFlag(datastructure.ItemFlagExpireNX) {
		protocol.WriteInteger(res, -1)
		return
	}

	leftToLive := time.Until(item.ExpiresAt)
	if u == unitSeconds {
		protocol.WriteInteger(res, int64(leftToLive/time.Second))
	}
	if u == unitMilliseconds {
		protocol.WriteInteger(res, int64(leftToLive/time.Millisecond))
	}
}

func ttlCommand(c *client.Client, res *bytes.Buffer) {
	ttlGenericCommand(c, res, unitSeconds)
}

func pttlCommand(c *client.Client, res *bytes.Buffer) {
	ttlGenericCommand(c, res, unitMilliseconds)
}
