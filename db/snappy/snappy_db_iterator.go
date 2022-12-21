package snappy

import (
	"github.com/golang/snappy"
	tmdb "github.com/tendermint/tm-db"
)

type SnappyIterator struct {
	iter       tmdb.Iterator
	compatMode int
	err        error
}

var (
	_ tmdb.Iterator = SnappyIterator{}
)

func NewSnappyIterator(iter tmdb.Iterator, compactMode int) SnappyIterator {
	return SnappyIterator{
		iter:       iter,
		compatMode: compactMode,
		err:        nil,
	}
}

func (s SnappyIterator) Domain() (start []byte, end []byte) {
	return s.iter.Domain()
}

func (s SnappyIterator) Valid() bool {
	return s.iter.Valid()
}

func (s SnappyIterator) Next() {
	s.iter.Next()
}

func (s SnappyIterator) Key() (key []byte) {
	return s.iter.Key()
}

func (s SnappyIterator) Value() (value []byte) {
	val := s.iter.Value()
	if val == nil {
		return nil
	}
	decoded, decodeErr := snappy.Decode(nil, val)
	s.err = decodeErr
	return decoded
}

func (s SnappyIterator) Error() error {
	if s.err != nil {
		return s.err
	}
	return s.iter.Error()
}

func (s SnappyIterator) Close() error {
	return s.iter.Close()
}
