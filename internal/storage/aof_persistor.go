package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HotPotatoC/kvstore/internal/logger"
)

const defaultAOFFileName = "kvstore-aof.log"

// AOFPersistor (Append-Only File Persistor) is used to persist the storage
// data in an Append-Only file
type AOFPersistor struct {
	file   *os.File
	writer *bufio.Writer
	mtx    sync.Mutex

	quit chan struct{}
}

// NewAOFPersistor creates a new append only file
// if no path is provided it will default to the current working directory
func NewAOFPersistor(path ...string) (*AOFPersistor, error) {
	var pathToFile string

	if len(path) < 1 {
		// Use current working directory if no path is provided
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		pathToFile = filepath.Join(wd, defaultAOFFileName)
	} else {
		p, err := filepath.Abs(filepath.Clean(path[0]))
		if err != nil {
			return nil, err
		}

		pathToFile = p
	}

	file, err := os.OpenFile(pathToFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	persistor := &AOFPersistor{
		file:   file,
		writer: bufio.NewWriter(file),
		quit:   make(chan struct{}),
	}

	return persistor, nil
}

// Run starts the AOF persistor infinite loop in the background and will
// flush data into the log file every given amount of time
func (aof *AOFPersistor) Run(after time.Duration) {
	go func() {
		t := time.NewTicker(after)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				aof.mtx.Lock()
				logger.L().Debug("aof writer tick")
				err := aof.writer.Flush()
				if err != nil {
					logger.L().Error("failed flushing buffered data into the aof log")
					logger.L().Errorf("reason: %v", err)
				}
				aof.mtx.Unlock()
			case <-aof.quit:
				logger.L().Debug("received aof log quit signal")
				return
			}
		}
	}()
}

// Add enqueue the given data into the AOF writer and will be
// written to the log file after a given amount of tick has passed
func (aof *AOFPersistor) Add(data string) {
	aof.mtx.Lock()
	fmt.Fprintln(aof.writer, data)
	logger.L().Debugf("wrote %s to the aof writer", data)
	aof.mtx.Unlock()
}

// Truncate completely clears the AOF log file content
func (aof *AOFPersistor) Truncate() error {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()
	logger.L().Debug("truncating aof log")
	return aof.file.Truncate(0)
}

// Close simply closes the file
func (aof *AOFPersistor) Close() error {
	logger.L().Debug("closing aof data persistor service")
	aof.quit <- struct{}{}
	aof.file.Sync()
	return aof.file.Close()
}
