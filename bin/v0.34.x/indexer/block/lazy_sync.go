package block

import (
	"encoding/json"
	"fmt"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/block_feed"
	"io/ioutil"
	"net/http"
)

func LazySync(height int64, rpcEndpoint string, indexerDB tmdb.DB) (json.RawMessage, error) {
	fmt.Printf("[indexer/block/lazysync] syncing block %d..\n", height)

	resp, err := http.Get(fmt.Sprintf("%v/block?height=%d", rpcEndpoint, height))
	if err != nil {
		panic(err)
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	block, err := block_feed.ExtractBlockFromRPCResponse(response)
	if err != nil {
		panic(err)
	}

	record := BlockRecord{
		Block: block.Block,
		BlockID: block.BlockID,
	}

	batch := indexerDB.NewBatch()
	if err := IndexBlock(batch, record.Block, record.BlockID, nil); err != nil {
		return nil, err
	}

	if err := batch.WriteSync(); err != nil {
		return nil, err
	}

	if err := batch.Close(); err != nil {
		return nil, err
	}

	if resp, err := tmjson.Marshal(record); err != nil{
		return nil, err
	} else {
		return resp, nil
	}
}