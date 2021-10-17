package block

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	EndpointGETBlocksHeight = "/index/blocks/{height}"
	EndpointPOSTBlock      = "/index/block"
)

func blockByHeightHandler(indexerDB tmdb.DB, height string) (json.RawMessage, error) {
	heightInInt, err := strconv.Atoi(height)
	if err != nil {
		return nil, fmt.Errorf("invalid height: %v", err)
	}
	return indexerDB.Get(getKey(uint64(heightInInt)))
}

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, postRouter *mux.Router, indexerDB tmdb.DB) {
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
		} else {
			writer.WriteHeader(200)
			writer.Write(block)
			return
		}
	}).Methods("GET")

	postRouter.HandleFunc(EndpointPOSTBlock, func(writer http.ResponseWriter, request *http.Request) {
		bz, err := ioutil.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "error reading body", 400)
			return
		}
		record := BlockRecord{}
		if err := tmjson.Unmarshal(bz, &record); err != nil {
			http.Error(writer, "error unmarshaling block record", 400)
			return
		}

		// prevent rewrite
		injectedHeight := record.Block.Height
		data, err := indexerDB.Get(getKey(uint64(injectedHeight)))
		if err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		if data != nil {
			writer.WriteHeader(204)
			writer.Write([]byte("already committed"))
			return
		}

		batch := indexerDB.NewBatch()
		if err := IndexBlock(batch, record.Block, record.BlockID, nil); err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		if err := batch.WriteSync(); err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}

		if err := batch.Close(); err != nil {
			http.Error(writer, err.Error(), 400)
			return
		}
	}).Methods("POST")
})
