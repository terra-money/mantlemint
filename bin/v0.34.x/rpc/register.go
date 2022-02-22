package rpc

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	terra "github.com/terra-money/core/app"
	"github.com/terra-money/core/app/params"
)

type RPCContext struct {
	server *api.Server
	config config.Config
}

func NewMantlemintRPCContext(
	app *terra.TerraApp,
	rpcclient rpcclient.Client,
	chainId string,
	codec params.EncodingConfig,
	invalidateTrigger chan int64,
	registerCustomRoutes func(router *mux.Router),
) *RPCContext {
	vp := viper.GetViper()
	cfg := config.GetConfig(vp)

	// create terra client; register all codecs
	context := client.
		Context{}.
		WithClient(rpcclient).
		WithCodec(codec.Marshaler).
		WithInterfaceRegistry(codec.InterfaceRegistry).
		WithTxConfig(codec.TxConfig).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithLegacyAmino(codec.Amino).
		WithHomeDir(terra.DefaultNodeHome).
		WithChainID(chainId)

	// create backends for response cache
	// - cache: used for latest states without `height` parameter
	// - archivalCache: used for historical states with `height` parameter; never flushed
	cache := NewCacheBackend(16384, "latest")
	archivalCache := NewCacheBackend(16384, "archival")

	// register cache invalidator
	go func() {
		for {
			height := <-invalidateTrigger
			fmt.Printf("[cache-middleware] purging cache at height %d\n", height)

			cache.Metric()
			archivalCache.Metric()

			// only purge latest cache
			cache.Purge()
		}
	}()

	// start new api server
	apiSrv := api.New(context, tmlog.NewTMLogger(ioutil.Discard))

	// register custom routes to default api server
	registerCustomRoutes(apiSrv.Router)

	// register all default GET routers...
	app.RegisterAPIRoutes(apiSrv, cfg.API)
	app.RegisterTendermintService(context)
	app.RegisterTxService(context)

	// caching middleware
	apiSrv.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			heightQuery := request.URL.Query().Get("height")
			height, err := strconv.ParseInt(heightQuery, 10, 64)

			// don't use archival cache if height is 0 or error
			if err == nil && height > 0 {
				// GRPC query parses height from header
				request.Header.Add("x-cosmos-block-height", heightQuery)
				archivalCache.HandleCachedHTTP(writer, request, next)
			} else {
				cache.HandleCachedHTTP(writer, request, next)
			}
		})
	})

	return &RPCContext{
		server: apiSrv,
		config: cfg,
	}
}

func (rpc *RPCContext) Start() error {
	errCh := make(chan error)

	// start api server in goroutine
	go func() {
		if err := rpc.server.Start(rpc.config); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(types.ServerStartTime): // assume server started successfully
	}

	return nil
}

func (rpc *RPCContext) GetRouter() (*mux.Router, *runtime.ServeMux) {
	return rpc.server.Router, rpc.server.GRPCGatewayRouter
}
