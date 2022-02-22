package block

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer"
	"net/http"
	"strconv"
)

var (
	EndpointGETBlocksHeight = "/index/blocks/{height}"
	ErrorInvalidHeight      = func(height string) string { return fmt.Sprintf("invalid height %s", height) }
	ErrorBlockNotFound      = func(height string) string { return fmt.Sprintf("block %s not found... yet.", height) }
)

func blockByHeightHandler(indexerDB tmdb.DB, height string) (json.RawMessage, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, fmt.Errorf("invalid height: %v", err)
	}
	return indexerDB.Get(getKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	router.HandleFunc(EndpointGETBlocksHeight, func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		height, ok := vars["height"]
		if !ok {
			http.Error(writer, ErrorInvalidHeight(height), 400)
			return
		}

		if block, err := blockByHeightHandler(indexerDB, height); err != nil {
			http.Error(writer, ErrorBlockNotFound(height), 400)
			return
		} else {
			writer.WriteHeader(200)
			writer.Write(block)
			return
		}
	}).Methods("GET")
})
