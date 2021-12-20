package block

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/config"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer"
	"net/http"
	"strconv"
)

var (
	EndpointGETBlocksHeight = "/index/blocks/{height}"
	EndpointPOSTBlock       = "/index/block"
)

func blockByHeightHandler(indexerDB tmdb.DB, height string) (json.RawMessage, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, fmt.Errorf("invalid height: %v", err)
	}
	return indexerDB.Get(getKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	var mantlemintConfig = config.NewConfig()
	router.HandleFunc(EndpointGETBlocksHeight, func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		height, ok := vars["height"]
		if !ok {
			writer.WriteHeader(400)
			writer.Write([]byte("invalid height"))
			return
		}

		if block, err := blockByHeightHandler(indexerDB, height); err != nil {
			writer.WriteHeader(400)
			writer.Write([]byte(err.Error()))
			return
		} else if block == nil {
			// block not seen;
			heightInInt, err := strconv.Atoi(height)
			if err != nil {
				http.Error(writer, fmt.Errorf("invalid height: %v", err).Error(), 400)
				return
			}
			if record, lazySyncErr := LazySync(int64(heightInInt), mantlemintConfig.RPCEndpoints[0], indexerDB); lazySyncErr != nil {
				http.Error(writer, lazySyncErr.Error(), 400)
				return
			} else {
				writer.WriteHeader(200)
				writer.Write(record)
			}
		} else {
			writer.WriteHeader(200)
			writer.Write(block)
			return
		}
	}).Methods("GET")
})
