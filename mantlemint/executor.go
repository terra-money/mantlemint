package mantlemint

import (
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/state"
	"os"
)

// NewMantlemintExecutor creates stock cometbft block executor, with stubbed mempool and evidence pool
func NewMantlemintExecutor(
	db tmdb.DB,
	conn proxy.AppConnConsensus,
) *state.BlockExecutor {
	return state.NewBlockExecutor(
		state.NewStore(db, state.StoreOptions{
			DiscardABCIResponses: false,
		}),

		// discard all tm logging
		log.NewTMLogger(os.Stdout),

		// use app connection as provided
		conn,

		// no mempool, as mantlemint doesn't handle tx broadcasts
		Mempool{},

		// no evidence pool, as mantlemint only receives evidence from other peers
		state.EmptyEvidencePool{},
	)
}
