package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	terra "github.com/terra-money/core/v2/app"
)

type Config struct {
	GenesisPath        string
	Home               string
	ChainID            string
	RPCEndpoints       []string
	WSEndpoints        []string
	MantlemintDB       string
	IndexerDB          string
	DisableSync        bool
	EnableExportModule bool
	RichlistLength     int
	RichlistThreshold  int64
}

var singleton Config

func init() {
	singleton = newConfig()
}

// GetConfig returns singleton config
func GetConfig() *Config {
	return &singleton
}

// newConfig converts envvars into consumable config chunks
func newConfig() Config {
	cfg := Config{
		// GenesisPath sets the location of genesis
		GenesisPath: getValidEnv("GENESIS_PATH"),

		// Home sets where the default terra home is.
		Home: getValidEnv("MANTLEMINT_HOME"),

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

		EnableExportModule: func() bool {
			enableExport := getValidEnv("ENABLE_EXPORT_MODULE")
			return enableExport == "true"
		}(),

		// RichlistLength have to be greater than or equal to 0, richlist function will be off if length is 0
		RichlistLength: func() int {
			lengthStr := getValidEnv("RICHLIST_LENGTH")
			length, err := strconv.Atoi(lengthStr)
			if err != nil {
				panic(err)
			}
			if length < 0 {
				panic(fmt.Errorf("RICHLIST_LENGTH(%s) is invalid", lengthStr))
			}
			return length
		}(),

		// RichlistThreshold (unit:LUNA, not uluna)
		RichlistThreshold: func() int64 {
			// don't need to read threshold env if the length of richlist is 0
			lengthStr := getValidEnv("RICHLIST_LENGTH")
			length, _ := strconv.Atoi(lengthStr)
			if length == 0 {
				return 0
			}

			thresholdStr := getValidEnv("RICHLIST_THRESHOLD")
			threshold, err := strconv.ParseInt(thresholdStr, 10, 64)
			if err != nil {
				panic(err)
			}
			if threshold < 1 {
				panic(fmt.Errorf("RICHLIST_THRESHOLD(%s) is invalid", thresholdStr))
			}
			return threshold
		}(),
	}

	viper.SetConfigType("toml")
	viper.SetConfigName("app")
	viper.AutomaticEnv()
	viper.AddConfigPath(filepath.Join(cfg.Home, "config"))

	pflag.Bool(crisis.FlagSkipGenesisInvariants, false, "Skip x/crisis invariants check on startup")
	pflag.Parse()
	if bindErr := viper.BindPFlags(pflag.CommandLine); bindErr != nil {
		panic(bindErr)
	}

	if err := viper.MergeInConfig(); err != nil {
		panic(fmt.Errorf("failed to merge configuration: %w", err))
	}

	return cfg
}

func (cfg Config) Print() {
	fmt.Printf("%+v\n", cfg)
}

func getValidEnv(tag string) string {
	if e := os.Getenv(tag); e == "" {
		panic(fmt.Errorf("environment variable %s not set; expected string, got %s \"\"", tag, e))
	} else {
		return e
	}
}
