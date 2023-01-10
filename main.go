package main

//nolint:staticcheck
import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/cosmos/cosmos-sdk/baseapp"
	abci "github.com/crescent-network/crescent/v4/app"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proxy"
	tendermint "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	blockFeeder "github.com/terra-money/mantlemint/blockfeed"
	"github.com/terra-money/mantlemint/config"
	"github.com/terra-money/mantlemint/db/heleveldb"
	"github.com/terra-money/mantlemint/db/hld"
	"github.com/terra-money/mantlemint/db/safebatch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/indexer/block"
	"github.com/terra-money/mantlemint/indexer/tx"
	"github.com/terra-money/mantlemint/mantlemint"
	"github.com/terra-money/mantlemint/rpc"
	"github.com/terra-money/mantlemint/store/rootmulti"
)

// initialize mantlemint for v0.34.x
//
//nolint:funlen,forbidigo
func main() {
	mantlemintConfig := config.GetConfig()
	mantlemintConfig.Print()

	//cmd.SetPrefixes(abci.AccountAddressPrefix)

	ldb, ldbErr := heleveldb.NewLevelDBDriver(&heleveldb.DriverConfig{
		Name: mantlemintConfig.MantlemintDB,
		Dir:  mantlemintConfig.Home,
		Mode: heleveldb.DriverModeKeySuffixDesc,
	})
	if ldbErr != nil {
		panic(ldbErr)
	}

	hldb := hld.ApplyHeightLimitedDB(
		ldb,
		&hld.HeightLimitedDBConfig{
			Debug: true,
		},
	)

	batched := safebatch.NewSafeBatchDB(hldb)
	batchedOrigin := batched.(safebatch.SafeBatchDBCloser)
	logger := tmlog.NewTMLogger(os.Stdout)
	encodingConfig := abci.MakeEncodingConfig()  

	// customize CMS to limit kv store's read height on query
	cms := rootmulti.NewStore(batched, hldb)
	vpr := viper.GetViper()

	app := abci.NewApp(
		logger,
		batched,
		nil,
		true, // need this so KVStores are set
		make(map[int64]bool),
		mantlemintConfig.Home,
		0,
		encodingConfig,
		vpr,
		fauxMerkleModeOpt,
		func(ba *baseapp.BaseApp) {
			ba.SetCMS(cms)
		},
	)

	// create app...
	appCreator := mantlemint.NewConcurrentQueryClientCreator(app)
	appConns := proxy.NewAppConns(appCreator)
	appConns.SetLogger(logger)
	if startErr := appConns.OnStart(); startErr != nil {
		panic(startErr)
	}

	go func() {
		a := <-appConns.Quit()
		fmt.Println(a)
	}()

	executor := mantlemint.NewMantlemintExecutor(batched, appConns.Consensus())
	mm := mantlemint.NewMantlemint(
		batched,
		appConns,
		executor,

		// run before
		nil,

		// RunAfter Inject callback
		nil,
	)

	// initialize using provided genesis
	genesisDoc := getGenesisDoc(mantlemintConfig.GenesisPath)
	initialHeight := genesisDoc.InitialHeight

	// set target initial write height to genesis.initialHeight;
	// this is safe as upon Inject it will be set with block.Height
	hldb.SetWriteHeight(initialHeight)
	batchedOrigin.Open()

	// initialize state machine with genesis
	if initErr := mm.Init(genesisDoc); initErr != nil {
		panic(initErr)
	}

	// flush to db; panic upon error (can't proceed)
	if rollback, flushErr := batchedOrigin.Flush(); flushErr != nil {
		debug.PrintStack()
		panic(flushErr)
	} else if rollback != nil {
		rollback.Close()
	}

	// load initial state to mantlemint
	if loadErr := mm.LoadInitialState(); loadErr != nil {
		panic(loadErr)
	}

	// initialization is done; clear write height
	hldb.ClearWriteHeight()

	// get blocks over some sort of transport, inject to mantlemint
	blockFeed := blockFeeder.NewAggregateBlockFeed(
		mm.GetCurrentHeight(),
		mantlemintConfig.RPCEndpoints,
		mantlemintConfig.WSEndpoints,
	)

	// create indexer service
	indexerInstance, indexerInstanceErr := indexer.NewIndexer(
		mantlemintConfig.IndexerDB, mantlemintConfig.Home, app, encodingConfig.TxConfig,
	)
	if indexerInstanceErr != nil {
		panic(indexerInstanceErr)
	}

	indexerInstance.RegisterIndexerService("tx", tx.IndexTx)
	indexerInstance.RegisterIndexerService("block", block.IndexBlock)

	abcicli, _ := appCreator.NewABCIClient()
	rpccli := rpc.NewRPCClient(abcicli)

	// rest cache invalidate channel
	cacheInvalidateChan := make(chan int64)

	// start RPC server
	rpcErr := rpc.StartRPC(
		app,
		rpccli,
		mantlemintConfig.ChainID,
		encodingConfig,
		cacheInvalidateChan,

		// callback for registering custom routers; primarily for indexers
		// default: noop,
		// todo: make this part injectable
		func(router *mux.Router) {
			indexerInstance.RegisterRESTRoute(router, tx.RegisterRESTRoute)
			indexerInstance.RegisterRESTRoute(router, block.RegisterRESTRoute)
		},

		// inject flag checker for synced
		blockFeed.IsSynced,
		mantlemintConfig,
	)

	if rpcErr != nil {
		panic(rpcErr)
	}

	// start subscribing to block
	//nolint:nestif
	if mantlemintConfig.DisableSync {
		fmt.Println("running without sync...")
		forever()
		return
	} else if cBlockFeed, blockFeedErr := blockFeed.Subscribe(0); blockFeedErr != nil {
		panic(blockFeedErr)
	} else {
		var rollbackBatch tmdb.Batch
		for {
			feed := <-cBlockFeed

			// open db batch
			hldb.SetWriteHeight(feed.Block.Height)
			batchedOrigin.Open()
			if injectErr := mm.Inject(feed.Block); injectErr != nil {
				// rollback last block
				if rollbackBatch != nil {
					fmt.Println("rollback previous block")
					rollbackBatch.WriteSync()
					rollbackBatch.Close()
				}

				debug.PrintStack()
				panic(injectErr)
			}

			// last block is okay -> dispose rollback batch
			if rollbackBatch != nil {
				rollbackBatch.Close()
				rollbackBatch = nil
			}

			// run indexer BEFORE batch flush
			if indexerErr := indexerInstance.Run(feed.Block, feed.BlockID, mm.GetCurrentEventCollector()); indexerErr != nil {
				debug.PrintStack()
				panic(indexerErr)
			}

			// flush db batch
			// returns rollback batch that reverts current block injection
			if rollback, flushErr := batchedOrigin.Flush(); flushErr != nil {
				debug.PrintStack()
				panic(flushErr)
			} else {
				rollbackBatch = rollback
			}

			hldb.ClearWriteHeight()

			cacheInvalidateChan <- feed.Block.Height
		}
	}
}

// Pass this in as an option to use a dbStoreAdapter instead of an IAVLStore for simulation speed.
func fauxMerkleModeOpt(app *baseapp.BaseApp) {
	app.SetFauxMerkleMode()
}

func getGenesisDoc(genesisPath string) *tendermint.GenesisDoc {
	jsonBlob, _ := os.ReadFile(genesisPath)
	shasum := sha1.New()
	shasum.Write(jsonBlob)
	sum := hex.EncodeToString(shasum.Sum(nil))

	log.Printf("[v0.34.x/sync] genesis shasum=%s", sum)

	if genesis, genesisErr := tendermint.GenesisDocFromFile(genesisPath); genesisErr != nil {
		panic(genesisErr)
	} else {
		return genesis
	}
}

func forever() {
	<-(chan int)(nil)
}
