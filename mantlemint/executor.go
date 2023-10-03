package mantlemint

import (
	"os"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool/mock"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/state"
	"github.com/terra-money/mantlemint/db/wrapped"

	dbm "github.com/tendermint/tm-db"
)

// NewMantlemintExecutor creates stock tendermint block executor, with stubbed mempool and evidence pool
func NewMantlemintExecutor(
	db dbm.DB,
	conn proxy.AppConnConsensus,
) *state.BlockExecutor {
	return state.NewBlockExecutor(
		state.NewStore(wrapped.NewWrappedDB(db), state.StoreOptions{
			DiscardABCIResponses: false,
		}),

		// discard all tm logging
		log.NewTMLogger(os.Stdout),

		// use app connection as provided
		conn,

		// no mempool, as mantlemint doesn't handle tx broadcasts
		mock.Mempool{},

		// no evidence pool, as mantlemint only receives evidence from other peers
		state.EmptyEvidencePool{},
	)
}
