package tlv

var (
	// DefaultTLVCodec default configured codec with type field size of two bytes
	// and len field size of four bytes
	DefaultTLVCodec = NewCodec(TwoBytes, FourBytes)
)

// Codec is the configuration of TLV encoding/decoding
type Codec struct {
	// TypeBytes defines the size of the type field
	TypeBytes ByteSize
	// LenBytes defines the size of the len field
	LenBytes ByteSize
}

// NewCodec creates a new codec for TLV encoding/decoding tasks
func NewCodec(typeBytes ByteSize, lenBytes ByteSize) *Codec {
	return &Codec{
		TypeBytes: typeBytes,
		LenBytes:  lenBytes,
	}
}
