package mantlemint

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/state"
	tmdb "github.com/tendermint/tm-db"
	"github.com/terra-money/core/app/params"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/heleveldb"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proxy"
	tendermint "github.com/tendermint/tendermint/types"
	terra "github.com/terra-money/core/app"
	core "github.com/terra-money/core/types"
	wasmconfig "github.com/terra-money/core/x/wasm/config"
	blockFeeder "github.com/terra-money/mantlemint-provider-v0.34.x/block_feed"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/safe_batch"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer"
	"github.com/terra-money/mantlemint-provider-v0.34.x/rpc"
	"github.com/terra-money/mantlemint-provider-v0.34.x/store/rootmulti"
)

type MantlemintApp struct {
	chainID                string
	genesis                *tendermint.GenesisDoc
	baseappDecorators      []func(app *baseapp.BaseApp)
	currentStateCacheSize  int
	archivalStateCacheSize int
	home                   string

	beforeCallback MantlemintCallbackBefore
	afterCallback  MantlemintCallbackAfter

	// optional configs
	indexerTags    []string
	indexers       []indexer.IndexFunc
	indexerClients []indexer.RESTRouteRegisterer

	customRoutes      []RouteRegisterer
	apiRouterConsumer RouteRegistererDual

	terraApp      *terra.TerraApp
	clientCreator proxy.ClientCreator
	replaceAnte   func(terraApp *terra.TerraApp)

	// channels
	encodingConfig      params.EncodingConfig
	cacheInvalidateChan chan int64

	// internal instances
	mantlemint      Mantlemint
	executor        Executor
	blockFeeder     blockFeeder.BlockFeed
	indexerInstance *indexer.Indexer

	hldb      hld.HLD
	safeBatch safe_batch.SafeBatchDBCloser

	// additional handlers
	broadcastTxCommitHandler BroadcastTxCommitHandler
	broadcastTxSyncHandler   BroadcastTxSyncHandler
	broadcastTxAsyncHandler  BroadcastTxAsyncHandler
}

func NewMantlemintApp(home string) *MantlemintApp {
	viper.SetConfigType("toml")
	viper.SetConfigName("app")
	viper.AddConfigPath(filepath.Join(home, "config"))

	if err := viper.MergeInConfig(); err != nil {
		panic(fmt.Errorf("failed to merge configuration: %w", err))
	}

	sdkConfig := sdk.GetConfig()
	sdkConfig.SetCoinType(core.CoinType)
	sdkConfig.SetFullFundraiserPath(core.FullFundraiserPath)
	sdkConfig.SetBech32PrefixForAccount(core.Bech32PrefixAccAddr, core.Bech32PrefixAccPub)
	sdkConfig.SetBech32PrefixForValidator(core.Bech32PrefixValAddr, core.Bech32PrefixValPub)
	sdkConfig.SetBech32PrefixForConsensusNode(core.Bech32PrefixConsAddr, core.Bech32PrefixConsPub)
	sdkConfig.SetAddressVerifier(core.AddressVerifier)
	//sdkConfig.Seal()

	return &MantlemintApp{
		chainID:                  "",
		genesis:                  nil,
		baseappDecorators:        []func(*baseapp.BaseApp){},
		currentStateCacheSize:    16384,
		archivalStateCacheSize:   16384,
		home:                     home,
		beforeCallback:           nil,
		afterCallback:            nil,
		indexerTags:              []string{},
		indexers:                 []indexer.IndexFunc{},
		indexerClients:           []indexer.RESTRouteRegisterer{},
		customRoutes:             []RouteRegisterer{},
		apiRouterConsumer:        nil,
		terraApp:                 nil,
		clientCreator:            nil,
		replaceAnte:              nil,
		encodingConfig:           terra.MakeEncodingConfig(),
		cacheInvalidateChan:      make(chan int64),
		mantlemint:               nil,
		executor:                 nil,
		blockFeeder:              nil,
		indexerInstance:          nil,
		hldb:                     nil,
		safeBatch:                nil,
		broadcastTxCommitHandler: nil,
		broadcastTxSyncHandler:   nil,
		broadcastTxAsyncHandler:  nil,
	}
}

func (app *MantlemintApp) WithChainID(chainID string) *MantlemintApp {
	app.chainID = chainID
	return app
}

