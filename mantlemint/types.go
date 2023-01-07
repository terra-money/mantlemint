package mantlemint

import (
	"github.com/tendermint/tendermint/state"
	tendermint "github.com/tendermint/tendermint/types"
)

type Mantlemint interface {
	Inject(*tendermint.Block) error
	Init(*tendermint.GenesisDoc) error
	LoadInitialState() error
	GetCurrentHeight() int64
	GetCurrentBlock() *tendermint.Block
	GetCurrentState() state.State
	GetCurrentEventCollector() *EventCollector
	SetBlockExecutor(executor Executor)
}

type Executor interface {
	ApplyBlock(state.State, tendermint.BlockID, *tendermint.Block) (state.State, int64, error)
	SetEventBus(publisher tendermint.BlockEventPublisher)
}

//nolint:revive
type (
	MantlemintCallbackBefore func(block *tendermint.Block) error
	MantlemintCallbackAfter  func(block *tendermint.Block, events *EventCollector) error
)

// --- internal types
