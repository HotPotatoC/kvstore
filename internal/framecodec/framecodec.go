package framecodec

import (
	"encoding/binary"
	"net"

	"github.com/panjf2000/gnet"
	"github.com/smallnest/goframe"
)

func NewLengthFieldBasedFrameCodec(ec gnet.EncoderConfig, dc gnet.DecoderConfig) *gnet.LengthFieldBasedFrameCodec {
	return gnet.NewLengthFieldBasedFrameCodec(ec, dc)
}

func NewLengthFieldBasedFrameCodecConn(ec goframe.EncoderConfig, dc goframe.DecoderConfig, conn net.Conn) goframe.FrameConn {
	return goframe.NewLengthFieldBasedFrameConn(ec, dc, conn)
}

// NewDefaultLengthFieldBasedFrameEncoderConfig creates a new default goframe encoder config
func NewDefaultLengthFieldBasedFrameEncoderConfig() goframe.EncoderConfig {
	return goframe.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
}

// NewDefaultLengthFieldBasedFrameEncoderConfig creates a new default goframe decoder config
func NewDefaultLengthFieldBasedFrameDecoderConfig() goframe.DecoderConfig {
	return goframe.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}
}

// NewGNETDefaultLengthFieldBasedFrameEncoderConfig creates a new default gnet encoder config
func NewGNETDefaultLengthFieldBasedFrameEncoderConfig() gnet.EncoderConfig {
	return gnet.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
}

// NewGNETDefaultLengthFieldBasedFrameDecoderConfig creates a new default gnet decoder config
func NewGNETDefaultLengthFieldBasedFrameDecoderConfig() gnet.DecoderConfig {
	return gnet.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}
}
