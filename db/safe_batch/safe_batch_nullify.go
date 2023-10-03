package safe_batch

import dbm "github.com/tendermint/tm-db"

var _ dbm.Batch = (*SafeBatchNullified)(nil)

type SafeBatchNullified struct {
	batch dbm.Batch
}

func NewSafeBatchNullify(batch dbm.Batch) dbm.Batch {
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
