package export

import (
	"fmt"
	"os"
	"strings"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrikeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	distrotypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	app "github.com/terra-money/core/v2/app"
)

var IsAccountExportWorkerRunning = false

func ExportAllAccounts(app *app.TerraApp) error {
	if IsAccountExportWorkerRunning {
		return fmt.Errorf("exporting is still running")
	}
	IsAccountExportWorkerRunning = true
	go runAccountExportWorker(app)
	return nil
}

func ExportCirculatingSupply(app *app.TerraApp) (sdktypes.Int, error) {
	height := app.LastBlockHeight()
	ctx := app.NewContext(true, tmproto.Header{Height: height})
	time := time.Now()
	totalVesting := sdktypes.NewInt(0)
	distQuerier := distrikeeper.NewQuerier(app.DistrKeeper)
	app.AccountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		switch account.(type) {
		case *vestingtypes.PeriodicVestingAccount:
			v := account.(*vestingtypes.PeriodicVestingAccount)
			totalVesting = totalVesting.Add(v.GetVestingCoins(time).AmountOf("uluna"))
		case *vestingtypes.ContinuousVestingAccount:
			v := account.(*vestingtypes.ContinuousVestingAccount)
			totalVesting = totalVesting.Add(v.GetVestingCoins(time).AmountOf("uluna"))
		case *vestingtypes.DelayedVestingAccount:
			v := account.(*vestingtypes.DelayedVestingAccount)
			totalVesting = totalVesting.Add(v.GetVestingCoins(time).AmountOf("uluna"))
		case *vestingtypes.PermanentLockedAccount:
			v := account.(*vestingtypes.PermanentLockedAccount)
			totalVesting = totalVesting.Add(v.GetVestingCoins(time).AmountOf("uluna"))
		default:
			return false
		}
		return false
	})
	totalSupply, err := app.BankKeeper.SupplyOf(sdktypes.WrapSDKContext(ctx), &banktypes.QuerySupplyOfRequest{
		Denom: "uluna",
	})
	if err != nil {
		return sdktypes.Int{}, err
	}
	lunaTotalSupply := totalSupply.Amount.Amount
	communityPool, err := distQuerier.CommunityPool(sdktypes.WrapSDKContext(ctx), &distrotypes.QueryCommunityPoolRequest{})
	if err != nil {
		return sdktypes.Int{}, err
	}
	lunaCommunityPool := communityPool.Pool.AmountOf("uluna").TruncateInt()

	return lunaTotalSupply.Sub(lunaCommunityPool).Sub(totalVesting), nil
}

func runAccountExportWorker(app *app.TerraApp) {
	app.Logger().Info("[export] exporting accounts")
	height := app.LastBlockHeight()
	ctx := app.NewContext(true, tmproto.Header{Height: height})
	// Should use lastest block time
	time := time.Now()
	var accounts []string
	count := 0
	distQuerier := distrikeeper.NewQuerier(app.DistrKeeper)
	app.AccountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		balance := app.BankKeeper.GetBalance(ctx, account.GetAddress(), "uluna").Amount
		delegationRewards, err := distQuerier.DelegationTotalRewards(sdktypes.WrapSDKContext(ctx), &types.QueryDelegationTotalRewardsRequest{
			DelegatorAddress: account.GetAddress().String(),
		})
		if err != nil {
			panic(err)
		} else {
			balance = balance.Add(delegationRewards.Total.AmountOf("uluna").TruncateInt())
		}
		switch account.(type) {
		case *vestingtypes.PeriodicVestingAccount:
			v := account.(*vestingtypes.PeriodicVestingAccount)
			vesting := v.GetVestingCoins(time).AmountOf("uluna")
			vested := balance.Add(v.DelegatedFree.AmountOf("uluna"))
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", v.Address, vested, vesting))
		case *vestingtypes.ContinuousVestingAccount:
			v := account.(*vestingtypes.ContinuousVestingAccount)
			vesting := v.GetVestingCoins(time).AmountOf("uluna")
			vested := balance.Add(v.DelegatedFree.AmountOf("uluna"))
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", v.Address, vested, vesting))
		case *vestingtypes.DelayedVestingAccount:
			v := account.(*vestingtypes.DelayedVestingAccount)
			vesting := v.GetVestingCoins(time).AmountOf("uluna")
			vested := balance.Add(v.DelegatedFree.AmountOf("uluna"))
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", v.Address, vested, vesting))
		case *vestingtypes.PermanentLockedAccount:
			v := account.(*vestingtypes.PermanentLockedAccount)
			vesting := v.GetVestingCoins(time).AmountOf("uluna")
			vested := balance.Add(v.DelegatedFree.AmountOf("uluna"))
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", v.Address, vested, vesting))
		case *authtypes.BaseAccount:
			delegations := app.StakingKeeper.GetAllDelegatorDelegations(ctx, account.GetAddress())
			for _, d := range delegations {
				val, ok := app.StakingKeeper.GetValidator(ctx, d.GetValidatorAddr())
				if ok {
					balance = balance.Add(val.TokensFromShares(d.GetShares()).TruncateInt())
				}
			}

			unbonding := app.StakingKeeper.GetAllUnbondingDelegations(ctx, account.GetAddress())
			for _, ub := range unbonding {
				for _, e := range ub.Entries {
					balance = balance.Add(e.Balance)
				}
			}
			vesting := "0"
			vested := balance
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", account.GetAddress(), vested, vesting))
		default:
			return false
		}
		count += 1
		if count%20000 == 0 {
			app.Logger().Info(fmt.Sprintf("[export] getting balance count: %d", count))
		}
		return false
	})
	os.WriteFile("accounts.csv", []byte(strings.Join(accounts, "\n")), 0700)
	IsAccountExportWorkerRunning = false
	app.Logger().Info("[export] exporting accounts completed")
}
