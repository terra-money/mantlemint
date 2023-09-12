package height

import (
	"fmt"

	tmjson "github.com/cometbft/cometbft/libs/json"
	tm "github.com/cometbft/cometbft/types"
	terra "github.com/terra-money/core/v2/app"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
)

var IndexHeight = indexer.CreateIndexer(func(indexerDB safe_batch.SafeBatchDB, block *tm.Block, _ *tm.BlockID, _ *mantlemint.EventCollector, _ *terra.TerraApp) error {
	defer fmt.Printf("[indexer/height] indexing done for height %d\n", block.Height)
	height := block.Height

	record := HeightRecord{Height: uint64(height)}
	recordJSON, recordErr := tmjson.Marshal(record)
	if recordErr != nil {
		return recordErr
	}

	return indexerDB.Set(getKey(), recordJSON)
})
