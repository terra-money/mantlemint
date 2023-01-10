package block

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/safebatch"
)

//nolint:forbidigo
func TestIndexBlock(t *testing.T) {
	db := tmdb.NewMemDB()
	blockFile, _ := os.Open("../fixtures/block_4724005_raw.json")
	blockJSON, _ := io.ReadAll(blockFile)

	record := BlockRecord{}
	_ = tmjson.Unmarshal(blockJSON, &record)

	batch := safebatch.NewSafeBatchDB(db)
	batch.(safebatch.SafeBatchDBCloser).Open()
	if err := IndexBlock(*batch.(*safebatch.SafeBatchDB), record.Block, record.BlockID, nil, nil, nil); err != nil {
		panic(err)
	}
	batch.(safebatch.SafeBatchDBCloser).Flush()

	block, err := blockByHeightHandler(db, "4724005")
	assert.Nil(t, err)
	assert.NotNil(t, block)

	fmt.Println(string(block))
}
