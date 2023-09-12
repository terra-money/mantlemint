package block_feed

import (
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
)

func extractBlockFromWSResponse(message []byte) (*BlockResult, error) {
	data := new(struct {
		Result struct {
			Data struct {
				Value *BlockResult `json:"value"`
			} `json:"data"`
		} `json:"result"`
	})

	if unmarshalErr := tmjson.Unmarshal(message, data); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return data.Result.Data.Value, nil
}

func ExtractBlockFromRPCResponse(message []byte) (*BlockResult, error) {
	data := new(struct {
		Result *BlockResult `json:"result"`
	})

	if err := tmjson.Unmarshal(message, data); err != nil {
		return nil, err
	}

	return data.Result, nil
}

func ExtractBlockResultFromRPCResponse(message []byte) ([]abci.ResponseDeliverTx, error) {
	data := new(struct {
		Result struct {
			TxsResult []abci.ResponseDeliverTx `json:"txs_results"`
		} `json:"result"`
	})

	if err := tmjson.Unmarshal(message, data); err != nil {
		return nil, err
	}

	return data.Result.TxsResult, nil
}
