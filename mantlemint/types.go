package mantlemint

import (
	"github.com/cometbft/cometbft/state"
	cometbfttypes "github.com/cometbft/cometbft/types"
)

type Mantlemint interface {
	Inject(*cometbfttypes.Block) error
	Init(*cometbfttypes.GenesisDoc) error
	LoadInitialState() error
	GetCurrentHeight() int64
	GetCurrentBlock() *cometbfttypes.Block
	GetCurrentState() state.State
	GetCurrentEventCollector() *EventCollector
	SetBlockExecutor(executor Executor)
}

type Executor interface {
	ApplyBlock(state.State, cometbfttypes.BlockID, *cometbfttypes.Block) (state.State, int64, error)
	SetEventBus(publisher cometbfttypes.BlockEventPublisher)
}

type MantlemintCallbackBefore func(block *cometbfttypes.Block) error
type MantlemintCallbackAfter func(block *cometbfttypes.Block, events *EventCollector) error

// --- internal types
