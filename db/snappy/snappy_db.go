package snappy

import (
	"encoding/json"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	tmdb "github.com/tendermint/tm-db"
	"sync"
)

const (
	CompatModeEnabled = iota
	CompatModeDisabled
)

var (
	errIteratorNotSupported = errors.New("iterator unsupported")
	errUnknownData          = errors.New("unknown format")
)

var _ tmdb.DB = (*SnappyDB)(nil)

// SnappyDB implements a tmdb.DB overlay with snappy compression/decompression
// Iterator is NOT supported -- main purpose of this library is to support indexer.db,
// which never makes use of iterators anyway
// NOTE: implement when needed
// NOTE2: monitor mem pressure, optimize by pre-allocating dst buf when there is bottleneck
type SnappyDB struct {
	db         tmdb.DB
	mtx        *sync.Mutex
	compatMode int
}

func NewSnappyDB(db tmdb.DB, compatMode int) *SnappyDB {
	return &SnappyDB{
		mtx:        new(sync.Mutex),
		db:         db,
		compatMode: compatMode,
	}
}

func (s *SnappyDB) Get(key []byte) ([]byte, error) {
	if item, err := s.db.Get(key); err != nil {
		return nil, err
	} else if item == nil && err == nil {
		return nil, nil
	} else {
		decoded, decodeErr := snappy.Decode(nil, item)

		// if snappy decode fails, try to replace the underlying
		// only recover & replace when the blob is a valid json
		if s.compatMode == CompatModeEnabled {
			if decodeErr != nil {
				if json.Valid(item) {
					s.mtx.Lock()
					// run item by Set() to encode & replace
					_ = s.db.Set(key, item)
					defer s.mtx.Unlock()

					return item, nil
				} else {
					return nil, errUnknownData
				}
			} else {
				return decoded, nil
			}
		}

		return decoded, decodeErr
	}
}

func (s *SnappyDB) Has(key []byte) (bool, error) {
	return s.db.Has(key)
}

func (s *SnappyDB) Set(key []byte, value []byte) error {
	return s.db.Set(key, snappy.Encode(nil, value))
}

func (s *SnappyDB) SetSync(key []byte, value []byte) error {
	return s.Set(key, value)
}

func (s *SnappyDB) Delete(key []byte) error {
	return s.db.Delete(key)
}

func (s *SnappyDB) DeleteSync(key []byte) error {
	return s.Delete(key)
}

func (s *SnappyDB) Iterator(start, end []byte) (iter tmdb.Iterator, err error) {
	iter, err = s.db.Iterator(start, end)
	if err != nil {
		return iter, err
	}
	return NewSnappyIterator(iter, s.compatMode), nil
}

func (s *SnappyDB) ReverseIterator(start, end []byte) (iter tmdb.Iterator, err error) {
	iter, err = s.db.ReverseIterator(start, end)
	if err != nil {
		return iter, err
	}
	return NewSnappyIterator(iter, s.compatMode), nil
}

func (s *SnappyDB) Close() error {
	return s.db.Close()
}

func (s *SnappyDB) NewBatch() tmdb.Batch {
	return NewSnappyBatch(s.db.NewBatch())
}

func (s *SnappyDB) Print() error {
	return s.db.Print()
}

func (s *SnappyDB) Stats() map[string]string {
	return s.db.Stats()
}
