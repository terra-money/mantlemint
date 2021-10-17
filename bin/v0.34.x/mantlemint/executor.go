package mantlemint

import (
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool/mock"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/state"
	tmdb "github.com/tendermint/tm-db"
	"os"
)

// NewMantlemintExecutor creates stock tendermint block executor, with stubbed mempool and evidence pool
func NewMantlemintExecutor(
	db tmdb.DB,
	conn proxy.AppConnConsensus,
) *state.BlockExecutor {
	return state.NewBlockExecutor(
		state.NewStore(db),

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
