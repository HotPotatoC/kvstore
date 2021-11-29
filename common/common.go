package common

import (
	"errors"
)

var (
	ErrInvalidInt = errors.New("invalid int")
	ErrOverflow   = errors.New("overflow")
)

func ByteToInt(b []byte) (int64, error) {
	if len(b) == 0 {
		return 0, nil
	}

	negate := false
	if b[0] == '-' {
		negate = true
		b = b[1:]
	}

	var n int64
	for _, c := range b {
		if c < '0' || c > '9' {
			return 0, ErrInvalidInt
		}

		n *= 10
		n += int64(c - '0')
	}

	if negate {
		return -n, nil
	}
	return n, nil
}
