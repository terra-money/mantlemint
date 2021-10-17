package hld

import (
	"bytes"
	"fmt"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/common"
	"github.com/terra-money/mantlemint/lib"
	"sync"
)

const (
	LatestHeight  = 0
	InvalidHeight = 0

	debugKeyGet = iota
	debugKeySet
	debugKeyIterator
	debugKeyReverseIterator
	debugKeyGetResult
)

var HeightLimitedDelimiter = []byte{'@'}
var LatestHeightBuf = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

var (
	errInvalidWriteHeight = "invalid operation: writeHeight is not set"
)

var _ HLD = (*HeightLimitedDB)(nil)

type HeightLimitedDB struct {
	odb         HeightLimitEnabledDB
	readMutex   *sync.RWMutex
	writeMutex  *sync.RWMutex
	writeHeight int64
	readHeight  int64
	config      *HeightLimitedDBConfig

	// writeBatch HeightLimitEnabledBatch
}

type HeightLimitedDBConfig struct {
	Debug bool
}

func ApplyHeightLimitedDB(db HeightLimitEnabledDB, config *HeightLimitedDBConfig) *HeightLimitedDB {
	return &HeightLimitedDB{
		writeHeight: 0,
		readHeight:  0,
		readMutex:   new(sync.RWMutex),
		writeMutex:  new(sync.RWMutex),
		odb:         db,
		config:      config,
		// writeBatch:  nil,
	}
}

// SetReadHeight sets a target read height in the db driver.
// It acts differently if the db mode is writer or reader:
// - Reader uses readHeight as the max height at which the retrieved key/value pair is limited to,
//   allowing full block snapshot history
func (hld *HeightLimitedDB) SetReadHeight(height int64) {
	hld.readHeight = height
}

// ClearReadHeight sets internal readHeight to LatestHeight
func (hld *HeightLimitedDB) ClearReadHeight() int64 {
	lastKnownReadHeight := hld.readHeight
	hld.readHeight = LatestHeight
	return lastKnownReadHeight
}

// GetCurrentReadHeight gets the current readHeight
func (hld *HeightLimitedDB) GetCurrentReadHeight() int64 {
	return hld.readHeight
}

// SetWriteHeight sets a target write height in the db driver.
// - Writer uses writeHeight to append along with the key, so later when fetching with the driver
// you can find the latest known key/value pair before the writeHeight
func (hld *HeightLimitedDB) SetWriteHeight(height int64) {
	if height != 0 {
		hld.writeHeight = height
		// hld.writeBatch = hld.NewBatch()
	}
}

// ClearWriteHeight sets the next target write Height
// NOTE: evaluate the actual usage of it
func (hld *HeightLimitedDB) ClearWriteHeight() int64 {
	fmt.Println("!!! clearing write height...")
	lastKnownWriteHeight := hld.writeHeight
	hld.writeHeight = InvalidHeight
	// if batchErr := hld.writeBatch.Write(); batchErr != nil {
	// 	panic(batchErr)
	// }
	// hld.writeBatch = nil
	return lastKnownWriteHeight
}

// GetCurrentWriteHeight gets the current write height
func (hld *HeightLimitedDB) GetCurrentWriteHeight() int64 {
	return hld.writeHeight
}

// Get fetches the value of the given key, or nil if it does not exist.
// CONTRACT: key, value readonly []byte
func (hld *HeightLimitedDB) Get(key []byte) ([]byte, error) {
	return hld.odb.Get(hld.GetCurrentReadHeight(), key)
}

// Has checks if a key exists.
// CONTRACT: key, value readonly []byte
func (hld *HeightLimitedDB) Has(key []byte) (bool, error) {
	return hld.odb.Has(hld.GetCurrentReadHeight(), key)
}

// Set sets the value for the given key, replacing it if it already exists.
// CONTRACT: key, value readonly []byte
func (hld *HeightLimitedDB) Set(key []byte, value []byte) error {
	return hld.odb.Set(hld.writeHeight, key, value)
}

