package common_test

import (
	"math"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/common"
)

func Test_ByteToInt(t *testing.T) {
	tc := []struct {
		name     string
		b        []byte
		expected int64
	}{
		{"empty", []byte(""), 0},
		{"one", []byte("1"), 1},
		{"ten", []byte("10"), 10},
		{"hundred", []byte("100"), 100},
		{"negative", []byte("-1"), -1},
		{"negative ten", []byte("-10"), -10},
		{"negative hundred", []byte("-100"), -100},
		{"max int32", []byte("2147483647"), math.MaxInt32},
		{"max int64", []byte("9223372036854775807"), math.MaxInt64},
		{"min int32", []byte("-2147483648"), math.MinInt32},
		{"min int64", []byte("-9223372036854775808"), math.MinInt64},
		{"overflow", []byte("9223372036854775808"), math.MinInt64},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			i, err := common.ByteToInt(tc.b)
			if err != nil {
				t.Error(err)
			}
			if i != tc.expected {
				t.Errorf("expected: %d, got: %d", tc.expected, i)
			}
		})
	}
}
