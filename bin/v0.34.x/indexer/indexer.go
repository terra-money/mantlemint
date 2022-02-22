package indexer

import (
	"fmt"
	"github.com/gorilla/mux"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/snappy"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint/event"
	"time"
)

type Indexer struct {
	db          tmdb.DB
	indexerTags []string
	indexers    []IndexFunc
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

func (idx *Indexer) Run(block *tm.Block, blockId *tm.BlockID, evc *event.EventCollector) error {
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

func (idx *Indexer) RegisterRESTRoute(router *mux.Router, registerer RESTRouteRegisterer) {
	registerer(router, idx.db)
}
