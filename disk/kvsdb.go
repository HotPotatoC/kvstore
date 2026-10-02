package disk

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HotPotatoC/kvstore-rewrite/datastructure"
	"github.com/vmihailenco/msgpack/v5"
)

// KVSDB is for persisting data to disk.
type KVSDB struct {
	mu   sync.Mutex
	file *os.File
	path string
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

	file, err := os.OpenFile(pathToFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &KVSDB{
		file: file,
		path: pathToFile,
	}, nil
}

// Write writes the given data to the kvsDB.
func (db *KVSDB) Write(data *datastructure.Map) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.replace(data)
}

// replace publishes a complete snapshot before closing the previous file.
// The caller holds db.mu. A nil map creates an empty snapshot for Clear.
func (db *KVSDB) replace(data *datastructure.Map) error {
	if db.file == nil {
		return os.ErrClosed
	}
	info, err := db.file.Stat()
	if err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(db.path))
	if err != nil {
		return err
	}
	defer dir.Close()
	file, err := os.CreateTemp(filepath.Dir(db.path), "."+filepath.Base(db.path)+"-*")
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			file.Close()
			os.Remove(file.Name())
		}
	}()
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	writer := bufio.NewWriterSize(file, 64*1024)
	encoder := msgpack.NewEncoder(writer)
	if data != nil {
		for _, item := range data.List() {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), db.path); err != nil {
		return err
	}
	previous := db.file
	db.file = file
	committed = true
	// Even if directory Sync fails, the rename has happened, so keep the new
	// handle for subsequent reads and writes while reporting durability errors.
	return errors.Join(previous.Close(), dir.Sync())
}

// Read reads the given data from the kvsDB.
func (db *KVSDB) Read() (*datastructure.Map, error) {
	return db.ReadWithLimit(0)
}

// ReadWithLimit restores live items within the map's memory budget.
// Non-positive limits allow all items.
func (db *KVSDB) ReadWithLimit(maxBytes int64) (*datastructure.Map, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.file == nil {
		return nil, os.ErrClosed
	}
	if _, err := db.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	data := datastructure.NewMap()
	data.SetMaxMemory(maxBytes)
	restored := false
	defer func() {
		if !restored {
			data.Close()
		}
	}()
	decoder := msgpack.NewDecoder(db.file)
	for {
		var item datastructure.Item
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if item.HasFlag(datastructure.ItemFlagExpireXX) && !time.Now().Before(item.ExpiresAt) {
			continue
		}
		if _, err := data.StoreLimited(&item); err != nil {
			return nil, err
		}
	}
	restored = true
	return data, nil
}

// Clear clears the kvsDB.
func (db *KVSDB) Clear() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.replace(nil)
}

// Close closes the kvsDB.
func (db *KVSDB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.file == nil {
		return os.ErrClosed
	}
	err := db.file.Close()
	db.file = nil
	return err
}
