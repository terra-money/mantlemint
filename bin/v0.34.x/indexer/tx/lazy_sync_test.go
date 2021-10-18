package tx

import (
	"fmt"
	tmdb "github.com/tendermint/tm-db"
	"testing"
)

func TestLazySync(t *testing.T) {
	height := 4944020
	rpcEndpoint := "http://public-node.terra.dev:26657"
	indexerDB := tmdb.NewMemDB()

	json, err := LazySync(int64(height), rpcEndpoint, indexerDB)
	fmt.Println(json, err)

	_, err = indexerDB.Get(getByHeightKey(uint64(height)))
	//fmt.Println(string(a))

}
