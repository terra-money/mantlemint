package indexer

import (
	"fmt"
	"github.com/gorilla/mux"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/snappy"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint"
	"net/http"
	"time"
)

type Indexer struct {
	db             tmdb.DB
	indexerTags    []string
	indexers       []IndexFunc
	sidesyncRouter *mux.Router
}

func NewIndexer(dbName, path string) (*Indexer, error) {
	indexerDB, indexerDBError := tmdb.NewGoLevelDB(dbName, path)
	if indexerDBError != nil {
		return nil, indexerDBError
	}

	indexerDBCompressed := snappy.NewSnappyDB(indexerDB, snappy.CompatModeEnabled)

	return &Indexer{
		db:          indexerDBCompressed,
		indexerTags: []string{},
		indexers:    []IndexFunc{},
	}, nil
}

func (idx *Indexer) RegisterIndexerService(tag string, indexerFunc IndexFunc) {
	idx.indexerTags = append(idx.indexerTags, tag)
	idx.indexers = append(idx.indexers, indexerFunc)
}

func (idx *Indexer) Run(block *tm.Block, blockId *tm.BlockID, evc *mantlemint.EventCollector) error {
	batch := idx.db.NewBatch()
	tStart := time.Now()
	for _, indexerFunc := range idx.indexers {
		if indexerErr := indexerFunc(batch, block, blockId, evc); indexerErr != nil {
			return indexerErr
		}
	}
	tEnd := time.Now()
	fmt.Printf("[indexer] finished %d indexers, %dms\n", len(idx.indexers), tEnd.Sub(tStart).Milliseconds())

	if err := batch.WriteSync(); err != nil {
		return err
	}

	if err := batch.Close(); err != nil {
		return err
	}

	return nil
}

func (idx *Indexer) WithSideSyncRouter(registerer func(sidesyncRouter *mux.Router)) *Indexer {
	idx.sidesyncRouter = mux.NewRouter()
	registerer(idx.sidesyncRouter)

	return idx
}

func (idx *Indexer) RegisterRESTRoute(router *mux.Router, postRouter *mux.Router, registerer RESTRouteRegisterer) {
	registerer(router, postRouter, idx.db)
}

func (idx *Indexer) StartSideSync(port int64) {
	if idx.sidesyncRouter == nil {
		panic(fmt.Errorf("sidesync router not set, perhaps you didn't call WithSyideSyncRouter first?\n"))
	}

	router := idx.sidesyncRouter
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		panic(err)
	}
}
