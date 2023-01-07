package rpc

//nolint:staticcheck
import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gorilla/mux"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	"github.com/spf13/viper"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	terra "github.com/terra-money/alliance/app"
	mconfig "github.com/terra-money/mantlemint/config"
)

//nolint:funlen
func StartRPC(
	app *cosmoscmd.App,
	rpcclient rpcclient.Client,
	chainID string,
	encodingConfig cosmoscmd.EncodingConfig,
	invalidateTrigger chan int64,
	registerCustomRoutes func(router *mux.Router),
	getIsSynced func() bool,
	mantlemintConfig *mconfig.Config,
) error {
	vp := viper.GetViper()
	cfg, _ := config.GetConfig(vp)
	// create terra client; register all codecs
	context := client.
		Context{}.
		WithClient(rpcclient).
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithLegacyAmino(encodingConfig.Amino).
		WithHomeDir(terra.DefaultNodeHome).
		WithChainID(chainID)

	// create backends for response cache
	// - cache: used for latest states without `height` parameter
	// - archivalCache: used for historical states with `height` parameter; never flushed
	cache := NewCacheBackend(16384, "latest")
	archivalCache := NewCacheBackend(16384, "archival")

	// register cache invalidator
	go func() {
		for {
			height := <-invalidateTrigger
			//nolint:forbidigo
			fmt.Printf("[cache-middleware] purging cache at height %d\n", height)

			cache.Metric()
			archivalCache.Metric()

			// only purge latest cache
			cache.Purge()
		}
	}()

	// start new api server
	apiSrv := api.New(context, tmlog.NewTMLogger(io.Discard))

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

	//// register all default GET routers...
	(*app).RegisterAPIRoutes(apiSrv, cfg.API)
	(*app).RegisterTendermintService(context)
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

	pm := NewProxyMiddleware(mantlemintConfig.LCDEndpoints)
	// proxy middleware to handle unimplemented queries
	apiSrv.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			pm.HandleRequest(writer, request, next)
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
