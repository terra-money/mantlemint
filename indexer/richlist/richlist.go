package richlist

import (
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tm "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	terra "github.com/terra-money/core/v2/app"
	"github.com/terra-money/mantlemint/config"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
)

const (
	eventCoinSpent    = "coin_spent"
	eventCoinReceived = "coin_received"
	attrSpender       = "spender"
	attrReceiver      = "receiver"
	attrAmount        = "amount"
)

//var cdc = terra.MakeEncodingConfig()

// global/static and the latest richlist
// for now, we only handle a richlist for LUNA
var richlist = NewRichlist(0, &thresholdLuna)

var cfg = config.GetConfig()

var IndexRichlist = indexer.CreateIndexer(func(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector, app *terra.TerraApp) (err error) {
	height := uint64(block.Height)

	// skip if this indexer is disabled or at genesis height. genesis block cannot be parsed here.
	if cfg.RichlistLength == 0 || height == 1 {
		// nop
		return nil
	}
	defer fmt.Printf("[indexer/richlist] indexing done for height %d\n", block.Height)

	if height == 2 || richlist.Len() < cfg.RichlistLength { // genesis or lack of items
		fmt.Printf("[indexer/richlist] generate list from states... height:%d, len:%d\n", height, richlist.Len())
		list, err := generateRichlistFromState(indexerDB, block, blockID, evc, app, height-1, richlist.threshold.Denom)
		if err != nil {
			return err
		}
		richlist = list  // replace
		if height == 2 { // save previous richlist before apply changes from height 2
			// extract richlist to be saved
			extracted, err := richlist.Extract(1, cfg.RichlistLength, nil)
			if err != nil {
				return fmt.Errorf("failed to extract richlist: %v", err)
			}
			richlistJSON, err := json.Marshal(extracted)
			if err != nil {
				return err
			}
			err = indexerDB.Set(getDefaultKey(0), richlistJSON)
			if err != nil {
				return err
			}
		}
	}

	// gather balance-changing accounts from events
	events := append([]abci.Event{}, evc.ResponseBeginBlock.GetEvents()...)
	for _, tx := range evc.ResponseDeliverTxs {
		events = append(events, tx.GetEvents()...)
	}
	events = append(events, evc.ResponseEndBlock.GetEvents()...)
	fmt.Printf("[DEBUG] events: %d\n", len(events))
	changes, err := filterCoinChanges(events, defaultDenom)
	if err != nil {
		return err
	}
	fmt.Printf("[DEBUG] %d balance-changed accounts: %+v\n", len(changes), changes)

	// apply changes into richlist
	err = richlist.Apply(changes, app, height, defaultDenom)
	if err != nil {
		return err
	}
	richlist.Height = height

	// extract richlist to be saved
	extracted, err := richlist.Extract(height, cfg.RichlistLength, nil)
	if err != nil {
		return fmt.Errorf("failed to extract richlist: %v", err)
	}
	richlistJSON, err := json.Marshal(extracted)
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG] rank length: %d / saved: %d - at height %d\n", richlist.Len(), extracted.Len(), height)

	// save exracted one
	// save one for latest
	err = indexerDB.Set(getDefaultKey(0), richlistJSON)
	if err != nil {
		return err
	}
	// save another one for specific height
	return indexerDB.Set(getDefaultKey(height), richlistJSON)
})

func generateRichlistFromState(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector, app *terra.TerraApp, height uint64, denom string) (list *Richlist, err error) {
	threshold := sdk.NewCoin(denom, sdk.NewInt(threshold*decimal))
	list = NewRichlist(height, &threshold)

	ctx := app.NewContext(true, tmproto.Header{Height: int64(height)})
	// Should use lastest block time
	app.BankKeeper.IterateAllBalances(ctx, func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
		fmt.Printf("iterate: %s / %s\n", address.String(), coin.String())
		if coin.Denom != denom {
			return false
		}
		if err = list.Rank(Ranker{Addresses: []string{address.String()}, Score: coin}); err != nil {
			return true
		}
		fmt.Printf("rank from state - addr:%s amount:%s / len:%d\n", address.String(), coin.Amount.String(), list.Rankers.Len())
		return false // don't return true. true will halt this iteration
	})

	return list, err
}

// ranker.Amount will be used as differential
func filterCoinChanges(events []abci.Event, denom string) (addresses []changing, err error) {
	coinMap := make(map[string]sdk.Int)

	for _, event := range events {
		fmt.Printf("[DEBUG] EVENT: %s\n", event.GetType())
		for _, a := range event.GetAttributes() {
			fmt.Printf("[DEBUG]     ATTRIBUTES %s\n", a.String())
		}
		var address string
		var changing *sdk.Int

		if event.Type == eventCoinSpent {
			address, changing = extractChange(event.GetAttributes(), attrSpender, denom)
			if address == "" || changing == nil {
				return nil, fmt.Errorf("invalid event found: %+v", event.String())
			}
			prev, found := coinMap[address]
			fmt.Printf("[indexer/richlist/debug] decreasing: %s / %s - %s\n", address, prev.String(), changing.String())
			if !found {
				coinMap[address] = sdk.ZeroInt().Sub(*changing)
			} else {
				coinMap[address] = prev.Sub(*changing)
			}
		} else if event.Type == eventCoinReceived {
			address, changing = extractChange(event.GetAttributes(), attrReceiver, denom)
			if address == "" || changing == nil {
				return nil, fmt.Errorf("invalid event found: %+v", event.String())
			}
			prev, found := coinMap[address]
			fmt.Printf("[indexer/richlist/debug] increasing: %s / %s + %s\n", address, prev.String(), changing.String())
			if !found {
				coinMap[address] = *changing
			} else {
				coinMap[address] = prev.Add(*changing)
			}
		}
		// nop for other events
	}
	for addr, amount := range coinMap {
		addresses = append(addresses, changing{AccAddresses: []string{addr}, Amount: amount})
	}
	return
}

func extractChange(attrs []abci.EventAttribute, attributeKey string, denom string) (address string, amount *sdk.Int) {
	for _, attr := range attrs {
		key := string(attr.GetKey())
		if key == attributeKey {
			address = string(attr.GetValue())
		} else if key == attrAmount {
			//coin, err := sdk.ParseDecCoin(string(attr.GetValue()))
			coin, err := sdk.ParseCoinNormalized(string(attr.GetValue()))
			if err != nil {
				return "", nil
			}
			if coin.GetDenom() != denom {
				continue
			}
			amtInt := sdk.NewIntFromBigInt(coin.Amount.BigInt())
			amount = &amtInt
		}
		if address != "" && amount != nil {
			fmt.Printf("extracted change: %s / %s\n", address, amount.String())
			return
		}
	}
	fmt.Printf("extracted change: %s / %s\n", address, amount.String())
	return
}
