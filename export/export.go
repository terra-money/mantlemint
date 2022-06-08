package export

import (
	"fmt"
	"os"
	"strings"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
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

func exportWorker(app *terra.TerraApp) {
	height := app.LastBlockHeight()
	ctx := app.NewContext(true, tmproto.Header{Height: height})
	// Should use lastest block time
	time := time.Now()
	var accounts []string
	count := 0
	app.AccountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		balance := app.BankKeeper.GetBalance(ctx, account.GetAddress(), "uluna").Amount
		switch account.(type) {
		case *vestingtypes.PeriodicVestingAccount:
			v := account.(*vestingtypes.PeriodicVestingAccount)
			vesting := v.GetVestingCoins(time).AmountOf("uluna")
			vested := balance
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", v.Address, vested, vesting))
		case *authtypes.BaseAccount:
			vesting := "0"
			vested := balance
			accounts = append(accounts, fmt.Sprintf("%s,%s,%s", account.GetAddress(), vested, vesting))
		default:
			return false
		}
		count += 1
		if count%100000 == 0 {
			app.Logger().Info(fmt.Sprintf("Getting balance count: %d", count))
		}
		return false
	})
	os.WriteFile("accounts.csv", []byte(strings.Join(accounts, "\n")), 0700)
	IsAccountExportRunning = false
}
