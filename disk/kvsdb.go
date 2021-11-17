package disk

import (
	"io"
	"os"
	"path/filepath"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/vmihailenco/msgpack/v5"
)

// KVSDB is for persisting data to disk.
type KVSDB struct {
	file *os.File
}

// OpenKVSDB opens a kvsDB at the given path.
func OpenKVSDB(path ...string) (*KVSDB, error) {
	var pathToFile string

	if len(path) < 1 {
		// Use current working directory if no path is provided
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		pathToFile = filepath.Join(wd, "dump.kvsdb")
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

	return &KVSDB{
		file: file,
	}, nil
}

// Write writes the given data to the kvsDB.
func (db *KVSDB) Write(data *datastructure.Map) error {
	encoder := msgpack.NewEncoder(db.file)
	for _, item := range data.List() {
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}
	return nil
}

// Read reads the given data from the kvsDB.
func (db *KVSDB) Read() (*datastructure.Map, error) {
	var data datastructure.Map
	decoder := msgpack.NewDecoder(db.file)
	for {
		var item datastructure.Item
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		data.Store(&item)
	}
	return &data, nil
}

// Clear clears the kvsDB.
func (db *KVSDB) Clear() error {
	return db.file.Truncate(0)
}

// Close closes the kvsDB.
func (db *KVSDB) Close() error {
	return db.file.Close()
}