func (app *MantlemintApp) WithGenesisPath(genesisPath string) *MantlemintApp {
	jsonBlob, _ := ioutil.ReadFile(genesisPath)
	shasum := sha1.New()
	shasum.Write(jsonBlob)
	sum := hex.EncodeToString(shasum.Sum(nil))

	log.Printf("[v0.34.x/sync] genesis shasum=%s", sum)

	if genesis, genesisErr := tendermint.GenesisDocFromFile(genesisPath); genesisErr != nil {
		panic(genesisErr)
	} else {
		app.genesis = genesis
	}

	return app
}

func (app *MantlemintApp) WithCurrentStateCacheSize(cacheSize int) *MantlemintApp {
	app.currentStateCacheSize = cacheSize
	return app
}

func (app *MantlemintApp) WithArchivalStateCacheSize(cacheSize int) *MantlemintApp {
	app.archivalStateCacheSize = cacheSize
	return app
}

func (app *MantlemintApp) WithBaseAppDecorators(decorators ...func(app *baseapp.BaseApp)) *MantlemintApp {
	app.baseappDecorators = decorators
	return app
}

func (app *MantlemintApp) WithCustomAnteHandler(replaceAnte func(terraApp *terra.TerraApp)) *MantlemintApp {
	app.replaceAnte = replaceAnte
	return app
}

func (app *MantlemintApp) WithBeforeBlockCallback(before MantlemintCallbackBefore) *MantlemintApp {
	app.beforeCallback = before
	return app
}

func (app *MantlemintApp) WithAfterBlockCallback(after MantlemintCallbackAfter) *MantlemintApp {
	app.afterCallback = after
	return app
}

func (app *MantlemintApp) WithIndexers(tag string, indexFunc indexer.IndexFunc, routeRegisterer indexer.RESTRouteRegisterer) *MantlemintApp {
	app.indexerTags = append(app.indexerTags, tag)
	app.indexers = append(app.indexers, indexFunc)
	app.indexerClients = append(app.indexerClients, routeRegisterer)
	return app
}

func (app *MantlemintApp) WithCustomRoutes(registerer RouteRegisterer) *MantlemintApp {
	app.customRoutes = append(app.customRoutes, registerer)
	return app
}

func (app *MantlemintApp) WithAPIRouterConsumer(apiRouterConsumer RouteRegistererDual) *MantlemintApp {
	app.apiRouterConsumer = apiRouterConsumer
	return app
}

func (app *MantlemintApp) GetBlockFeeder() blockFeeder.BlockFeed {
	return app.blockFeeder
}

func (app *MantlemintApp) GetTerraApp() *terra.TerraApp {
	return app.terraApp
}

func (app *MantlemintApp) GetEncodingConfig() params.EncodingConfig {
	return app.encodingConfig
}

func (app *MantlemintApp) GetLastState() state.State {
	return app.mantlemint.GetCurrentState()
}

func (app *MantlemintApp) WithRunBefore(runBefore MantlemintCallbackBefore) *MantlemintApp {
	app.beforeCallback = runBefore
	return app
}

func (app *MantlemintApp) WithRunAfter(runAfter MantlemintCallbackAfter) *MantlemintApp {
	app.afterCallback = runAfter
	return app
}

func (app *MantlemintApp) WithBroadcastTxCommitHandler(handler BroadcastTxCommitHandler) *MantlemintApp {
	app.broadcastTxCommitHandler = handler
	return app
}

func (app *MantlemintApp) WithBroadcastTxSyncHandler(handler BroadcastTxSyncHandler) *MantlemintApp {
	app.broadcastTxSyncHandler = handler
	return app
}

func (app *MantlemintApp) WithBroadcastTxAsyncHandler(handler BroadcastTxAsyncHandler) *MantlemintApp {
	app.broadcastTxAsyncHandler = handler
	return app
}

