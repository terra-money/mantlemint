package config

import (
	"fmt"
	terra "github.com/terra-money/core/app"
	"os"
	"strings"
)

type Config struct {
	GenesisPath         string
	Home                string
	ChainID             string
	RPCEndpoints        []string
	WSEndpoints         []string
	MantlemintDB        string
	IndexerDB           string
	DisableSync         bool
}

// NewConfig converts envvars into consumable config chunks
func NewConfig() Config {
	return Config{
		// GenesisPath sets the location of genesis
		GenesisPath: getValidEnv("GENESIS_PATH"),

		// Home sets where the default terra home is.
		Home: getValidEnv("HOME"),

		// ChainID sets expected chain id for this mantlemint instance
		ChainID: getValidEnv("CHAIN_ID"),

		// RPCEndpoints is where to pull txs from when fast-syncing
		RPCEndpoints: func() []string {
			endpoints := getValidEnv("RPC_ENDPOINTS")
			return strings.Split(endpoints, ",")
		}(),

		// WSEndpoints is where to pull txs from when normal syncing
		WSEndpoints: func() []string {
			endpoints := getValidEnv("WS_ENDPOINTS")
			return strings.Split(endpoints, ",")
		}(),

		// MantlemintDB is the db name for mantlemint. Defaults to terra.DefaultHome
		MantlemintDB: func() string {
			mantlemintDB := getValidEnv("MANTLEMINT_DB")
			if mantlemintDB == "" {
				return terra.DefaultNodeHome
			} else {
				return mantlemintDB
			}
		}(),

		// IndexerDB is the db name for indexed data
		IndexerDB: getValidEnv("INDEXER_DB"),

		// DisableSync sets a flag where if true mantlemint won't accept any blocks (usually for debugging)
		DisableSync: func() bool {
			disableSync := getValidEnv("DISABLE_SYNC")
			return disableSync == "true"
		}(),
	}
}

func (cfg Config) Print() {
	fmt.Println(cfg)
}

func getValidEnv(tag string) string {
	if e := os.Getenv(tag); e == "" {
		panic(fmt.Errorf("environment variable %s not set; expected string, got %s \"\"", tag, e))
	} else {
		return e
	}
}
