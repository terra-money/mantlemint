package heleveldb

import (
	"bytes"
	"encoding/gob"

	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"github.com/terra-money/mantlemint/lib"
)

var _ hld.HeightLimitEnabledBatch = (*LevelBatch)(nil)

type LevelBatch struct {
	height int64
	batch  tmdb.Batch
}

func (b *LevelBatch) keyBytesWithHeight(key []byte) []byte {
	return append(prefixHeightSnapshotKey(key), lib.UintToBigEndian(uint64(b.height))...)
}

func NewLevelDBBatch(atHeight int64, driver *Driver) *LevelBatch {
	return &LevelBatch{
		height: atHeight,
		batch:  driver.session.NewBatch(),
	}
}

func (b *LevelBatch) Set(key, value []byte) error {
	newKey := b.keyBytesWithHeight(key)
	item := Item{
		Key:      key,
		Value:    value,
		AtHeight: b.height,
		Deleted:  false,
	}

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(item); err != nil {
		return err
	}

	if err := b.batch.Set(prefixOriginalDataKey(key), buff.Bytes()); err != nil {
		return err
	}
	return b.batch.Set(newKey, buff.Bytes())
}

func (b *LevelBatch) Delete(key []byte) error {
	newKey := b.keyBytesWithHeight(key)
	item := Item{
		Key:      key,
		Value:    []byte{},
		AtHeight: b.height,
		Deleted:  true,
	}

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(item); err != nil {
		return err
	}

	if err := b.batch.Set(prefixOriginalDataKey(key), buff.Bytes()); err != nil {
		return err
	}
	return b.batch.Set(newKey, buff.Bytes())
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