func (app *MantlemintApp) Seal(genesisGetter func() *tendermint.GenesisDoc) *MantlemintApp {
	ldb, ldbErr := heleveldb.NewLevelDBDriver(&heleveldb.DriverConfig{
		Name: "mantlemint",
		Dir:  app.home,
		Mode: heleveldb.DriverModeKeySuffixDesc,
	})
	if ldbErr != nil {
		panic(ldbErr)
	}

	var hldb = hld.ApplyHeightLimitedDB(
		ldb,
		&hld.HeightLimitedDBConfig{
			Debug: true,
		},
	)

	batched := safe_batch.NewSafeBatchDB(hldb)
	batchedOrigin := batched.(safe_batch.SafeBatchDBCloser)
	logger := tmlog.NewTMLogger(os.Stdout)
	codec := app.encodingConfig

	var terraApp, terraAppErr = createTerraApp(hldb, batchedOrigin, logger, app.home, codec, app.baseappDecorators, app.replaceAnte)
	if terraAppErr != nil {
		panic(errors.Wrapf(terraAppErr, "error during creating app"))
	}

	app.terraApp = terraApp
	app.hldb = hldb
	app.safeBatch = batchedOrigin

	// create app...
	var appCreator = NewConcurrentQueryClientCreator(terraApp)
	appConns := proxy.NewAppConns(appCreator)
	appConns.SetLogger(logger)
	if startErr := appConns.OnStart(); startErr != nil {
		panic(startErr)
	}
	app.clientCreator = appCreator

	go func() {
		a := <-appConns.Quit()
		fmt.Println(a)
	}()

	var executor = NewMantlemintExecutor(batched, appConns.Consensus())
	var mm = NewMantlemint(
		batched,
		appConns,
		executor,

		// run before
		app.beforeCallback,

		// RunAfter Inject callback
		app.afterCallback,
	)

	app.mantlemint = mm
	app.executor = executor

	if loadLatestErr := app.terraApp.LoadLatestVersion(); loadLatestErr != nil {
		panic(errors.Wrapf(loadLatestErr, "error during starting mantlemint app"))
	}

	app.genesis = genesisGetter()

	// initialize using provided genesis
	initialHeight := app.genesis.InitialHeight

	// set target initial write height to genesis.initialHeight;
	// this is safe as upon Inject it will be set with block.Height
	app.hldb.SetWriteHeight(initialHeight)
	app.safeBatch.Open()

	// initialize state machine with genesis
	if initErr := app.mantlemint.Init(app.genesis); initErr != nil {
		panic(initErr)
	}

	// flush to db; panic upon error (can't proceed)
	if rollback, flushErr := app.safeBatch.Flush(); flushErr != nil {
		debug.PrintStack()
		panic(flushErr)
	} else if rollback != nil {
		rollback.Close()
	}

	// load initial state to mantlemint
	if loadErr := app.mantlemint.LoadInitialState(); loadErr != nil {
		panic(loadErr)
	}

	// initialization is done; clear write height
	app.hldb.ClearWriteHeight()

	// create indexer service
	// - sets app.indexer
	indexerInstance, indexerErr := startIndexer(app.home, app.indexerTags, app.indexers)
	if indexerErr != nil {
		panic(errors.Wrapf(indexerErr, "error during indexer initialization"))
	}
	app.indexerInstance = indexerInstance

	return app
}

