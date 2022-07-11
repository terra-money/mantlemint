package richlist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/indexer"
)

var (
	EndpointGETBlocksHeight = "/index/richlist/{height}"
)

var (
	ErrorInvalidHeight    = func(height string) string { return fmt.Sprintf("invalid height %s", height) }
	ErrorRichlistNotFound = func(height string) string { return fmt.Sprintf("richlist %s not found... yet.", height) }
)

func richlistByHeightHandler(indexerDB tmdb.DB, height string) (json.RawMessage, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, errors.New(ErrorInvalidHeight(height))
	}
	return indexerDB.Get(getDefaultKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	router.HandleFunc(EndpointGETBlocksHeight, func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		height, ok := vars["height"]
		if !ok {
			http.Error(writer, ErrorInvalidHeight(height), 400)
		}
		if height == "latest" {
			height = "0"
		}

		if list, err := richlistByHeightHandler(indexerDB, height); err != nil {
			http.Error(writer, indexer.ErrorInternal(err), 500)
			return
		} else if list == nil {
			// block not seen;
			http.Error(writer, ErrorRichlistNotFound(height), 400)
			return
		} else {
			writer.WriteHeader(200)
			writer.Write(list)
			return
		}
	}).Methods("GET")
})
