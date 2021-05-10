package command_test

import (
	"fmt"

	"github.com/HotPotatoC/kvstore/internal/storage"
)

func NewTempDB(n int) storage.Store {
	db := storage.New()
	for i := 0; i < n; i++ {
		db.Set(fmt.Sprintf("k%d", i+1), fmt.Sprintf("v%d", i+1))
	}
	return db
}