func (app *MantlemintApp) Start(rpcEndpoints, wsEndpoints []string) {
	// get blocks over some sort of transport, inject to mantlemint
	// - sets app.blockFeeder
	feeder, feederErr := startBlockFeeder(
		app.mantlemint.GetCurrentHeight(),
		rpcEndpoints,
		wsEndpoints,
	)
	if feederErr != nil {
		panic(errors.Wrapf(feederErr, "error during block feeder initialization"))
	}
	app.blockFeeder = feeder

	abciClient, _ := app.clientCreator.NewABCIClient()
	rpcClient := NewRPCClient(
		abciClient,
		app.broadcastTxCommitHandler,
		app.broadcastTxSyncHandler,
		app.broadcastTxAsyncHandler,
	)

	// start rpc server
	// port binding is from app.toml
	if rpcErr := startRPCServer(
		app.terraApp,
		app.chainID,
		app.encodingConfig,
		rpcClient,
		app.cacheInvalidateChan,
		app.indexerInstance,
		app.indexerClients,
		app.customRoutes,
		app.blockFeeder.IsSynced,
		app.apiRouterConsumer,
	); rpcErr != nil {
		panic(errors.Wrapf(rpcErr, "error during rpc server initialization"))
	}

	cBlockFeed := app.blockFeeder.GetBlockFeedChannel()
	go func() {
		var rollbackBatch tmdb.Batch
		for {
			feed := <-cBlockFeed

			// open db batch
			app.hldb.SetWriteHeight(feed.Block.Height)
			app.safeBatch.Open()
			if injectErr := app.mantlemint.Inject(feed.Block); injectErr != nil {

				// rollback last block
				if rollbackBatch != nil {
					log.Println("rolling back to previous block..")
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

			// flush db batch
			if rollback, flushErr := app.safeBatch.Flush(); flushErr != nil {
				debug.PrintStack()
				panic(flushErr)
			} else if rollback != nil {
				rollbackBatch = rollback
			}

			app.hldb.ClearWriteHeight()

			// run indexer
			if indexerErr := app.indexerInstance.Run(feed.Block, feed.BlockID, app.mantlemint.GetCurrentEventCollector()); indexerErr != nil {
				debug.PrintStack()
				panic(indexerErr)
			}

			app.cacheInvalidateChan <- feed.Block.Height
		}
	}()
}

func (app *MantlemintApp) Subscribe() {
	if _, blockFeedErr := app.blockFeeder.Subscribe(); blockFeedErr != nil {
		panic(errors.Wrapf(blockFeedErr, "error during establishing subscription in blockFeeder"))
	} else {
		forever()
	}
}

func createTerraApp(
	hldb *hld.HeightLimitedDB,
	batched tmdb.DB,
	logger tmlog.Logger,
	home string,
	codec params.EncodingConfig,
	baseAppDecorators []func(app *baseapp.BaseApp),
	replaceAnte func(app *terra.TerraApp),
) (*terra.TerraApp, error) {
	// customize CMS to limit kv store's read height on query
	decorators := []func(app *baseapp.BaseApp){
		decorateFauxMerkleMode(),
		decorateCMS(rootmulti.NewStore(batched, hldb)),
	}

	for _, decorator := range baseAppDecorators {
		baseAppDecorators = append(baseAppDecorators, decorator)
	}

	var terraApp = terra.NewTerraApp(
		logger,
		batched,
		nil,
		false,
		make(map[int64]bool),
		home,
		0,
		codec,
		simapp.EmptyAppOptions{},
		&wasmconfig.Config{
			ContractQueryGasLimit:   3000000,
			ContractDebugMode:       false,
			ContractMemoryCacheSize: 1024,
		},
		decorators...,
	)

	replaceAnte(terraApp)

	return terraApp, nil
}

func startBlockFeeder(
	currentHeight int64,
	rpcEndpoints []string,
	wsEndpoints []string,
) (blockFeeder.BlockFeed, error) {
	rpcSubscription, rpcSubscriptionErr := blockFeeder.NewRpcSubscription(rpcEndpoints)
	if rpcSubscriptionErr != nil {
		return nil, rpcSubscriptionErr
	}

	wsSubscription, wsSubscriptionErr := blockFeeder.NewWSSubscription(wsEndpoints)
	if wsSubscriptionErr != nil {
		return nil, wsSubscriptionErr
	}

	bf := blockFeeder.NewAggregateBlockFeed(
		currentHeight,
		rpcSubscription,
		wsSubscription,
	)

	return bf, nil
}

func startRPCServer(
	terraApp *terra.TerraApp,
	chainID string,
	codec params.EncodingConfig,
	rpcClient rpcclient.Client,
	cacheInvalidateChan chan int64,
	indexerInstance *indexer.Indexer,
	indexerClients []indexer.RESTRouteRegisterer,
	customRoutes []RouteRegisterer,
	getSyncState func() bool,
	indirectSync func(*mux.Router, *runtime.ServeMux),
) error {

	// start RPC server
	rpcCtx := rpc.NewMantlemintRPCContext(
		terraApp,
		rpcClient,
		chainID,
		codec,
		cacheInvalidateChan,

		// callback for registering custom routers; primarily for indexers
		// default: noop,
		// todo: make this part injectable
		func(router *mux.Router) {
			// register indexer-related routes
			for _, registerer := range indexerClients {
				indexerInstance.RegisterRESTRoute(router, registerer)
			}

			// register custom routes
			for _, registerer := range customRoutes {
				registerer(router)
			}
		})

	// if indirectSync, a router consumer, is provided, call that function instead of starting rpc server directly
	if indirectSync != nil {
		indirectSync(rpcCtx.GetRouter())
		return nil
	} else {
		return rpcCtx.Start()
	}
}

func startIndexer(home string, indexerTags []string, indexFuncs []indexer.IndexFunc) (*indexer.Indexer, error) {
	indexerInstance, indexerInstanceErr := indexer.NewIndexer("indexer", home)
	if indexerInstanceErr != nil {
		return nil, indexerInstanceErr
	}

	for indexI, indexFunc := range indexFuncs {
		indexerInstance.RegisterIndexerService(indexerTags[indexI], indexFunc)
	}

	return indexerInstance, nil
}

func forever() {
	<-(chan int)(nil)
}
