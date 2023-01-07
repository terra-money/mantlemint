package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	GenesisPath          string
	Home                 string
	ChainID              string
	RPCEndpoints         []string
	WSEndpoints          []string
	LCDEndpoints         []string
	MantlemintDB         string
	IndexerDB            string
	DisableSync          bool
	EnableExportModule   bool
	RichlistLength       int
	AccountAddressPrefix string
	BondDenom            string
	RichlistThreshold    *sdk.Coin
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
		//Feather chains are going to have different prefixes
		AccountAddressPrefix: getValidEnv("ACCOUNT_ADDRESS_PREFIX"),
		//Feather chains are going to have different denoms
		BondDenom: getValidEnv("BOND_DENOM"),
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

		// LCDEndpoints is where to forward unhandled queries to a node
		LCDEndpoints: func() []string {
			endpoints := getValidEnv("LCD_ENDPOINTS")
			return strings.Split(endpoints, ",")
		}(),

		// MantlemintDB is the db name for mantlemint. Defaults to mantlemint
		MantlemintDB: func() string {
			mantlemintDB := getValidEnv("MANTLEMINT_DB")
			if mantlemintDB == "" {
				return "mantlemint"
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

		// RichlistThreshold (format: {amount}{denom} like 1000000000000uluna)
		RichlistThreshold: func() *sdk.Coin {
			// don't need to read threshold env if the length of richlist is 0
			lengthStr := getValidEnv("RICHLIST_LENGTH")
			length, _ := strconv.Atoi(lengthStr)
			if length == 0 {
				return nil
			}

			thresholdCoin, err := sdk.ParseCoinNormalized(getValidEnv("RICHLIST_THRESHOLD"))
			if err != nil {
				panic(fmt.Errorf("RICHLIST_THRESHOLD is invalid: %v", err))
			}
			return &thresholdCoin
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
