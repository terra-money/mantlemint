package block

import (
	"fmt"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tm "github.com/tendermint/tendermint/types"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
)

var IndexBlock = indexer.CreateIndexer(func(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, _ *mantlemint.EventCollector, _ *cosmoscmd.App) error {
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

func IterateBlocks(indexerDb safe_batch.SafeBatchDB, start int64, end int64, cb func(block *tm.Block) (stop bool)) (err error) {
	iter, err := indexerDb.Iterator(getKey(uint64(start)), getKey(uint64(end)))
	if err != nil {
		return err
	}
	for iter.Valid() {
		b := iter.Value()
		var block BlockRecord
		err := tmjson.Unmarshal(b, &block)
		if err != nil {
			return err
		}
		stop := cb(block.Block)
		if stop {
			return nil
		}
		iter.Next()
	}
	return nil
}
