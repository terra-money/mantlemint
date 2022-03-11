package heleveldb

import (
	"bytes"

	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/hld"
)

var _ hld.HeightLimitEnabledIterator = (*Iterator)(nil)

type Iterator struct {
	driver *Driver
	tmdb.Iterator

	maxHeight int64
	start     []byte
	end       []byte

	// caching last validated key and value
	// since Valid and Value functions are expensive but called repeatedly
	lastValidKey   []byte
	lastValidValue []byte
}

func NewLevelDBIterator(d *Driver, maxHeight int64, start, end []byte) (*Iterator, error) {
	pdb := tmdb.NewPrefixDB(d.session, cKeysForIteratorPrefix)
	iter, err := pdb.Iterator(start, end)
	if err != nil {
		return nil, err
	}

	return &Iterator{
		driver:   d,
		Iterator: iter,

		maxHeight: maxHeight,
		start:     start,
		end:       end,
	}, nil
}
func NewLevelDBReverseIterator(d *Driver, maxHeight int64, start, end []byte) (*Iterator, error) {
	pdb := tmdb.NewPrefixDB(d.session, cKeysForIteratorPrefix)
	iter, err := pdb.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}

	return &Iterator{
		driver:   d,
		Iterator: iter,

		maxHeight: maxHeight,
		start:     start,
		end:       end,
	}, nil
}

func (i *Iterator) Domain() (start []byte, end []byte) {
	panic("implement me")
}

func (i *Iterator) Valid() bool {
	// filter out items with Deleted = true
	// it should return somewhere during the loop
	// otherwise iterator has reached the end without finding any record
	// with Delete = false, return false in such case.

	for ; i.Iterator.Valid(); i.Iterator.Next() {
		if bytes.Equal(i.lastValidKey, i.Key()) {
			return true
		}
		if val, _ := i.driver.Get(i.maxHeight, i.Key()); val != nil {
			i.lastValidKey = i.Key()
			i.lastValidValue = val
			return true
		}
	}
	return false

}

func (i *Iterator) Value() (value []byte) {
	if bytes.Equal(i.lastValidKey, i.Key()) {
		return i.lastValidValue
	}
	val, err := i.driver.Get(i.maxHeight, i.Key())
	if err != nil {
		panic(err)
	}
	return val
}
