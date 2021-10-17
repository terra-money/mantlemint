package hld

import "github.com/terra-money/mantlemint-provider-v0.34.x/db/common"

type HLD interface {
	common.DB
	SetReadHeight(int64)
	ClearReadHeight() int64
	SetWriteHeight(int64)
	ClearWriteHeight() int64
}

type HeightLimitEnabledDB interface {
	// Get fetches the value of the given key, or nil if it does not exist.
	// CONTRACT: key, value readonly []byte
	Get(maxHeight int64, key []byte) ([]byte, error)

	// Has checks if a key exists.
	// CONTRACT: key, value readonly []byte
	Has(maxHeight int64, key []byte) (bool, error)

	// Set sets the value for the given key, replacing it if it already exists.
	// CONTRACT: key, value readonly []byte
	Set(atHeight int64, key, value []byte) error

	// SetSync sets the value for the given key, and flushes it to storage before returning.
	SetSync(atHeight int64, key, value []byte) error

	// Delete deletes the key, or does nothing if the key does not exist.
	// CONTRACT: key readonly []byte
	Delete(atHeight int64, key []byte) error

	// DeleteSync deletes the key, and flushes the delete to storage before returning.
	DeleteSync(atHeight int64, key []byte) error

	// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
	// Close when done. End is exclusive, and start must be less than end. A nil start iterates
	// from the first key, and a nil end iterates to the last key (inclusive).
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	Iterator(maxHeight int64, start, end []byte) (HeightLimitEnabledIterator, error)

	// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
	// must call Close when done. End is exclusive, and start must be less than end. A nil end
	// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// CONTRACT: start, end readonly []byte
	ReverseIterator(maxHeight int64, start, end []byte) (HeightLimitEnabledIterator, error)

	// Close closes the database connection.
	Close() error

	// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
	NewBatch(atHeight int64) HeightLimitEnabledBatch

	// Print is used for debugging.
	Print() error

	// Stats returns a map of property values for all keys and the size of the cache.
	Stats() map[string]string
}

type HeightLimitEnabledIterator interface {
	common.Iterator
}

type HeightLimitEnabledBatch interface {
	common.Batch
}

