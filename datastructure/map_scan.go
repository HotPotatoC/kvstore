package datastructure

import (
	"errors"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/common"
)

const MaxScanCount = 1024

// Scan inspects at most count slots, including deleted slots and nonmatches.
// Cursors encode a shard and stable slot position. No per-client state or full
// key list is retained. Keys continuously present throughout a complete scan
// are returned; keys inserted or removed during the scan may be returned.
func (m *Map) Scan(cursor uint64, pattern string, count, maxBytes int) (uint64, []string, error) {
	if count <= 0 {
		count = 10
	}
	count = min(count, MaxScanCount)
	if maxBytes > 0 && maxBytes < 64 {
		return 0, nil, errors.New("ERR response exceeds output limit")
	}
	var keys []string
	if m.Len() == 0 {
		return 0, keys, nil
	}
	shardIndex, position := int(cursor&(numShards-1)), cursor>>8
	examined, remaining, matchWork := 0, maxBytes-64, 1<<20
	deadline := time.Now().Add(time.Millisecond)
	for shardIndex < numShards {
		s := &m.shards[shardIndex]
		s.mu.RLock()
		if position >= uint64(len(s.slots)) {
			s.mu.RUnlock()
			shardIndex++
			position = 0
			continue
		}
		current := position<<8 | uint64(shardIndex)
		if examined >= count || (examined > 0 && examined%32 == 0 && !time.Now().Before(deadline)) {
			s.mu.RUnlock()
			return current, keys, nil
		}
		item := s.slots[position].item
		// Decide visibility under the lock: a concurrent TTL extension must not
		// leave us checking an old deadline after the key was kept alive.
		live := item != nil && (!item.HasFlag(ItemFlagExpireXX) || time.Now().Before(item.ExpiresAt))
		s.mu.RUnlock()
		examined++
		if live {
			matches, err := common.GlobMatch(pattern, item.Key, &matchWork)
			if err != nil {
				return 0, nil, err
			}
			if matches {
				if maxBytes > 0 && len(item.Key)+32 > remaining {
					if len(keys) == 0 {
						return 0, nil, errors.New("ERR response exceeds output limit")
					}
					return current, keys, nil
				}
				keys = append(keys, item.Key)
				remaining -= len(item.Key) + 32
			}
		}
		position++
	}
	return 0, keys, nil
}
