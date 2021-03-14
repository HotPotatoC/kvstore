package utils_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore/pkg/utils"
)

func TestByteCount(t *testing.T) {
	tc := []struct {
		n    uint64
		want string
	}{
		{1024, "1.0 kB"},
		{1023, "1023 B"},
		{75834, "74.1 kB"},
		{987654321, "941.9 MB"},
		{math.MaxInt8, "127 B"},
		{math.MaxInt16, "32.0 kB"},
		{math.MaxInt32, "2.0 GB"},
		{math.MaxInt64, "8.0 EB"},
	}
	for i, tt := range tc {
		t.Run(fmt.Sprintf("#%d ByteCount(%d)", i+1, tt.n), func(t *testing.T) {
			if utils.ByteCount(tt.n) != tt.want {
				t.Errorf("Failed TestByteCount -> Expected: %s | Got: %s", tt.want, utils.ByteCount(tt.n))
			}
		})
	}
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func TestFormatDuration(t *testing.T) {
	tc := []struct {
		td   time.Duration
		want string
	}{
		{time.Duration(60*time.Hour + 60*time.Minute + 60*time.Second), "2d13h1m0s"},
		{time.Duration(2*day + 2*time.Hour + 2*time.Minute), "2d2h2m0s"},
		{time.Duration(10*day + 26*time.Hour + 90*time.Minute), "11d3h30m0s"},
		{time.Duration(456*day + 2*time.Hour + 859*time.Minute), "1y91d16h19m0s"},
		{time.Duration(12*year + 365*day + 5*day + 5*time.Hour), "13y5d5h0m0s"},
	}
	for i, tt := range tc {
		t.Run(fmt.Sprintf("#%d FormatDuration(%d)", i+1, tt.td), func(t *testing.T) {
			if utils.FormatDuration(tt.td) != tt.want {
				t.Errorf("Failed TestFormatDuration -> Expected: %s | Got: %s", tt.want, utils.FormatDuration(tt.td))
			}
		})
	}
}
