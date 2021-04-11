package command_test

import (
	"fmt"

	"github.com/HotPotatoC/kvstore/pkg/hashtable"
)


func HTPopulate(n int) *hashtable.HashTable {
	ht := hashtable.New()
	for i := 0; i < n; i++ {
		ht.Set(fmt.Sprintf("k%d", i+1), fmt.Sprintf("v%d", i+1))
	}
	return ht
}