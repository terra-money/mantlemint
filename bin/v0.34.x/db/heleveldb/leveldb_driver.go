package heleveldb

import (
	"fmt"

	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"github.com/terra-money/mantlemint/lib"
)

type Driver struct {
	session *tmdb.GoLevelDB
}

//TODO: remove it!
func (d *Driver) Session() tmdb.DB {
	return d.session
}

func NewLevelDBDriver(config *DriverConfig) (*Driver, error) {
	ldb, err := tmdb.NewGoLevelDB(config.Name, config.Dir)
	if err != nil {
		return nil, err
	}

	return &Driver{
		session: ldb,
	}, nil
}

func (d *Driver) Get(maxHeight int64, key []byte) ([]byte, error) {
	var requestHeight = hld.Height(maxHeight).CurrentOrLatest().ToInt64()
	var requestHeightMin = hld.Height(0).CurrentOrNever().ToInt64()

	// check if requestHeightMin is
	if requestHeightMin > requestHeight {
		return nil, fmt.Errorf("invalid height")
	}

	pdb := tmdb.NewPrefixDB(d.session, prefixHeightSnapshotKey(key))

	heightEnd := lib.UintToBigEndian(uint64(requestHeight + 1))
	iter, _ := pdb.ReverseIterator(nil, heightEnd)
	defer iter.Close()

	// in tm-db@v0.6.4, key not found is NOT an error
	if !iter.Valid() {
		return nil, nil
	}

	value := iter.Value()
	deleted := value[0]
	if deleted == 1 {
		return nil, nil
	} else {
		if len(value) > 1 {
			return value[1:], nil
		}
		return []byte{}, nil
	}
}

func (d *Driver) Has(maxHeight int64, key []byte) (bool, error) {
	var requestHeight = hld.Height(maxHeight).CurrentOrLatest().ToInt64()
	var requestHeightMin = hld.Height(0).CurrentOrNever().ToInt64()

	// check if requestHeightMin is
	if requestHeightMin > requestHeight {
		return false, fmt.Errorf("invalid height")
	}

	pdb := tmdb.NewPrefixDB(d.session, prefixHeightSnapshotKey(key))

	heightEnd := lib.UintToBigEndian(uint64(requestHeight + 1))
	iter, _ := pdb.ReverseIterator(nil, heightEnd)
	defer iter.Close()

	// in tm-db@v0.6.4, key not found is NOT an error
	if !iter.Valid() {
		return false, nil
	}

	deleted := iter.Value()[0]

	if deleted == 1 {
		return false, nil
	} else {
		return true, nil
	}
}

func (d *Driver) Set(atHeight int64, key, value []byte) error {
	// should never reach here, all should be batched in tiered+hld
	panic("should never reach here")
}

func (d *Driver) SetSync(atHeight int64, key, value []byte) error {
	// should never reach here, all should be batched in tiered+hld
	panic("should never reach here")
}

func (d *Driver) Delete(atHeight int64, key []byte) error {
	// should never reach here, all should be batched in tiered+hld
	panic("should never reach here")
}

func (d *Driver) DeleteSync(atHeight int64, key []byte) error {
	return d.Delete(atHeight, key)
}

func (d *Driver) Iterator(maxHeight int64, start, end []byte) (hld.HeightLimitEnabledIterator, error) {
	return NewLevelDBIterator(d, maxHeight, start, end)
}

func (d *Driver) ReverseIterator(maxHeight int64, start, end []byte) (hld.HeightLimitEnabledIterator, error) {
	return NewLevelDBReverseIterator(d, maxHeight, start, end)
}

func (d *Driver) Close() error {
	d.session.Close()
	return nil
}

func (d *Driver) NewBatch(atHeight int64) hld.HeightLimitEnabledBatch {
	return NewLevelDBBatch(atHeight, d)
}

// TODO: Implement me
func (d *Driver) Print() error {
	return nil
}

func (d *Driver) Stats() map[string]string {
	return nil
}
