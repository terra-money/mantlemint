package indexer

import (
	"log"
	"net/http"
	"runtime"

	tmdb "github.com/cometbft/cometbft-db"
	tm "github.com/cometbft/cometbft/types"
	"github.com/gorilla/mux"
	terra "github.com/terra-money/core/v2/app"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/mantlemint"
)

type IndexFunc func(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockId *tm.BlockID, evc *mantlemint.EventCollector, app *terra.TerraApp) error
type ClientHandler func(w http.ResponseWriter, r *http.Request) error
type RESTRouteRegisterer func(router *mux.Router, indexerDB tmdb.DB)

func CreateIndexer(idf IndexFunc) IndexFunc {
	return idf
}

func CreateRESTRoute(registerer RESTRouteRegisterer) RESTRouteRegisterer {
	return registerer
}

var (
	ErrorInternal = func(err error) string {
		_, fn, fl, ok := runtime.Caller(1)

		if !ok {
			// ...
		} else {
			log.Printf("ErrorInternal[%s:%d] %v\n", fn, fl, err.Error())
		}

		return "internal server error"
	}
)
