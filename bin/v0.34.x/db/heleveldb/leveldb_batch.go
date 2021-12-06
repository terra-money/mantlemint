package heleveldb

import (
	"math"

	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"github.com/terra-money/mantlemint-provider-v0.34.x/lib"
)

var _ hld.HeightLimitEnabledBatch = (*LevelBatch)(nil)

type LevelBatch struct {
	height int64
	batch  tmdb.Batch
	mode   int
}

func (b *LevelBatch) keyBytesWithHeight(key []byte) []byte {
	if b.mode == DriverModeKeySuffixAsc {
		return append(prefixDataWithHeightKey(key), lib.UintToBigEndian(uint64(b.height))...)
	} else {
		return append(prefixDataWithHeightKey(key), lib.UintToBigEndian(math.MaxUint64-uint64(b.height))...)
	}

}

func NewLevelDBBatch(atHeight int64, driver *Driver) *LevelBatch {
	return &LevelBatch{
		height: atHeight,
		batch:  driver.session.NewBatch(),
		mode:   driver.mode,
	}
}

func (b *LevelBatch) Set(key, value []byte) error {
	newKey := b.keyBytesWithHeight(key)

	// make fixed size byte slice for performance
	buf := make([]byte, 0, len(value)+1)
	buf = append(buf, byte(0)) // 0 => not deleted
	buf = append(buf, value...)

	if err := b.batch.Set(prefixCurrentDataKey(key), buf[1:]); err != nil {
		return err
	}
	if err := b.batch.Set(prefixKeysForIteratorKey(key), []byte{}); err != nil {
		return err
	}
	return b.batch.Set(newKey, buf)
}

func (b *LevelBatch) Delete(key []byte) error {
	newKey := b.keyBytesWithHeight(key)

	buf := []byte{1}

	if err := b.batch.Delete(prefixCurrentDataKey(key)); err != nil {
		return err
	}
	if err := b.batch.Set(prefixKeysForIteratorKey(key), buf); err != nil {
		return err
	}
	return b.batch.Set(newKey, buf)
}

func (b *LevelBatch) Write() error {
	return b.batch.Write()
}

func (b *LevelBatch) WriteSync() error {
	return b.batch.WriteSync()
}

func (b *LevelBatch) Close() error {
	return b.batch.Close()
}
