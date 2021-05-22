package storage_test

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HotPotatoC/kvstore/internal/storage"
)

var (
	tmpLogPath = filepath.Join(os.TempDir(), ".kvstore-aof-test.log")
)

func TestAOFPersistor(t *testing.T) {
	aof, err := storage.NewAOFPersistor(tmpLogPath)
	if err != nil {
		t.Errorf("Failed TestAOFPersistor -> Expected nil error | Got: %v", err)
	}

	go func() {
		aof.Run(2 * time.Second)
	}()

	tc := []string{
		"set 1 1",
		"set 2 2"}

	for _, tt := range tc {
		aof.Add(tt)
	}

	time.Sleep(2 * time.Second)

	err = aof.Close()
	if err != nil {
		t.Errorf("Failed TestAOFPersistor -> Expected nil error | Got: %v", err)
	}

	f, err := os.Open(tmpLogPath)
	if err != nil {
		t.Errorf("Failed TestAOFPersistor -> Expected nil error | Got: %v", err)
	}
	defer f.Close()

	var content []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		t.Errorf("Failed TestAOFPersistor -> Expected nil error | Got: %v", err)
	}

	for i, c := range content {
		if c != tc[i] {
			t.Errorf("Failed TestAOFPersistor -> Expected: %s | Got: %s", tc[i], c)
		}
	}

	err = os.Remove(tmpLogPath)
	if err != nil {
		t.Errorf("Failed TestAOFPersistor -> Expected nil error | Got: %v", err)
	}
}

func TestAOFPersistor_Truncate(t *testing.T) {
	aof, err := storage.NewAOFPersistor(tmpLogPath)
	if err != nil {
		t.Errorf("Failed TestAOFPersistor_Truncate -> Expected nil error | Got: %v", err)
	}

	go func() {
		aof.Run(2 * time.Second)
	}()

	tc := []string{
		"set 1 1",
		"set 2 2"}

	for _, tt := range tc {
		aof.Add(tt)
	}

	time.Sleep(2 * time.Second)

	err = aof.Truncate()
	if err != nil {
		t.Errorf("Failed TestAOFPersistor_Truncate -> Expected nil error | Got: %v", err)
	}

	err = aof.Close()
	if err != nil {
		t.Errorf("Failed TestAOFPersistor_Truncate -> Expected nil error | Got: %v", err)
	}

}
