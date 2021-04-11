package tlv

// ByteSize represents the size of a field in the message
type ByteSize int

const (
	// OneByte represents the size of integer 1
	OneByte ByteSize = 0x1
	// TwoBytes represents the size of integer 2
	TwoBytes ByteSize = 0x2
	// FourBytes represents the size of integer 4
	FourBytes ByteSize = 0x4
	// EightBytes represents the size of integer 8
	EightBytes ByteSize = 0x8
)
