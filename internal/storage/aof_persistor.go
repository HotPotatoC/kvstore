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
type aofPersistor struct {
	file   *os.File
	writer *bufio.Writer
	reader *bufio.Scanner
	mtx    sync.Mutex

	quit chan struct{}
}

var _ Persistor = (*aofPersistor)(nil)

// NewAOFPersistor creates a new append only file
// if no path is provided it will default to the current working directory
func NewAOFPersistor(path ...string) (Persistor, error) {
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

	persistor := &aofPersistor{
		file:   file,
		writer: bufio.NewWriter(file),
		reader: bufio.NewScanner(file),
		quit:   make(chan struct{}),
	}

	return persistor, nil
}

// Run starts the AOF persistor infinite loop in the background and will
// flush data into the log file every given amount of time
func (aof *aofPersistor) Run(after time.Duration) {
	go func() {
		t := time.NewTicker(after)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				aof.mtx.Lock()
				logger.S().Info("AOF writer tick")
				err := aof.writer.Flush()
				if err != nil {
					logger.S().Error("failed flushing buffered data into the AOF log")
					logger.S().Errorf("reason: %v", err)
				}
				aof.mtx.Unlock()
			case <-aof.quit:
				logger.S().Debug("received AOF log quit signal")
				return
			}
		}
	}()
}

// Read reads the AOF log using bufio.Scanner per line
//
// Usage:
//	for data := range aof.Read() {
//		fmt.Println(data)
//	}
func (aof *aofPersistor) Read() <-chan string {
	l := make(chan string)
	go func() {
		aof.mtx.Lock()
		for aof.reader.Scan() {
			l <- aof.reader.Text()
		}
		close(l)
		aof.mtx.Unlock()
	}()
	return l
}

// Write enqueue the given data into the AOF writer and will be
// written to the log file after a given amount of tick has passed
func (aof *aofPersistor) Write(data string) {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()
	fmt.Fprintln(aof.writer, data)
	logger.S().Debugf("wrote %s to the AOF writer", data)
}

// Flush flushes buffered inputs manually
func (aof *aofPersistor) Flush() error {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()
	err := aof.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

// Truncate completely clears the AOF log file content
func (aof *aofPersistor) Truncate() error {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()
	logger.S().Debug("truncating AOF log")
	return aof.file.Truncate(0)
}

// File exposes the AOF Persistor log file os.File struct
func (aof *aofPersistor) File() *os.File {
	aof.mtx.Lock()
	defer aof.mtx.Unlock()
	return aof.file
}

// Close simply closes the AOF log file and stops
// the AOF loop
func (aof *aofPersistor) Close() error {
	logger.S().Debug("closing AOF data persistor service")
	aof.quit <- struct{}{}
	aof.file.Sync()
	return aof.file.Close()
}
