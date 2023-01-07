package tx

import (
	"encoding/json"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/terra-money/mantlemint/lib"
)

//nolint:revive
type TxRecord struct {
	Tx         json.RawMessage `json:"tx"`
	TxResponse json.RawMessage `json:"tx_response"`
}

//nolint:revive
type TxByHeightRecord struct {
	Code      uint32          `json:"code"`
	Codespace string          `json:"codespace"`
	GasUsed   int64           `json:"gas_used"`
	GasWanted int64           `json:"gas_wanted"`
	Height    int64           `json:"height"`
	RawLog    string          `json:"raw_log"`
	Logs      json.RawMessage `json:"logs"`
	TxHash    string          `json:"txhash"`
	Timestamp time.Time       `json:"timestamp"`
	Tx        json.RawMessage `json:"tx"`
}

var (
	txPrefix = []byte("tx/hash:")
	getKey   = func(hash string) []byte {
		return lib.ConcatBytes(txPrefix, []byte(hash))
	}
)

var (
	byHeightPrefix = []byte("tx/height:")
	getByHeightKey = func(height uint64) []byte {
		return lib.ConcatBytes(byHeightPrefix, lib.UintToBigEndian(height))
	}
)

type ResponseDeliverTx struct {
	Code      uint32  `json:"code"`
	Data      []byte  `json:"data,omitempty"`
	Log       string  `json:"log,omitempty"`
	Info      string  `json:"info,omitempty"`
	GasWanted int64   `json:"gas_wanted,omitempty"`
	GasUsed   int64   `json:"gas_used,omitempty"`
	Events    []Event `json:"events,omitempty"`
	Codespace string  `json:"codespace,omitempty"`
}

type Event struct {
	Type       string           `json:"type,omitempty"`
	Attributes []EventAttribute `json:"attributes,omitempty"`
}

type EventAttribute struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func ToResponseDeliverTxJSON(responseDeliverTx *abci.ResponseDeliverTx) *ResponseDeliverTx {
	result := &ResponseDeliverTx{}
	result.Code = responseDeliverTx.Code
	result.Data = responseDeliverTx.Data
	result.Log = responseDeliverTx.Log
	result.Info = responseDeliverTx.Info
	result.GasWanted = responseDeliverTx.GasWanted
	result.GasUsed = responseDeliverTx.GasUsed
	result.Codespace = responseDeliverTx.Codespace
	result.Events = []Event{}

	for _, event := range responseDeliverTx.Events {
		nEvent := Event{}
		nEvent.Type = event.Type
		for _, attribute := range event.Attributes {
			nAttribute := EventAttribute{
				Key:   string(attribute.Key),
				Value: string(attribute.Value),
			}

			nEvent.Attributes = append(nEvent.Attributes, nAttribute)
		}

		result.Events = append(result.Events, nEvent)
	}

	return result
}
