package storage

import (
	"os"
	"time"
)

type mockAOFPersistor struct{}

var _ Persistor = (*mockAOFPersistor)(nil)

// NewMockAOFPersistor creates a new mock AOFPersistor implementation
// used for when AOF service is disabled
func NewMockAOFPersistor(_ ...string) (Persistor, error) {
	return &mockAOFPersistor{}, nil
}

func (aof *mockAOFPersistor) Run(time.Duration)   {}
func (aof *mockAOFPersistor) Read() <-chan string { return nil }
func (aof *mockAOFPersistor) Write(data string)   {}
func (aof *mockAOFPersistor) Flush() error        { return nil }
func (aof *mockAOFPersistor) Truncate() error     { return nil }
func (aof *mockAOFPersistor) File() *os.File      { return nil }
func (aof *mockAOFPersistor) Close() error        { return nil }
