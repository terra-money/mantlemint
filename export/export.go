package export

import (
	"fmt"
	"os"
	"strings"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	distrotypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	terra "github.com/terra-money/core/v2/app"
)

var IsAccountExportRunning = false

func ExportAllAccounts(app *terra.TerraApp) error {
	if IsAccountExportRunning {
		return fmt.Errorf("exporting is still running")
	}
	IsAccountExportRunning = true
	go exportWorker(app)
	return nil
}

func ExportCirculatingSupply(app *terra.TerraApp) (sdktypes.Int, error) {
	height := app.LastBlockHeight()
	ctx := app.NewContext(true, tmproto.Header{Height: height})
	time := time.Now()
	totalVesting := sdktypes.NewInt(0)
	app.AccountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		switch account.(type) {
		case *vestingtypes.PeriodicVestingAccount:
			v := account.(*vestingtypes.PeriodicVestingAccount)
			totalVesting = totalVesting.Add(v.GetVestingCoins(time).AmountOf("uluna"))
		default:
			return false
		}
		return false
	})
	totalSupply, err := app.BankKeeper.TotalSupply(sdktypes.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	if err != nil {
		return sdktypes.Int{}, err
	}
	lunaTotalSupply := totalSupply.Supply.AmountOf("uluna")
	communityPool, err := app.DistrKeeper.CommunityPool(sdktypes.WrapSDKContext(ctx), &distrotypes.QueryCommunityPoolRequest{})
	if err != nil {
		return sdktypes.Int{}, err
	}
	lunaCommunityPool := communityPool.Pool.AmountOf("uluna").TruncateInt()
	feePool := app.DistrKeeper.GetFeePool(ctx)
	lunaFeePool := feePool.CommunityPool.AmountOf("uluna").TruncateInt()

	govAccount := app.GovKeeper.GetGovernanceAccount(ctx)
	lunaGovAccount := app.BankKeeper.GetBalance(ctx, govAccount.GetAddress(), "uluna").Amount

	return lunaTotalSupply.Sub(lunaCommunityPool).Sub(totalVesting).Sub(lunaFeePool).Sub(lunaGovAccount), nil
}

func exportWorker(app *terra.TerraApp) {
	app.Logger().Info("[export] exporting accounts")
	height := app.LastBlockHeight()
	ctx := app.NewContext(true, tmproto.Header{Height: height})
	// Should use lastest block time
	time := time.Now()
	var accounts []string
	count := 0
	app.AccountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		balance := app.BankKeeper.GetBalance(ctx, account.GetAddress(), "uluna").Amount
		delegationRewards, err := app.DistrKeeper.DelegationTotalRewards(sdktypes.WrapSDKContext(ctx), &types.QueryDelegationTotalRewardsRequest{
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
	IsAccountExportRunning = false
	app.Logger().Info("[export] exporting accounts completed")
}
