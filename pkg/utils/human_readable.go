package utils

import (
	"fmt"
	"strings"
	"time"
)

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

// ByteCount convert sizes in bytes into a human-readable string
// src: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCount(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// FormatDuration formats the given time duration into a human-readable string
// src: https://gist.github.com/harshavardhana/327e0577c4fed9211f65#gistcomment-2557682
func FormatDuration(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder
	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	fmt.Fprintf(&b, "%dd%s", days, d)

	return b.String()
}