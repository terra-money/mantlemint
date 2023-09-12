package tx

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	tmjson "github.com/cometbft/cometbft/libs/json"
	cometbfttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/assert"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/mantlemint"
)

func TestIndexTx(t *testing.T) {
	db := tmdb.NewMemDB()
	block := &cometbfttypes.Block{}
	blockFile, _ := os.Open("../fixtures/block_4814775.json")
	blockJSON, _ := ioutil.ReadAll(blockFile)
	if err := tmjson.Unmarshal(blockJSON, block); err != nil {
		t.Fail()
	}

	eventFile, _ := os.Open("../fixtures/response_4814775.json")
	eventJSON, _ := ioutil.ReadAll(eventFile)
	evc := mantlemint.NewMantlemintEventCollector()
	event := cometbfttypes.EventDataTx{}
	if err := tmjson.Unmarshal(eventJSON, &event.Result); err != nil {
		panic(err)
	}

	_ = evc.PublishEventTx(event)

	safebatch := safe_batch.NewSafeBatchDB(db)
	if err := IndexTx(*safebatch.(*safe_batch.SafeBatchDB), block, nil, evc, nil); err != nil {
		panic(err)
	}
	safebatch.(safe_batch.SafeBatchDBCloser).Flush()

	txn, err := txByHashHandler(db, "C794D5CE7179AED455C10E8E7645FE8F8A40BA0C97F1275AB87B5E88A52CB2C3")
	assert.Nil(t, err)
	assert.NotNil(t, txn)
	fmt.Println(string(txn))

	txns, err := txsByHeightHandler(db, "4814775")
	assert.Nil(t, err)
	assert.NotNil(t, txns)
	fmt.Println(string(txns))
}
