package heleveldb

import (
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
)

var _ hld.HeightLimitEnabledIterator = (*Iterator)(nil)

type Iterator struct {
	driver   *Driver
	iterator tmdb.Iterator

	maxHeight int64
	start     []byte
	end       []byte
}

func NewLevelDBIterator(d *Driver, maxHeight int64, start, end []byte) (*Iterator, error) {
	pdb := tmdb.NewPrefixDB(d.session, []byte(cOriginalDataPrefix))
	iter, err := pdb.Iterator(start, end)
	if err != nil {
		return nil, err
	}

	return &Iterator{
		driver:   d,
		iterator: iter,

		maxHeight: maxHeight,
		start:     start,
		end:       end,
	}, nil
}
func NewLevelDBReverseIterator(d *Driver, maxHeight int64, start, end []byte) (*Iterator, error) {
	pdb := tmdb.NewPrefixDB(d.session, []byte(cOriginalDataPrefix))
	iter, err := pdb.Iterator(start, end)
	if err != nil {
		return nil, err
	}

	return &Iterator{
		driver:   d,
		iterator: iter,

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

	for ; i.iterator.Valid(); i.iterator.Next() {
		if exist, _ := i.driver.Has(i.maxHeight, i.Key()); exist {
			return true
		}
	}
	return false

}

func (i *Iterator) Next() {
	i.iterator.Next()
}

func (i *Iterator) Key() (key []byte) {
	return i.iterator.Key()
}

func (i *Iterator) Value() (value []byte) {
	val, err := i.driver.Get(i.maxHeight, i.Key())
	if err != nil {
		panic(err)
	}
	return val
}

func (i *Iterator) Error() error {
	return i.iterator.Error()
}

func (i *Iterator) Close() error {
	return i.iterator.Close()
}
