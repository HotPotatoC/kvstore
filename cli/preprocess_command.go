package cli

import (
	"bytes"
	"fmt"

	"github.com/HotPotatoC/kvstore/command"
	"github.com/HotPotatoC/kvstore/packet"
)

func preprocess(cmd, args []byte) (*bytes.Buffer, error) {
	var packet *packet.Packet
	var err error

	switch string(cmd) {
	case command.SET.String():
		if packet, err = set(args); err != nil {
			return nil, err
		}
	case command.SETEX.String():
		if packet, err = setex(args); err != nil {
			return nil, err
		}
	case command.GET.String():
		if packet, err = get(args); err != nil {
			return nil, err
		}
	case command.DEL.String():
		if packet, err = del(args); err != nil {
			return nil, err
		}
	case command.LIST.String():
		if packet, err = list(); err != nil {
			return nil, err
		}
	case command.KEYS.String():
		if packet, err = keys(); err != nil {
			return nil, err
		}
	case command.FLUSH.String():
		if packet, err = flush(); err != nil {
			return nil, err
		}
	case command.INFO.String():
		if packet, err = info(); err != nil {
			return nil, err
		}
	default:
		return nil, command.ErrCommandDoesNotExist
	}

	buffer, err := packet.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed processing input: %v", err)
	}

	return buffer, nil
}

func set(args []byte) (*packet.Packet, error) {
	if len(bytes.Split(args, []byte(" "))) < 2 {
		return nil, command.ErrMissingKeyValueArg
	}
	return packet.NewPacket(command.SET, args), nil
}

func setex(args []byte) (*packet.Packet, error) {
	if len(bytes.Split(args, []byte(" "))) < 3 {
		return nil, command.ErrInvalidArgLength
	}
	return packet.NewPacket(command.SETEX, args), nil
}

func get(args []byte) (*packet.Packet, error) {
	if bytes.Equal(args, []byte("")) {
		return nil, command.ErrMissingKeyArg
	}
	return packet.NewPacket(command.GET, args), nil
}

func del(args []byte) (*packet.Packet, error) {
	if bytes.Equal(args, []byte("")) {
		return nil, command.ErrMissingKeyArg
	}
	return packet.NewPacket(command.DEL, args), nil
}

func list() (*packet.Packet, error) {
	return packet.NewPacket(command.LIST, []byte("")), nil
}

func keys() (*packet.Packet, error) {
	return packet.NewPacket(command.KEYS, []byte("")), nil
}

func flush() (*packet.Packet, error) {
	return packet.NewPacket(command.FLUSH, []byte("")), nil
}

func info() (*packet.Packet, error) {
	return packet.NewPacket(command.INFO, []byte("")), nil
}
