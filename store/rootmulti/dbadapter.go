package rootmulti

import (
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/types"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	dbm "github.com/tendermint/tm-db"
)

var commithash = []byte("FAKE_HASH")

//----------------------------------------
// commitDBStoreWrapper should only be used for simulation/debugging,
// as it doesn't compute any commit hash, and it cannot load older state.

// Wrapper type for dbm.Db with implementation of KVStore
type commitDBStoreAdapter struct {
	dbadapter.Store
	prefix []byte
}

func (cdsa commitDBStoreAdapter) Commit() types.CommitID {
	return types.CommitID{
		Version: -1,
		Hash:    commithash,
	}
}

func (cdsa commitDBStoreAdapter) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: -1,
		Hash:    commithash,
	}
}
func (cdsa commitDBStoreAdapter) SetPruning(_ pruningtypes.PruningOptions) {}

// GetPruning is a no-op as pruning options cannot be directly set on this store.
// They must be set on the root commit multi-store.
func (cdsa commitDBStoreAdapter) GetPruning() pruningtypes.PruningOptions {
	return pruningtypes.NewPruningOptions(pruningtypes.PruningUndefined)
}

func (cdsa *commitDBStoreAdapter) BranchStoreWithHeightLimitedDB(hldb dbm.DB) types.CommitKVStore {
	db := dbm.NewPrefixDB(hldb, cdsa.prefix)

	return commitDBStoreAdapter{Store: dbadapter.Store{DB: db}, prefix: cdsa.prefix}
}
