package tlv

// Record is the representation of data that is
// encoded into a TLV format message
type Record struct {
	Payload []byte
	Type    uint
}

// NewRecord creates a new record for TLV encoding
func NewRecord(payload []byte, recordType uint) *Record {
	return &Record{
		Payload: payload,
		Type:    recordType,
	}
}
