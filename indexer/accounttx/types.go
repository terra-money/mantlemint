package accounttx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/mantlemint/indexer/tx"
	"time"
)

type AccountTx struct {
	TxHash      string    `json:"txhash"`
	BlockHeight uint64    `json:"height"`
	Timestamp   time.Time `json:"timestamp"`
}

var (
	AccountTxPrefix = []byte("account_tx/address:")
)

func GetAccountTxKeyByAddr(addr string) (key []byte) {
	key = append(AccountTxPrefix, addr...)
	return key
}

func GetAccountTxKey(addr string, blockHeight uint64, txIndex uint64) (key []byte) {
	key = append(GetAccountTxKeyByAddr(addr), sdk.Uint64ToBigEndian(blockHeight)...)
	key = append(key, sdk.Uint64ToBigEndian(txIndex)...)
	return key
}

type GetAccountTxsResponse struct {
	Limit  uint64                `json:"limit"`
	Offset uint64                `json:"offset"`
	Txs    []tx.TxByHeightRecord `json:"txs"`
}
