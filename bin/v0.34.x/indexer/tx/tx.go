package tx

import (
	"encoding/json"
	"fmt"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	terra "github.com/terra-money/core/app"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint"
)

var cdc = terra.MakeEncodingConfig()

var IndexTx = indexer.CreateIndexer(func(batch tmdb.Batch, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector) error {
	// encoder; proto -> mem -> json
	txDecoder := cdc.TxConfig.TxDecoder()
	jsonEncoder := cdc.TxConfig.TxJSONEncoder()

	txHashes := make([]string, len(block.Txs))
	txRecords := make([]TxRecord, len(block.Txs))

	// by hash
	for txIndex, txByte := range block.Txs {
		txRecord := TxRecord{}

		hash := txByte.Hash()
		tx, decodeErr := txDecoder(txByte)

		if decodeErr != nil {
			return decodeErr
		}

		// encode tx to JSON for max compat & shave deserialization cost at serving
		txJSON, _ := jsonEncoder(tx)

		// handle response -> json
		response := ToResponseDeliverTxJSON(evc.ResponseDeliverTxs[txIndex])
		responseJSON, responseMarshalErr := tmjson.Marshal(response)

		if responseMarshalErr != nil {
			return responseMarshalErr
		}

		// populate txRecord
		txRecord.Tx = txJSON
		txRecord.TxResponse = responseJSON

		txHashes[txIndex] = fmt.Sprintf("%X", hash)
		txRecords[txIndex] = txRecord
	}

	// 1. byHash -- matching the interface for /cosmos/tx/v1beta1/txs/{hash}
	for txIndex, txRecord := range txRecords {
		txRecordJSON, marshalErr := tmjson.Marshal(txRecord)
		if marshalErr != nil {
			return marshalErr
		}

		batchSetErr := batch.Set(getKey(txHashes[txIndex]), txRecordJSON)
		if batchSetErr != nil {
			return batchSetErr
		}
	}

	// 2. byHeight -- custom endpoint (use tx_response only)
	byHeightPayload := make([]json.RawMessage, len(txRecords))
	for txIndex, txRecord := range txRecords {
		byHeightPayload[txIndex] = txRecord.TxResponse
	}
	txByHeightJSON, marshalErr := tmjson.Marshal(byHeightPayload)
	if marshalErr != nil {
		return marshalErr
	}
	batchSetErr := batch.Set(getByHeightKey(uint64(block.Height)), txByHeightJSON)
	if batchSetErr != nil {
		return batchSetErr
	}

	return nil
})
