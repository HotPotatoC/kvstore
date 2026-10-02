package server

import (
	"bytes"
	"strconv"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
	"github.com/HotPotatoC/kvstore-rewrite/common"
	"github.com/HotPotatoC/kvstore-rewrite/protocol"
)

// getCommand gets the value of a key in the database
func getCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc != 1 {
		res.Write(NewGenericError("wrong number of arguments for 'get' command"))
		return
	}

	key := string(c.Argv[0])

	v, ok := c.DB.Get(key)
	if !ok {
		protocol.WriteNull(res)
		return
	}

	if c.OutputLimit > 0 && len(v.Data) > c.OutputLimit-32 {
		protocol.WriteError(res, "ERR response exceeds output limit")
		return
	}
	protocol.WriteBulkString(res, v.Data)
}

// setCommand sets the value of a key in the database
func setCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc < 2 {
		res.Write(NewGenericError("wrong number of arguments for 'set' command"))
		return
	}

	expiry := time.Duration(0)
	condition := ""
	if c.Argc > 2 {
		option := string(bytes.ToLower(c.Argv[2]))
		switch {
		case (option == "ex" || option == "px"): // set expire time
			if c.Argc != 4 {
				res.Write(NewGenericError("syntax error"))
				return
			}

			n, err := common.ByteToInt(c.Argv[3])
			if err != nil {
				res.Write(NewGenericError("syntax error"))
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
				res.Write(NewGenericError("syntax error"))
				return
			}

			condition = "nx"
		case option == "xx": // set only if key exists
			if c.Argc != 3 {
				res.Write(NewGenericError("syntax error"))
				return
			}

			condition = "xx"
		}
	}

	if !c.DB.StoreBytes(c.Argv[0], c.Argv[1], expiry, condition == "nx", condition == "xx") {
		protocol.WriteNull(res)
		return
	}

	res.Write(protocol.RespOK)
}

// delCommand deletes a key from the database
func delCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc != 1 {
		res.Write(NewGenericError("wrong number of arguments for 'del' command"))
		return
	}

	key := string(c.Argv[0])

	n := c.DB.Delete(key)

	protocol.WriteInteger(res, n)
}

// keysCommand returns all keys in the database
func keysCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc != 1 {
		res.Write(NewGenericError("wrong number of arguments for 'keys' command"))
		return
	}

	pattern := string(c.Argv[0])
	dbKeys, exceeded := c.DB.KeysWithPatternLimit(pattern, c.OutputLimit)
	if exceeded {
		protocol.WriteError(res, "ERR response exceeds output limit")
		return
	}
	res.WriteByte(protocol.Array)
	res.WriteString(strconv.Itoa(len(dbKeys)))
	res.Write(protocol.CRLF)
	for _, key := range dbKeys {
		protocol.WriteBulkString(res, key)
	}
}
