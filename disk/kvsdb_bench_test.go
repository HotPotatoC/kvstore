package disk

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
)

func BenchmarkSnapshotWrite(b *testing.B) {
	for _, count := range []int{1000, 10000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			var data datastructure.Map
			value := strings.Repeat("x", 32)
			for i := 0; i < count; i++ {
				data.Store(datastructure.NewItem(fmt.Sprintf("key:%08d", i), value, 0))
			}
			db, err := OpenKVSDB(filepath.Join(b.TempDir(), "dump.kvsdb"))
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() { db.Close() })
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := db.Write(&data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
