package mantlemint

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Pass this in as an option to use a dbStoreAdapter instead of an IAVLStore for simulation speed.
func decorateFauxMerkleMode() func(app *baseapp.BaseApp) {
	return func(app *baseapp.BaseApp) {
		app.SetFauxMerkleMode()
	}
}

func decorateCMS(cms types.CommitMultiStore) func(app *baseapp.BaseApp) {
	return func(app *baseapp.BaseApp) {
		app.SetCMS(cms)
	}
}
