package rpc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	terra "github.com/classic-terra/core/app"
	"github.com/classic-terra/core/app/params"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

func StartRPC(
	app *terra.TerraApp,
	rpcclient rpcclient.Client,
	chainId string,
	codec params.EncodingConfig,
	invalidateTrigger chan int64,
	registerCustomRoutes func(router *mux.Router),
	getIsSynced func() bool,
) error {
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

	// custom healthcheck endpoint
	apiSrv.Router.Handle("/health", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		isSynced := getIsSynced()
		if isSynced {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
		} else {
			writer.WriteHeader(http.StatusServiceUnavailable)
			writer.Write([]byte("NOK"))
		}
	})).Methods("GET")

	// register all default GET routers...
	app.RegisterAPIRoutes(apiSrv, cfg.API)
	app.RegisterTendermintService(context)
	errCh := make(chan error)

	// caching middleware
	apiSrv.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/health" {
				next.ServeHTTP(writer, request)
				return
			}

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

	// start api server in goroutine
	go func() {
		if err := apiSrv.Start(cfg); err != nil {
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
