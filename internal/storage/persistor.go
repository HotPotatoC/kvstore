package storage

import (
	"os"
	"time"
)

type Persistor interface {
	Run(time.Duration)
	Read() <-chan string
	Write(data string)
	Flush() error
	Truncate() error
	File() *os.File
	Close() error
}
