package snappy

import (
	"github.com/stretchr/testify/assert"
	db "github.com/tendermint/tm-db"
	"testing"
)

func TestSnappyDB(t *testing.T) {
	snappy := NewSnappyDB(db.NewMemDB())

	assert.Nil(t, snappy.Set([]byte("test"), []byte("testValue")))

	var v []byte
	var err error

	// nil buffer test
	v, err = snappy.Get([]byte("non-existing"))
	assert.Nil(t, v)
	assert.Nil(t, err)

	v, err = snappy.Get([]byte("test"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("testValue"), v)

	assert.Nil(t, snappy.Delete([]byte("test")))
	v, err = snappy.Get([]byte("test"))
	assert.Nil(t, v)
	assert.Nil(t, err)

	// iterator is not supported
	var it db.Iterator
	it, err = snappy.Iterator([]byte("start"), []byte("end"))
	assert.Nil(t, it)
	assert.Equal(t, errIteratorNotSupported, err)

	it, err = snappy.ReverseIterator([]byte("start"), []byte("end"))
	assert.Nil(t, it)
	assert.Equal(t, errIteratorNotSupported, err)

	// batched store is compressed as well
	var batch db.Batch
	batch = snappy.NewBatch()

	assert.Nil(t, batch.Set([]byte("key"), []byte("batchedValue")))
	assert.Nil(t, batch.Write())
	assert.Nil(t, batch.Close())

	v, err = snappy.Get([]byte("key"))
	assert.Equal(t, []byte("batchedValue"), v)

	batch = snappy.NewBatch()
	assert.Nil(t, batch.Delete([]byte("key")))
	assert.Nil(t, batch.Write())
	assert.Nil(t, batch.Close())

	v, err = snappy.Get([]byte("key"))
	assert.Nil(t, v)
	assert.Nil(t, err)
}
