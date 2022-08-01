package richlist

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/google/btree"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	terra "github.com/terra-money/core/v2/app"
	"github.com/terra-money/mantlemint/lib"
)

// keys for storage
var (
	prefix = []byte("richlist/height:")
	getKey = func(height uint64, denom string) []byte {
		return lib.ConcatBytes(prefix, lib.UintToBigEndian(height), []byte(":"), []byte(denom))
	}
	getDefaultKey = func(height uint64) []byte {
		return getKey(height, cfg.RichlistThreshold.Denom)
	}
)

// Ranker

type Ranker struct {
	Position int `json:"position,omitempty"` // use only for marshaling
	// a ranker can have multiple addresses because btree cannot handle multiple keys(amount)
	Addresses []string `json:"address"`
	Score     sdk.Coin `json:"score"`
}

func (ranker Ranker) Less(than btree.Item) bool {
	return ranker.Score.IsLT(than.(Ranker).Score)
}

// multiple accounts with exactly same amount of luna? rare, but may exist
func (ranker *Ranker) Enlist(accAddress string) {
	for _, listed := range ranker.Addresses {
		if listed == accAddress {
			return // already listed
		}
	}
	ranker.Addresses = append(ranker.Addresses, accAddress)
}

func (ranker *Ranker) Unlist(accAddress string) ( /*isEmpty*/ bool /*err*/, error) {
	for i, listed := range ranker.Addresses {
		if listed == accAddress {
			ranker.Addresses = append(ranker.Addresses[:i], ranker.Addresses[i+1:]...)
			return len(ranker.Addresses) <= 0, nil
		}
	}
	return false, fmt.Errorf("%s is not on the ranker list", accAddress)
}

// Richlist

type Richlist struct {
	Height uint64 `json:",omitempty"`
	// NOTE: NOT thread-safe for write operations whereas read operaions are safe.
	//       mutex will be needed if indexer runs in async
	Rankers *btree.BTree `json:"rankers"`
	// internal use only. keep it private
	threshold *sdk.Coin
}

func NewRichlist(height uint64, threshold *sdk.Coin) *Richlist {
	return &Richlist{
		Height:    height,
		threshold: threshold,
		Rankers:   btree.New(2),
	}
}

func (list *Richlist) Rank(ranker Ranker) (err error) {
	if ranker.Score.IsLT(*list.threshold) {
		return nil
	}
	item := list.Rankers.Get(ranker)
	var ranked Ranker
	if item != nil {
		ranked = item.(Ranker)
		for _, addr := range ranker.Addresses {
			ranked.Enlist(addr)
		}
	} else { // new entry
		ranked = ranker
	}
	list.Rankers.ReplaceOrInsert(ranked)
	return nil
}

func (list *Richlist) Unrank(ranker Ranker) (err error) {
	if ranker.Score.Amount.LT(list.threshold.Amount) {
		return nil
	}
	item := list.Rankers.Get(ranker)
	if item == nil {
		return fmt.Errorf("failed to unkrank: cannot find %+v", ranker)
	}

	unranked := item.(Ranker)
	for _, addr := range ranker.Addresses {
		unranked.Unlist(addr)
	}

	if len(unranked.Addresses) == 0 {
		list.Rankers.Delete(unranked)
	} else {
		list.Rankers.ReplaceOrInsert(unranked)
	}
	return nil
}

func (list Richlist) Min() Ranker {
	return list.Rankers.Min().(Ranker)
}

func (list Richlist) Max() Ranker {
	return list.Rankers.Max().(Ranker)
}

// Len returns the number of tree entries
func (list Richlist) Len() int {
	return list.Rankers.Len()
}

// Count returns the number of ranker in the list
func (list Richlist) Count() (count int) {
	list.Rankers.Descend(func(entry btree.Item) bool {
		count += len(entry.(Ranker).Addresses)
		return true
	})
	return
}

func (list Richlist) Extract(height uint64, len int, threshold *sdk.Coin) (extracted *Richlist, err error) {
	if list.Len() < 1 {
		return nil, fmt.Errorf("richlist empty")
	}
	temp := Ranker{}
	if threshold != nil {
		temp.Score = *threshold
	}
	extracted = NewRichlist(height, nil)
	list.Rankers.Descend(func(entry btree.Item) bool {
		ranker := entry.(Ranker)
		if (extracted.Count() >= len) || (threshold != nil && ranker.Less(temp)) {
			return false
		}
		extracted.Rankers.ReplaceOrInsert(ranker)
		return true
	})
	return
}

func (list *Richlist) Apply(changes map[string]sdk.Int, app *terra.TerraApp, height uint64, denom string) (err error) {
	ctx := app.NewContext(true, tmproto.Header{})

	for address, amount := range changes {
		accAddress, _ := sdk.AccAddressFromBech32(address) // a ranker have only one address

		// skip module accounts
		account := app.AccountKeeper.GetAccount(ctx, accAddress)
		_, isModule := account.(*authtypes.ModuleAccount)
		if isModule {
			continue
		}

		amountPrev := app.BankKeeper.GetBalance(ctx, accAddress, denom)
		amountAfter := sdk.NewCoin(denom, amountPrev.Amount.Add(amount))

		ranker := Ranker{Addresses: []string{address}}
		// remove outdated rank
		ranker.Score = amountPrev
		err = list.Unrank(ranker)
		if err != nil {
			return
		}

		// apply new rank
		ranker.Score = amountAfter
		err = list.Rank(ranker)
		if err != nil {
			return
		}
	}

	return err
}

// internal structure to marshal richlist
type jsonlist struct {
	Height  uint64   `json:"height,omitempty"`
	Rankers []Ranker `json:"table"`
}

// btree have no MarshalJSON() to we need to implement it
// no need to unmarshal
func (list Richlist) MarshalJSON() (b []byte, err error) {
	l := jsonlist{Height: list.Height}
	n := 0

	list.Rankers.Descend(func(i btree.Item) bool {
		n++
		rank := i.(Ranker)
		rank.Position = n
		l.Rankers = append(l.Rankers, rank)
		return true
	})

	return json.Marshal(l)
}
