package server

import (
	"bytes"
	"strconv"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

func scanCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc < 1 {
		res.Write(NewGenericError("wrong number of arguments for 'scan' command"))
		return
	}
	for _, b := range c.Argv[0] {
		if b < '0' || b > '9' {
			res.Write(NewGenericError("invalid cursor"))
			return
		}
	}
	cursor, err := strconv.ParseUint(string(c.Argv[0]), 10, 64)
	if err != nil {
		res.Write(NewGenericError("invalid cursor"))
		return
	}
	pattern, count := "*", 10
	for i := 1; i < c.Argc; i += 2 {
		if i+1 >= c.Argc {
			res.Write(NewGenericError("syntax error"))
			return
		}
		switch {
		case bytes.EqualFold(c.Argv[i], []byte("MATCH")):
			pattern = string(c.Argv[i+1])
		case bytes.EqualFold(c.Argv[i], []byte("COUNT")):
			n, err := parseExpireInteger(c.Argv[i+1])
			if err != nil || n <= 0 {
				res.Write(NewGenericError("value is not an integer or out of range"))
				return
			}
			if n > datastructure.MaxScanCount {
				n = datastructure.MaxScanCount
			}
			count = int(n)
		default:
			res.Write(NewGenericError("syntax error"))
			return
		}
	}
	next, keys, err := c.DB.Scan(cursor, pattern, count, c.OutputLimit)
	if err != nil {
		protocol.WriteError(res, err.Error())
		return
	}
	res.WriteString("*2\r\n")
	protocol.WriteBulkString(res, strconv.FormatUint(next, 10))
	res.WriteByte(protocol.Array)
	res.WriteString(strconv.Itoa(len(keys)))
	res.Write(protocol.CRLF)
	for _, key := range keys {
		protocol.WriteBulkString(res, key)
	}
}
