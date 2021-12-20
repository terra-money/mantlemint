package indexer

import (
	"github.com/gorilla/mux"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint"
	"net/http"
)

type IndexFunc func(indexerDB tmdb.Batch, block *tm.Block, blockId *tm.BlockID, evc *mantlemint.EventCollector) error
type ClientHandler func(w http.ResponseWriter, r *http.Request) error
type RESTRouteRegisterer func(router *mux.Router, indexerDB tmdb.DB)

func CreateIndexer(idf IndexFunc) IndexFunc {
	return idf
}

func CreateHandler(ich ClientHandler) ClientHandler {
	return ich
}

func CreateRESTRoute(registerer RESTRouteRegisterer) RESTRouteRegisterer {
	return registerer
}
