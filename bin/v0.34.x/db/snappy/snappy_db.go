package snappy

import (
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	tmdb "github.com/tendermint/tm-db"
	"sync"
)

var (
	errIteratorNotSupported = errors.New("iterator unsupported")
)

const (
	CompatModeEnabled = iota
	CompatModeDisabled
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

		// if compat is enabled, try to see if this blob is snappy
		// (by checking if the first byte is '{')
		if s.compatMode == CompatModeEnabled {
			var isSnappyBlob = item[0] != '{'
			if isSnappyBlob {
				return snappy.Decode(nil, item)
			} else {
				s.mtx.Lock()
				// run item by Set() to encode & replace
				_ = s.db.Set(key, item)
				defer s.mtx.Unlock()

				return item, nil
			}

		} else {
			return snappy.Decode(nil, item)
		}
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

func (s *SnappyDB) Iterator(start, end []byte) (tmdb.Iterator, error) {
	return nil, errIteratorNotSupported
}

func (s *SnappyDB) ReverseIterator(start, end []byte) (tmdb.Iterator, error) {
	return nil, errIteratorNotSupported
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
