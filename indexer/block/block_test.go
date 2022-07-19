package block

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/safe_batch"
)

func TestIndexBlock(t *testing.T) {
	db := tmdb.NewMemDB()
	blockFile, _ := os.Open("../fixtures/block_4724005_raw.json")
	blockJSON, _ := ioutil.ReadAll(blockFile)

	record := BlockRecord{}
	_ = tmjson.Unmarshal(blockJSON, &record)

	batch := safe_batch.NewSafeBatchDB(db)
	batch.(safe_batch.SafeBatchDBCloser).Open()
	if err := IndexBlock(*batch.(*safe_batch.SafeBatchDB), record.Block, record.BlockID, nil, nil); err != nil {
		panic(err)
	}
	batch.(safe_batch.SafeBatchDBCloser).Flush()

	block, err := blockByHeightHandler(db, "4724005")
	assert.Nil(t, err)
	assert.NotNil(t, block)

	fmt.Println(string(block))
}
