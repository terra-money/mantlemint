package safe_batch

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/rollbackable"
)

var (
	_ dbm.DB            = (*SafeBatchDB)(nil)
	_ SafeBatchDBCloser = (*SafeBatchDB)(nil)
)

type SafeBatchDBCloser interface {
	dbm.DB
	Open()
	Flush() (dbm.Batch, error)
}

type SafeBatchDB struct {
	db    dbm.DB
	batch dbm.Batch
}

// open batch
func (s *SafeBatchDB) Open() {
	s.batch = s.db.NewBatch()
}

// flush batch and return rollback batch if rollbackable
func (s *SafeBatchDB) Flush() (dbm.Batch, error) {
	defer func() {
		if s.batch != nil {
			s.batch.Close()
		}
		s.batch = nil
	}()

	if batch, ok := s.batch.(rollbackable.HasRollbackBatch); ok {
		return batch.RollbackBatch(), s.batch.WriteSync()
	} else {
		return nil, s.batch.WriteSync()
	}
}

func NewSafeBatchDB(db dbm.DB) dbm.DB {
	return &SafeBatchDB{
		db:    db,
		batch: nil,
	}
}

func (s *SafeBatchDB) Get(bytes []byte) ([]byte, error) {
	return s.db.Get(bytes)
}

func (s *SafeBatchDB) Has(key []byte) (bool, error) {
	return s.db.Has(key)
}

func (s *SafeBatchDB) Set(key, value []byte) error {
	if s.batch != nil {
		return s.batch.Set(key, value)
	} else {
		return s.db.Set(key, value)
	}
}

func (s *SafeBatchDB) SetSync(key, value []byte) error {
	return s.Set(key, value)
}

func (s *SafeBatchDB) Delete(key []byte) error {
	if s.batch != nil {
		return s.batch.Delete(key)
	} else {
		return s.db.Delete(key)
	}
}

func (s *SafeBatchDB) DeleteSync(key []byte) error {
	return s.Delete(key)
}

func (s *SafeBatchDB) Iterator(start, end []byte) (dbm.Iterator, error) {
	return s.db.Iterator(start, end)
}

func (s *SafeBatchDB) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	return s.db.ReverseIterator(start, end)
}

func (s *SafeBatchDB) Close() error {
	return s.db.Close()
}

func (s *SafeBatchDB) NewBatch() dbm.Batch {
	if s.batch != nil {
		return NewSafeBatchNullify(s.batch)
	} else {
		fmt.Println("=== warn! should never enter here")
		return s.db.NewBatch()
	}
}

func (s *SafeBatchDB) Print() error {
	return s.db.Print()
}

func (s *SafeBatchDB) Stats() map[string]string {
	return s.db.Stats()
}
