package common_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/common"
)

func TestGlobMatchBytes(t *testing.T) {
	for _, tc := range []struct {
		pattern, value string
		match          bool
	}{
		{"", "", true}, {"", "a", false}, {"*", "", true},
		{"**", "", false}, {"***", "", false}, {"?*", "", false},
		{"[^a]", "", false}, {"a*", "a", true}, {"a*", "abc", true},
		{"a*b*c", "axbyc", true}, {"*ab*abc", "ababc", true},
		{"a*b", "abc", false}, {"*a*b", "axaybz", false},
		{"a?c", "abc", true}, {"a?c", "ac", false},
		{"?", "é", false}, {"??", "é", true}, {"?", "\xff", true},
		{"*", "a/b\x00\xff", true}, {"a*b", "a/x/b", true},
		{"a?b", "a/b", true}, {"a/b", "a/b", true},
		{`a\*b`, "a*b", true}, {`a\*b`, "axxb", false},
		{`\?\[\]\\`, `?[]\`, true}, {`\`, `\`, true},
		{"[abc]", "b", true}, {"[abc]", "z", false},
		{"[^abc]", "z", true}, {"[^abc]", "b", false},
		{"[a-z]", "m", true}, {"[z-a]", "m", true}, {"[a-z]", "A", false},
		{`[\]]`, "]", true}, {`[\-]`, "-", true},
		{"[!a]", "!", true}, {"[!a]", "b", false},
		{"[]", "a", false}, {"[^]", "a", true},
		{"[abc", "b", true}, {"[", "[", false},
		{"[\x80-\xff]", "\xfe", true}, {"[\x00-\xff]", "\xfe", false},
		{"[\x00-\xff]", "\x00", true}, {"[\xff-\x7f]", "\x7f", true},
		{"a\x00?", "a\x00\xff", true},
		{"a\xff*", "a\xff/b", true}, {"a\xff*", "a\xfe/b", false},
	} {
		t.Run(tc.pattern+"/"+tc.value, func(t *testing.T) {
			work := 10000
			match, err := common.GlobMatch(tc.pattern, tc.value, &work)
			if err != nil || match != tc.match {
				t.Fatalf("match=%t err=%v want %t", match, err, tc.match)
			}
		})
	}
}

func TestGlobMatchWorkLimit(t *testing.T) {
	for _, tc := range []struct{ pattern, value string }{
		{"*" + strings.Repeat("a", 100) + "b", strings.Repeat("a", 1000)},
		{"[" + strings.Repeat("a", 1000) + "]", "a"},
		{strings.Repeat("*a", 1000) + "b", strings.Repeat("a", 1000)},
	} {
		work := 100
		if _, err := common.GlobMatch(tc.pattern, tc.value, &work); !errors.Is(err, common.ErrGlobWorkLimit) || work != 0 {
			t.Fatalf("work=%d err=%v", work, err)
		}
	}
	work := 0
	if _, err := common.GlobMatch("*", "anything", &work); !errors.Is(err, common.ErrGlobWorkLimit) {
		t.Fatalf("exhausted budget err=%v", err)
	}
	work = 3
	if match, err := common.GlobMatch("*", strings.Repeat("x", 1<<20), &work); err != nil || !match || work != 2 {
		t.Fatalf("star fast path match=%t work=%d err=%v", match, work, err)
	}
}

func TestGlobMatchSharedBudget(t *testing.T) {
	work := 10
	for i := 0; i < 10; i++ {
		match, err := common.GlobMatch("*", "key", &work)
		if err != nil || !match {
			t.Fatalf("match %d: %t/%v", i, match, err)
		}
	}
	if _, err := common.GlobMatch("*", "key", &work); !errors.Is(err, common.ErrGlobWorkLimit) {
		t.Fatalf("shared exhaustion err=%v", err)
	}
}
