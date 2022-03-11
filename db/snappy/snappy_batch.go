package snappy

import (
	"github.com/golang/snappy"
	tmdb "github.com/tendermint/tm-db"
)

var _ tmdb.Batch = (*SnappyBatch)(nil)

type SnappyBatch struct {
	batch tmdb.Batch
}

func NewSnappyBatch(batch tmdb.Batch) *SnappyBatch {
	return &SnappyBatch{
		batch: batch,
	}
}

func (s *SnappyBatch) Set(key, value []byte) error {
	return s.batch.Set(key, snappy.Encode(nil, value))
}

func (s *SnappyBatch) Delete(key []byte) error {
	return s.batch.Delete(key)
}

func (s *SnappyBatch) Write() error {
	return s.batch.Write()
}

func (s *SnappyBatch) WriteSync() error {
	return s.batch.WriteSync()
}

func (s *SnappyBatch) Close() error {
	return s.batch.Close()
}
