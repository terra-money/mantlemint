package height

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tm "github.com/tendermint/tendermint/types"
	"github.com/terra-money/mantlemint/db/safebatch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
)

var IndexHeight = indexer.CreateIndexer(func(
	indexerDB safebatch.SafeBatchDB,
	block *tm.Block,
	blockID *tm.BlockID,
	evc *mantlemint.EventCollector,
	app indexer.ABCIApp,
	txConfig client.TxConfig,
) error {
	//nolint:forbidigo
	defer fmt.Printf("[indexer/height] indexing done for height %d\n", block.Height)
	height := block.Height

	record := HeightRecord{Height: uint64(height)}
	recordJSON, recordErr := tmjson.Marshal(record)
	if recordErr != nil {
		return recordErr
	}

	return indexerDB.Set(getKey(), recordJSON)
})
