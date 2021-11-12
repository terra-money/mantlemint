package rpc

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

	cache := NewCacheBackend(1024000)

	// register cache invalidator
	go func() {
		for {
			height := <-invalidateTrigger
			fmt.Printf("[cache-middleware] purging cache at height %d, lastLength=%d\n", height, cache.Purge())
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
			cache.HandleCachedHTTP(writer, request, next)
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
