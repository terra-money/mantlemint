package config

import (
	"fmt"
	terra "github.com/terra-money/core/app"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	GenesisPath         string
	Home                string
	ChainID             string
	RPCEndpoints        []string
	MantlemintDB        string
	IndexerDB           string
	IndexerSideSyncPort int64
	DisableSync         bool
}

// NewConfig converts envvars into consumable config chunks
func NewConfig() Config {
	return Config{
		// GenesisPath sets the location of genesis
		GenesisPath: os.Getenv("GENESIS_PATH"),

		// Home sets where the default terra home is.
		Home: os.Getenv("HOME"),

		// ChainID sets expected chain id for this mantlemint instance
		ChainID: os.Getenv("CHAIN_ID"),

		// RPCEndpoint is where to pull txs from when fast-syncing
		RPCEndpoints: func() []string {
			endpoints := os.Getenv("RPC_ENDPOINTS")
			return strings.Split(endpoints, ",")
		}(),

		// MantlemintDB is the db name for mantlemint. Defaults to terra.DefaultHome
		MantlemintDB: func() string {
			mantlemintDB := os.Getenv("MANTLEMINT_DB")
			if mantlemintDB == "" {
				return terra.DefaultNodeHome
			} else {
				return mantlemintDB
			}
		}(),

		// IndexerDB is the db name for indexed data
		IndexerDB: os.Getenv("INDEXER_DB"),

		// IndexerSideSyncPort is
		IndexerSideSyncPort: func() int64 {
			port, portErr := strconv.Atoi(os.Getenv("INDEXER_SIDESYNC_PORT"))
			if portErr != nil {
				panic(portErr)
			}
			return int64(port)
		}(),

		// DisableSync sets a flag where if true mantlemint won't accept any blocks (usually for debugging)
		DisableSync: func() bool {
			disableSync := os.Getenv("DISABLE_SYNC")
			return disableSync == "true"
		}(),
	}
}

func (cfg Config) Print() {
	fmt.Println(cfg)
}
