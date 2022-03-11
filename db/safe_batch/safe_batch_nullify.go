package safe_batch

import tmdb "github.com/tendermint/tm-db"

var _ tmdb.Batch = (*SafeBatchNullified)(nil)

type SafeBatchNullified struct {
	batch tmdb.Batch
}

func NewSafeBatchNullify(batch tmdb.Batch) tmdb.Batch {
	return &SafeBatchNullified{
		batch: batch,
	}
}

func (s SafeBatchNullified) Set(key, value []byte) error {
	return s.batch.Set(key, value)
}

func (s SafeBatchNullified) Delete(key []byte) error {
	return s.batch.Delete(key)
}

func (s SafeBatchNullified) Write() error {
	// noop
	return nil
}

func (s SafeBatchNullified) WriteSync() error {
	return s.Write()
}

func (s SafeBatchNullified) Close() error {
	// noop
	return nil
}
