package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	tendermint "github.com/tendermint/tendermint/types"
	core "github.com/terra-money/core/types"
	"github.com/terra-money/mantlemint-provider-v0.34.x/config"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer/block"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer/tx"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint"
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	mantlemintConfig := config.NewConfig()

	viper.SetConfigType("toml")
	viper.SetConfigName("app")
	viper.AddConfigPath(filepath.Join(mantlemintConfig.Home, "config"))

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
	sdkConfig.Seal()

	// core related params
	mantlemintApp := mantlemint.NewMantlemintApp(mantlemintConfig.Home)
	mantlemintApp.
		WithChainID(mantlemintConfig.ChainID).
		WithBaseAppDecorators().
		WithIndexers("tx", tx.IndexTx, tx.RegisterRESTRoute).
		WithIndexers("block", block.IndexBlock, block.RegisterRESTRoute).

		// seal mantlemint
		Seal(func() *tendermint.GenesisDoc {
			return getGenesisDoc(mantlemintConfig.GenesisPath)
		}).
		Start(
			mantlemintConfig.RPCEndpoints,
			mantlemintConfig.WSEndpoints,
		)
}

func getGenesisDoc(genesisPath string) *tendermint.GenesisDoc {
	jsonBlob, _ := ioutil.ReadFile(genesisPath)
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
