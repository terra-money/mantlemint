package tx

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tendermint "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/safebatch"
	"github.com/terra-money/mantlemint/mantlemint"
)

//nolint:forbidigo
func TestIndexTx(t *testing.T) {
	db := tmdb.NewMemDB()
	block := &tendermint.Block{}
	blockFile, _ := os.Open("../fixtures/block_4814775.json")
	blockJSON, _ := io.ReadAll(blockFile)
	if err := tmjson.Unmarshal(blockJSON, block); err != nil {
		t.Fail()
	}

	eventFile, _ := os.Open("../fixtures/response_4814775.json")
	eventJSON, _ := io.ReadAll(eventFile)
	evc := mantlemint.NewMantlemintEventCollector()
	event := tendermint.EventDataTx{}
	if err := tmjson.Unmarshal(eventJSON, &event.Result); err != nil {
		panic(err)
	}

	_ = evc.PublishEventTx(event)

	batch := safebatch.NewSafeBatchDB(db)

	if err := IndexTx(*batch.(*safebatch.SafeBatchDB), block, nil, evc, nil, nil); err != nil {
		panic(err)
	}
	batch.(safebatch.SafeBatchDBCloser).Flush()

	txn, err := txByHashHandler(db, "C794D5CE7179AED455C10E8E7645FE8F8A40BA0C97F1275AB87B5E88A52CB2C3")
	assert.Nil(t, err)
	assert.NotNil(t, txn)
	fmt.Println(string(txn))

	txns, err := txsByHeightHandler(db, "4814775")
	assert.Nil(t, err)
	assert.NotNil(t, txns)
	fmt.Println(string(txns))
}
