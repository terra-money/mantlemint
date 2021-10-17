package block_feed

import (
	tmjson "github.com/tendermint/tendermint/libs/json"
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
