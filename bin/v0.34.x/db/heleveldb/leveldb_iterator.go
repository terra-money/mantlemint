package heleveldb

import (
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
)

var _ hld.HeightLimitEnabledIterator = (*Iterator)(nil)

type Iterator struct {
	driver *Driver
	tmdb.Iterator

	maxHeight int64
	start     []byte
	end       []byte
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
		if exist, _ := i.driver.Has(i.maxHeight, i.Key()); exist {
			return true
		}
	}
	return false

}

func (i *Iterator) Value() (value []byte) {
	val, err := i.driver.Get(i.maxHeight, i.Key())
	if err != nil {
		panic(err)
	}
	return val
}
