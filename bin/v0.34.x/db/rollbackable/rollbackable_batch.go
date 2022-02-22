package rollbackable

import (
	tmdb "github.com/tendermint/tm-db"
)

type HasRollbackBatch interface {
	RollbackBatch() tmdb.Batch
}

var _ tmdb.Batch = (*RollbackableBatch)(nil)

type RollbackableBatch struct {
	tmdb.Batch

	db            tmdb.DB
	RollbackBatch tmdb.Batch
}

func NewRollbackableBatch(db tmdb.DB) *RollbackableBatch {
	return &RollbackableBatch{
		db:            db,
		Batch:         db.NewBatch(),
		RollbackBatch: db.NewBatch(),
	}
}

// revert value for key to previous state
func (b *RollbackableBatch) backup(key []byte) error {
	data, err := b.db.Get(key)
	if err != nil {
		return err
	}
	if data == nil {
		return b.RollbackBatch.Delete(key)
	} else {
		return b.RollbackBatch.Set(key, data)
	}
}

func (b *RollbackableBatch) Set(key, value []byte) error {
	if err := b.backup(key); err != nil {
		return err
	}
	return b.Batch.Set(key, value)
}

func (b *RollbackableBatch) Delete(key []byte) error {
	if err := b.backup(key); err != nil {
		return err
	}
	return b.Batch.Delete(key)
}