// SetSync sets the value for the given key, and flushes it to storage before returning.
func (hld *HeightLimitedDB) SetSync(key []byte, value []byte) error {
	return hld.Set(key, value)
}

// Delete deletes the key, or does nothing if the key does not exist.
// CONTRACT: key readonly []byte
// NOTE(mantlemint): delete should be marked?
func (hld *HeightLimitedDB) Delete(key []byte) error {
	return hld.odb.Delete(hld.writeHeight, key)
}

// DeleteSync deletes the key, and flushes the delete to storage before returning.
func (hld *HeightLimitedDB) DeleteSync(key []byte) error {
	return hld.Delete(key)
}

// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
// Close when done. End is exclusive, and start must be less than end. A nil start iterates
// from the first key, and a nil end iterates to the last key (inclusive).
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// CONTRACT: start, end readonly []byte
func (hld *HeightLimitedDB) Iterator(start, end []byte) (common.Iterator, error) {
	if bytes.Compare(start, end) == 0 {
		return nil, fmt.Errorf("invalid iterator operation; start_store_key=%v, end_store_key=%v", start, end)
	}

	return hld.odb.Iterator(hld.GetCurrentReadHeight(), start, end)
}

// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
// must call Close when done. End is exclusive, and start must be less than end. A nil end
// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// CONTRACT: start, end readonly []byte
func (hld *HeightLimitedDB) ReverseIterator(start, end []byte) (common.Iterator, error) {
	if bytes.Compare(start, end) == 0 {
		return nil, fmt.Errorf("invalid iterator operation; start_store_key=%v, end_store_key=%v", start, end)
	}

	return hld.odb.Iterator(hld.GetCurrentReadHeight(), start, end)
}

// Close closes the database connection.
func (hld *HeightLimitedDB) Close() error {
	return hld.odb.Close()
}

// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
func (hld *HeightLimitedDB) NewBatch() common.Batch {
	// if hld.writeBatch != nil {
	// 	// TODO: fix me
	// 	return hld.writeBatch
	// } else {
	// 	fmt.Println("!!! opening hld.batch", hld.GetCurrentWriteHeight())
	// 	hld.writeBatch = hld.odb.NewBatch(hld.GetCurrentWriteHeight())
	// 	return hld.writeBatch
	// }
	// 
	return hld.odb.NewBatch(hld.GetCurrentWriteHeight())
}


//
// func (hld *HeightLimitedDB) FlushBatch() error {
// 	hld.writeBatch
// }

// Print is used for debugging.
func (hld *HeightLimitedDB) Print() error {
	return hld.odb.Print()
}

// Stats returns a map of property values for all keys and the size of the cache.
func (hld *HeightLimitedDB) Stats() map[string]string {
	return hld.odb.Stats()
}

func (hld *HeightLimitedDB) Debug(debugType int, key []byte, value []byte) {
	if !hld.config.Debug {
		return
	}

	keyFamily := key[:len(key)-9]
	keyHeight := key[len(key)-8:]

	var debugPrefix string
	switch debugType {
	case debugKeyGet:
		debugPrefix = "get"
	case debugKeySet:
		debugPrefix = "set"
	case debugKeyIterator:
		debugPrefix = "get/it"
	case debugKeyReverseIterator:
		debugPrefix = "get/rit"

	case debugKeyGetResult:
		debugPrefix = "get/response"
	}

	var actualKeyHeight int64
	if bytes.Compare(keyHeight, LatestHeightBuf) == 0 {
		actualKeyHeight = -1
	} else {
		actualKeyHeight = int64(lib.BigEndianToUint(keyHeight))
	}

	fmt.Printf("<%s @ %d> %s", debugPrefix, actualKeyHeight, keyFamily)
	fmt.Printf("\n")
}
