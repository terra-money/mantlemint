package tx

import (
	"encoding/json"
	"fmt"
	tendermint "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/block_feed"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint"
	"io/ioutil"
	"net/http"
)

func LazySync(height int64, rpcEndpoint string, indexerDB tmdb.DB) (json.RawMessage, error) {
	fmt.Printf("[indexer/tx/lazysync] syncing block %d..\n", height)

	// get block
	resp, err := http.Get(fmt.Sprintf("%v/block?height=%d", rpcEndpoint, height))
	if err != nil {
		return nil, err
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	block, err := block_feed.ExtractBlockFromRPCResponse(response)
	if err != nil {
		return nil, err
	}


	// get results
	results, err := http.Get(fmt.Sprintf("%v/block_results?height=%d", rpcEndpoint, height))
	if err != nil {
		return nil, err
	}

	response, err = ioutil.ReadAll(results.Body)

	blockResults, err := block_feed.ExtractBlockResultFromRPCResponse(response)
	if err != nil {
		return nil, err
	}

	evc := mantlemint.NewMantlemintEventCollector()
	for _, result := range blockResults {
		_ = evc.PublishEventTx(tendermint.EventDataTx{
			TxResult: result,
		})
	}

	batch := indexerDB.NewBatch()
	if indexerErr := IndexTx(batch, block.Block, block.BlockID, evc); indexerErr != nil {
		return nil, indexerErr
	}

	if err := batch.WriteSync(); err != nil {
		return nil, err
	}

	if err := batch.Close(); err != nil {
		return nil, err
	}

	return nil, nil
}
