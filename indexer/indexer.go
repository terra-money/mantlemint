package indexer

//nolint:staticcheck
import (
	"fmt"
	"time"

	"github.com/gorilla/mux"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/mantlemint/db/safebatch"
	"github.com/terra-money/mantlemint/db/snappy"
	"github.com/terra-money/mantlemint/mantlemint"
)

type Indexer struct {
	db          tmdb.DB
	indexerTags []string
	indexers    []IndexFunc
	app         *cosmoscmd.App
}

func NewIndexer(dbName, path string, app *cosmoscmd.App) (*Indexer, error) {
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

func (idx *Indexer) Run(block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector) error {
	// batch := idx.db.NewBatch()
	batch := safebatch.NewSafeBatchDB(idx.db)
	batchedOrigin := batch.(safebatch.SafeBatchDBCloser)
	batchedOrigin.Open()

	tStart := time.Now()
	for _, indexerFunc := range idx.indexers {
		if indexerErr := indexerFunc(*batch.(*safebatch.SafeBatchDB), block, blockID, evc, idx.app); indexerErr != nil {
			return indexerErr
		}
	}
	tEnd := time.Now()

	//nolint:forbidigo
	fmt.Printf("[indexer] finished %d indexers, %dms\n", len(idx.indexers), tEnd.Sub(tStart).Milliseconds())

	if _, err := batchedOrigin.Flush(); err != nil {
		return err
	}

	return nil
}

func (idx *Indexer) RegisterRESTRoute(router *mux.Router, registerer RESTRouteRegisterer) {
	registerer(router, idx.db)
}
