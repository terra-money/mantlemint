package heleveldb

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/hld"
	"github.com/terra-money/mantlemint/db/rollbackable"
)

var (
	_ hld.HeightLimitEnabledBatch   = (*LevelBatch)(nil)
	_ rollbackable.HasRollbackBatch = (*LevelBatch)(nil)
)

type LevelBatch struct {
	height int64
	batch  *rollbackable.RollbackableBatch
	mode   int
}

func (b *LevelBatch) keyBytesWithHeight(key []byte) []byte {
	return append(prefixDataWithHeightKey(key), serializeHeight(b.mode, b.height)...)
}

func NewLevelDBBatch(atHeight int64, driver *Driver) *LevelBatch {
	return &LevelBatch{
		height: atHeight,
		batch:  rollbackable.NewRollbackableBatch(driver.session),
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

func (b *LevelBatch) RollbackBatch() dbm.Batch {
	b.Metric()
	return b.batch.RollbackBatch
}

func (b *LevelBatch) Metric() {
	fmt.Printf("[rollback-batch] rollback batch for height %d's record length %d\n",
		b.height,
		b.batch.RecordCount,
	)
}
