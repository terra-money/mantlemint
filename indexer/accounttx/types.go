package accounttx

import sdk "github.com/cosmos/cosmos-sdk/types"

type AccountTx struct {
	TxHash string `json:"txhash"`
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
