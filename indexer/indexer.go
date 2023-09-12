package indexer

import (
	"fmt"
	"time"

	"github.com/gorilla/mux"
	tm "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/snappy"
	"github.com/terra-money/mantlemint/mantlemint"
)

type Indexer struct {
	db          dbm.DB
	indexerTags []string
	indexers    []IndexFunc
}

func NewIndexer(dbName, path string) (*Indexer, error) {
	indexerDB, indexerDBError := dbm.NewGoLevelDB(dbName, path)
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

func (idx *Indexer) RegisterRESTRoute(router *mux.Router, registerer RESTRouteRegisterer) {
	registerer(router, idx.db)
}
