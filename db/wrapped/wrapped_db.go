package wrapped

import (
	cmdb "github.com/cometbft/cometbft-db"
	tmdb "github.com/tendermint/tm-db"
)

// WrappedDB wraps the DB of tm-db and implements the cometbft-db DB interface.
var _ cmdb.DB = (*WrappedDB)(nil)

type WrappedDB struct {
	db tmdb.DB
}

func NewWrappedDB(db tmdb.DB) *WrappedDB {
	return &WrappedDB{
		db: db,
	}
}

// Get fetches the value of the given key, or nil if it does not exist.
// CONTRACT: key, value readonly []byte
func (wdb WrappedDB) Get(key []byte) ([]byte, error) { return wdb.db.Get(key) }

// Has checks if a key exists.
// CONTRACT: key, value readonly []byte
func (wdb WrappedDB) Has(key []byte) (bool, error) { return wdb.db.Has(key) }

// Set sets the value for the given key, replacing it if it already exists.
// CONTRACT: key, value readonly []byte
func (wdb WrappedDB) Set(key []byte, value []byte) error { return wdb.db.Set(key, value) }

// SetSync sets the value for the given key, and flushes it to storage before returning.
func (wdb WrappedDB) SetSync(key []byte, value []byte) error { return wdb.db.SetSync(key, value) }

// Delete deletes the key, or does nothing if the key does not exist.
// CONTRACT: key readonly []byte
func (wdb WrappedDB) Delete(key []byte) error { return wdb.db.Delete(key) }

// DeleteSync deletes the key, and flushes the delete to storage before returning.
func (wdb WrappedDB) DeleteSync(key []byte) error { return wdb.db.DeleteSync(key) }

// Iterator returns an iterator over a domain of keys, in ascending order. The caller must call
// Close when done. End is exclusive, and start must be less than end. A nil start iterates
// from the first key, and a nil end iterates to the last key (inclusive). Empty keys are not
// valid.
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// CONTRACT: start, end readonly []byte
func (wdb WrappedDB) Iterator(start, end []byte) (cmdb.Iterator, error) {
	return wdb.db.Iterator(start, end)
}

// ReverseIterator returns an iterator over a domain of keys, in descending order. The caller
// must call Close when done. End is exclusive, and start must be less than end. A nil end
// iterates from the last key (inclusive), and a nil start iterates to the first key (inclusive).
// Empty keys are not valid.
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
// CONTRACT: start, end readonly []byte
func (wdb WrappedDB) ReverseIterator(start, end []byte) (cmdb.Iterator, error) {
	return wdb.db.ReverseIterator(start, end)
}

// Close closes the database connection.
func (wdb WrappedDB) Close() error { return wdb.db.Close() }

// NewBatch creates a batch for atomic updates. The caller must call Batch.Close.
func (wdb WrappedDB) NewBatch() cmdb.Batch { return wdb.db.NewBatch() }

// Print is used for debugging.
func (wdb WrappedDB) Print() error { return wdb.db.Print() }

// Stats returns a map of property values for all keys and the size of the cache.
func (wdb WrappedDB) Stats() map[string]string { return wdb.db.Stats() }
