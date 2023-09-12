package block

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	tmjson "github.com/tendermint/tendermint/libs/json"
	dbm "github.com/tendermint/tm-db"
)

func TestIndexBlock(t *testing.T) {
	db := dbm.NewMemDB()
	blockFile, _ := os.Open("../fixtures/block_4724005_raw.json")
	blockJSON, _ := io.ReadAll(blockFile)

	record := BlockRecord{}
	_ = tmjson.Unmarshal(blockJSON, &record)

	batch := db.NewBatch()
	if err := IndexBlock(batch, record.Block, record.BlockID, nil); err != nil {
		panic(err)
	}
	_ = batch.WriteSync()
	_ = batch.Close()

	block, err := blockByHeightHandler(db, "4724005")
	assert.Nil(t, err)
	assert.NotNil(t, block)

	fmt.Println(string(block))
}
