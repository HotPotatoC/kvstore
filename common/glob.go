package common

import "errors"

var ErrGlobWorkLimit = errors.New("ERR scan match work limit exceeded")

func globStep(work *int) bool {
	if *work <= 0 {
		return false
	}
	*work -= 1
	return true
}

// GlobMatch matches Redis glob syntax on bytes, including binary strings and
// ordinary slashes. work is a shared budget for comparisons and pattern scans.
// Only the latest star needs retrying; no recursion or pattern copies are used.
func GlobMatch(pattern, value string, work *int) (bool, error) {
	if !globStep(work) {
		return false, ErrGlobWorkLimit
	}
	if pattern == "*" {
		return true, nil
	}
	// SCAN bypasses matching for "*". Other nonempty patterns do not match
	// an empty key in Redis's byte matcher, including repeated stars.
	if len(value) == 0 {
		return len(pattern) == 0, nil
	}
	p, v := 0, 0
	starPattern, starValue := -1, 0
	for v < len(value) {
		if !globStep(work) {
			return false, ErrGlobWorkLimit
		}
		if p < len(pattern) && pattern[p] == '*' {
			p++
			starPattern, starValue = p, v
			if p == len(pattern) {
				return true, nil
			}
			continue
		}
		next, match := p, false
		if p < len(pattern) {
			var err error
			next, match, err = globByte(pattern, p, value[v], work)
			if err != nil {
				return false, err
			}
		}
		if match {
			p, v = next, v+1
			continue
		}
		if starPattern < 0 {
			return false, nil
		}
		starValue++
		p, v = starPattern, starValue
	}
	for p < len(pattern) && pattern[p] == '*' {
		if !globStep(work) {
			return false, ErrGlobWorkLimit
		}
		p++
	}
	return p == len(pattern), nil
}

func globByte(pattern string, p int, value byte, work *int) (int, bool, error) {
	switch pattern[p] {
	case '?':
		return p + 1, true, nil
	case '\\':
		if p+1 < len(pattern) {
			p++
		}
		return p + 1, pattern[p] == value, nil
	case '[':
		p++
		negate := p < len(pattern) && pattern[p] == '^'
		if negate {
			p++
		}
		match := false
		for p < len(pattern) {
			if !globStep(work) {
				return p, false, ErrGlobWorkLimit
			}
			switch {
			case pattern[p] == '\\' && p+1 < len(pattern):
				p++
				match = match || pattern[p] == value
			case pattern[p] == ']':
				if negate {
					match = !match
				}
				return p + 1, match, nil
			case p+2 < len(pattern) && pattern[p+1] == '-':
				// Redis compares range endpoints as signed char values.
				start, end, c := int8(pattern[p]), int8(pattern[p+2]), int8(value)
				if start > end {
					start, end = end, start
				}
				match = match || (c >= start && c <= end)
				p += 2
			default:
				match = match || pattern[p] == value
			}
			p++
		}
		if negate {
			match = !match
		}
		return p, match, nil
	default:
		return p + 1, pattern[p] == value, nil
	}
}
