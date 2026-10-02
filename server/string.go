package server

import (
	"bytes"
	"math"
	"strconv"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/client"
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

	var expiry time.Duration
	var expiryArg []byte
	var expiryOption, condition string
	for i := 2; i < c.Argc; i++ {
		option := string(bytes.ToLower(c.Argv[i]))
		switch option {
		case "ex", "px":
			if i+1 >= c.Argc || (expiryOption != "" && expiryOption != option) {
				res.Write(NewGenericError("syntax error"))
				return
			}
			expiryOption = option
			i++
			expiryArg = c.Argv[i]
		case "nx", "xx":
			if condition != "" && condition != option {
				res.Write(NewGenericError("syntax error"))
				return
			}
			condition = option
		default:
			res.Write(NewGenericError("syntax error"))
			return
		}
	}
	if expiryOption != "" {
		n, err := parseExpireInteger(expiryArg)
		if err != nil {
			res.Write(NewGenericError("value is not an integer or out of range"))
			return
		}
		multiplier := time.Second
		if expiryOption == "px" {
			multiplier = time.Millisecond
		}
		if n <= 0 || n > math.MaxInt64/int64(multiplier) {
			res.Write(NewGenericError("invalid expire time in 'set' command"))
			return
		}
		expiry = time.Duration(n) * multiplier
	}

	stored, err := c.DB.StoreBytesLimited(c.Argv[0], c.Argv[1], expiry, condition == "nx", condition == "xx")
	if err != nil {
		protocol.WriteError(res, "OOM command not allowed when used memory > 'maxmemory'.")
		return
	}
	if !stored {
		protocol.WriteNull(res)
		return
	}

	res.Write(protocol.RespOK)
}

// delCommand deletes a key from the database
func delCommand(c *client.Client, res *bytes.Buffer) {
	if c.Argc < 1 {
		res.Write(NewGenericError("wrong number of arguments for 'del' command"))
		return
	}

	var n int64
	for _, key := range c.Argv {
		n += c.DB.Delete(string(key))
	}

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
