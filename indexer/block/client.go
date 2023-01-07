package block

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/indexer"
)

var EndpointGETBlocksHeight = "/index/blocks/{height}"

var (
	ErrorInvalidHeight = func(height string) string { return fmt.Sprintf("invalid height %s", height) }
	ErrorBlockNotFound = func(height string) string { return fmt.Sprintf("block %s not found... yet.", height) }
)

func blockByHeightHandler(indexerDB tmdb.DB, height string) (json.RawMessage, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, errors.New(ErrorInvalidHeight(height))
	}
	return indexerDB.Get(getKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	router.HandleFunc(EndpointGETBlocksHeight, func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		height, ok := vars["height"]
		if !ok {
			http.Error(writer, ErrorInvalidHeight(height), http.StatusBadRequest)
			return
		}

		if block, err := blockByHeightHandler(indexerDB, height); err != nil {
			http.Error(writer, indexer.ErrorInternal(err), http.StatusInternalServerError)
			return
		} else if block == nil {
			// block not seen;
			http.Error(writer, ErrorBlockNotFound(height), http.StatusBadRequest)
			return
		} else {
			writer.WriteHeader(http.StatusOK)
			writer.Write(block)
			return
		}
	}).Methods("GET")
})
