package block

import (
	"fmt"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tm "github.com/tendermint/tendermint/types"
	"github.com/terra-money/mantlemint/db/safebatch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
	"github.com/cosmos/cosmos-sdk/client"
)

var IndexBlock = indexer.CreateIndexer(func(
	indexerDB safebatch.SafeBatchDB,
	block *tm.Block,
	blockID *tm.BlockID,
	evc *mantlemint.EventCollector,
	app indexer.ABCIApp,
	txConfig client.TxConfig,
) error {
	//nolint:forbidigo
	defer fmt.Printf("[indexer/block] indexing done for height %d\n", block.Height)
	record := BlockRecord{
		Block:   block,
		BlockID: blockID,
	}

	recordJSON, recordErr := tmjson.Marshal(record)
	if recordErr != nil {
		return recordErr
	}

	return indexerDB.Set(getKey(uint64(block.Height)), recordJSON)
})
