package tx

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/indexer"
)

var (
	ErrorInvalidHeight = func(height string) string { return fmt.Sprintf("invalid height %s", height) }
	ErrorTxsNotFound   = func(height string) string { return fmt.Sprintf("txs at height %s not found... yet.", height) }
	ErrorInvalidHash   = func(hash string) string { return fmt.Sprintf("invalid hash %s", hash) }
	ErrorTxNotFound    = func(hash string) string { return fmt.Sprintf("tx (%s) not found... yet or forever.", hash) }
)

func txByHashHandler(indexerDB tmdb.DB, txHash string) ([]byte, error) {
	return indexerDB.Get(getKey(txHash))
}

func txsByHeightHandler(indexerDB tmdb.DB, height string) ([]byte, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, errors.New(ErrorInvalidHeight(height))
	}
	return indexerDB.Get(getByHeightKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	router.HandleFunc("/index/tx/by_hash/{hash}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		hash, ok := vars["hash"]
		if !ok {
			http.Error(writer, ErrorInvalidHash(hash), http.StatusBadRequest)
			return
		}

		if txn, err := txByHashHandler(indexerDB, hash); err != nil {
			http.Error(writer, indexer.ErrorInternal(err), http.StatusInternalServerError)
			return
		} else if txn == nil {
			http.Error(writer, ErrorTxNotFound(hash), http.StatusBadRequest)
			return
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Write(txn)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/index/tx/by_height/{height}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		height, ok := vars["height"]
		if !ok {
			http.Error(writer, ErrorInvalidHeight(height), http.StatusBadRequest)
			return
		}

		if txns, err := txsByHeightHandler(indexerDB, height); err != nil {
			http.Error(writer, indexer.ErrorInternal(err), http.StatusBadRequest)
			return
		} else if txns == nil {
			http.Error(writer, ErrorTxsNotFound(height), http.StatusBadRequest)
			return
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Write(txns)
			return
		}
	}).Methods("GET")
})
