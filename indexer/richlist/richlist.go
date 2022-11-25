package richlist

import (
	"encoding/json"
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tm "github.com/tendermint/tendermint/types"
	terra "github.com/terra-money/alliance/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/mantlemint/config"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
)

const (
	haltOnApplyFailure = false
)

const (
	eventCoinSpent         = "coin_spent"
	eventCoinReceived      = "coin_received"
	eventCompleteUnbonding = "complete_unbonding"
	attrSpender            = "spender"
	attrReceiver           = "receiver"
	attrAmount             = "amount"
	attrDelegator          = "delegator"
)

var cfg = config.GetConfig()

// for now, we only handle a richlist for LUNA
var richlist = NewRichlist(0, cfg.RichlistThreshold)

var IndexRichlist = indexer.CreateIndexer(func(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector, app *cosmoscmd.App) (err error) {
	height := uint64(block.Height)
	// skip if this indexer is disabled or at genesis height. genesis block cannot be parsed here.
	if cfg.RichlistLength == 0 || height == 1 {
		// nop
		return nil
	}
	defer fmt.Printf("[indexer/richlist] indexing done for height %d - %d items are in richlist\n", block.Height, richlist.Len())

	if height == 2 || richlist.Len() < cfg.RichlistLength { // genesis or lack of items
		fmt.Printf("[indexer/richlist] generate list from states... height:%d, len:%d\n", height, richlist.Len())
		list, err := generateRichlistFromState(indexerDB, block, blockID, evc, app, height-1, *richlist.threshold)
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
			err = indexerDB.Set(getDefaultKey(1), richlistJSON)
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
	changes, err := filterCoinChanges(events, richlist.threshold.Denom)
	if err != nil {
		return err
	}

	// apply changes into richlist
	err = richlist.Apply(changes, app, height, richlist.threshold.Denom)
	if err != nil && haltOnApplyFailure {
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

	// save exracted one
	// save one for latest
	err = indexerDB.Set(getDefaultKey(0), richlistJSON)
	if err != nil {
		return err
	}
	// save another one for specific height
	return indexerDB.Set(getDefaultKey(height), richlistJSON)
})

func generateRichlistFromState(indexerDB safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector, capp *cosmoscmd.App, height uint64, threshold sdk.Coin) (list *Richlist, err error) {
	app, ok := (*capp).(*terra.App)
	if !ok {
		return nil, fmt.Errorf("invalid app expect: %T got %T", terra.App{}, capp)
	}
	list = NewRichlist(height, &threshold)
	ctx := app.NewContext(true, tmproto.Header{Height: int64(height)})

	app.AccountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		if _, isModule := account.(*authtypes.ModuleAccount); isModule {
			return false
		}
		balance := app.BankKeeper.GetBalance(ctx, account.GetAddress(), threshold.Denom)
		if err = list.Rank(Ranker{Addresses: []string{account.GetAddress().String()}, Score: balance}); err != nil {
			return true // stop iteration and return err
		}
		return false
	})

	return list, err
}

// ranker.Amount will be used as differential
func filterCoinChanges(events []abci.Event, denom string) (coinMap map[string]sdk.Int, err error) {
	coinMap = make(map[string]sdk.Int)

	for _, event := range events {
		var address string
		var changing *sdk.Int

		switch event.Type {
		case eventCoinSpent:
			address, changing = extractChange(event.GetAttributes(), attrSpender, denom)
			if address == "" || changing == nil {
				fmt.Printf("invalid spent event found: %+v\n", event.String())
				continue
			}
			prev, found := coinMap[address]
			if !found {
				coinMap[address] = sdk.ZeroInt().Sub(*changing)
			} else {
				coinMap[address] = prev.Sub(*changing)
			}
		case eventCoinReceived:
			address, changing = extractChange(event.GetAttributes(), attrReceiver, denom)
			if address == "" || changing == nil {
				fmt.Printf("invalid receive event found: %+v\n", event.String())
				continue
			}
			prev, found := coinMap[address]
			if !found {
				coinMap[address] = *changing
			} else {
				coinMap[address] = prev.Add(*changing)
			}
		}
		// nop for other events
	}
	return
}

func extractChange(attrs []abci.EventAttribute, attributeKey string, denom string) (address string, amount *sdk.Int) {
	for _, attr := range attrs {
		key := string(attr.GetKey())
		if key == attributeKey {
			address = string(attr.GetValue())
		} else if key == attrAmount {
			coins, err := sdk.ParseCoinsNormalized(string(attr.GetValue()))
			if err != nil {
				return "", nil
			}
			for _, coin := range coins {
				if coin.GetDenom() != denom {
					continue
				}
				amtInt := sdk.NewIntFromBigInt(coin.Amount.BigInt())
				amount = &amtInt
			}
		}
		if address != "" && amount != nil {
			return
		}
	}
	return
}
