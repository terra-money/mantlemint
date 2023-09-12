package indexer

import (
	"fmt"
	"time"

	tmdb "github.com/cometbft/cometbft-db"
	tm "github.com/cometbft/cometbft/types"
	"github.com/gorilla/mux"
	terra "github.com/terra-money/core/v2/app"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/db/snappy"
	"github.com/terra-money/mantlemint/mantlemint"
)

type Indexer struct {
	db          tmdb.DB
	indexerTags []string
	indexers    []IndexFunc
	app         *terra.TerraApp
}

func NewIndexer(dbName, path string, app *terra.TerraApp) (*Indexer, error) {
	indexerDB, indexerDBError := tmdb.NewGoLevelDB(dbName, path)
	if indexerDBError != nil {
		return nil, indexerDBError
	}

	indexerDBCompressed := snappy.NewSnappyDB(indexerDB, snappy.CompatModeEnabled)

	return &Indexer{
		db:          indexerDBCompressed,
		indexerTags: []string{},
		indexers:    []IndexFunc{},
		app:         app,
	}, nil
}

func (idx *Indexer) RegisterIndexerService(tag string, indexerFunc IndexFunc) {
	idx.indexerTags = append(idx.indexerTags, tag)
	idx.indexers = append(idx.indexers, indexerFunc)
}

func (idx *Indexer) Run(block *tm.Block, blockId *tm.BlockID, evc *mantlemint.EventCollector) error {
	//batch := idx.db.NewBatch()
	batch := safe_batch.NewSafeBatchDB(idx.db)
	batchedOrigin := batch.(safe_batch.SafeBatchDBCloser)
	batchedOrigin.Open()

	tStart := time.Now()
	for _, indexerFunc := range idx.indexers {
		if indexerErr := indexerFunc(*batch.(*safe_batch.SafeBatchDB), block, blockId, evc, idx.app); indexerErr != nil {
			return indexerErr
		}
	}
	tEnd := time.Now()
	fmt.Printf("[indexer] finished %d indexers, %dms\n", len(idx.indexers), tEnd.Sub(tStart).Milliseconds())

	if _, err := batchedOrigin.Flush(); err != nil {
		return err
	}

	return nil
}

func (idx *Indexer) RegisterRESTRoute(router *mux.Router, registerer RESTRouteRegisterer) {
	registerer(router, idx.db)
}
