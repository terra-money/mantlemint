package tx

import (
	"fmt"
	"github.com/gorilla/mux"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/config"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer"
	"net/http"
	"strconv"
)

func txByHashHandler(indexerDB tmdb.DB, txHash string) ([]byte, error) {
	return indexerDB.Get(getKey(txHash))
}

func txsByHeightHandler(indexerDB tmdb.DB, height string) ([]byte, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, fmt.Errorf("invalid height: %v", err)
	}
	return indexerDB.Get(getByHeightKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	var mantlemintConfig = config.NewConfig()
	router.HandleFunc("/index/tx/by_hash/{hash}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		hash, ok := vars["hash"]
		if !ok {
			http.Error(writer, "txn not found", 400)
			return
		}

		if txn, err := txByHashHandler(indexerDB, hash); err != nil {
			http.Error(writer, err.Error(), 400)
			return
		} else {
			writer.WriteHeader(200)
			writer.Write(txn)
			return
		}
	}).Methods("GET")

	router.HandleFunc("/index/tx/by_height/{height}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		height, ok := vars["height"]
		if !ok {
			http.Error(writer, "invalid height", 400)
			return
		}

		if txns, err := txsByHeightHandler(indexerDB, height); err != nil {
			http.Error(writer, err.Error(), 400)
			return
		} else if txns == nil {
			// block not seen;
			heightInInt, err := strconv.Atoi(height)
			if err != nil {
				http.Error(writer, fmt.Errorf("invalid height: %v", err).Error(), 400)
				return
			}
			if _, lazySyncErr := LazySync(int64(heightInInt), mantlemintConfig.RPCEndpoints[0], indexerDB); lazySyncErr != nil {
				http.Error(writer, lazySyncErr.Error(), 400)
				return
			} else {
				txns, err := txsByHeightHandler(indexerDB, height)
				if err != nil {
					http.Error(writer, err.Error(), 400)
					return
				} else {
					writer.WriteHeader(200)
					writer.Write(txns)
				}
			}
			return
		} else {
			writer.WriteHeader(200)
			writer.Write(txns)
			return
		}
	}).Methods("GET")
})
