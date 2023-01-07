package rollbackable

import (
	tmdb "github.com/tendermint/tm-db"
)

type HasRollbackBatch interface {
	RollbackBatch() tmdb.Batch
}

var _ tmdb.Batch = (*Batch)(nil)

type Batch struct {
	tmdb.Batch

	db            tmdb.DB
	RollbackBatch tmdb.Batch
	RecordCount   int
}

func NewRollbackableBatch(db tmdb.DB) *Batch {
	return &Batch{
		db:            db,
		Batch:         db.NewBatch(),
		RollbackBatch: db.NewBatch(),
	}
}

// revert value for key to previous state.
func (b *Batch) backup(key []byte) error {
	b.RecordCount++

	data, err := b.db.Get(key)
	if err != nil {
		return err
	}

	if data == nil {
		return b.RollbackBatch.Delete(key)
	}
	return b.RollbackBatch.Set(key, data)
}

func (b *Batch) Set(key, value []byte) error {
	if err := b.backup(key); err != nil {
		return err
	}

	return b.Batch.Set(key, value)
}

func (b *Batch) Delete(key []byte) error {
	if err := b.backup(key); err != nil {
		return err
	}

	return b.Batch.Delete(key)
}
