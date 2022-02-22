package mantlemint

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/state"
	tendermint "github.com/tendermint/tendermint/types"
	"github.com/terra-money/mantlemint-provider-v0.34.x/mantlemint/event"
)

type Mantlemint interface {
	Inject(*tendermint.Block) error
	Init(*tendermint.GenesisDoc) error
	LoadInitialState() error
	GetCurrentHeight() int64
	GetCurrentBlock() *tendermint.Block
	GetCurrentState() state.State
	GetCurrentEventCollector() *event.EventCollector
	SetBlockExecutor(executor Executor)
}

type Executor interface {
	ApplyBlock(state.State, tendermint.BlockID, *tendermint.Block) (state.State, int64, error)
	SetEventBus(publisher tendermint.BlockEventPublisher)
}

type RouteRegisterer func(router *mux.Router)
type RouteRegistererDual func(router *mux.Router, router2 *runtime.ServeMux)

type MantlemintCallbackBefore func(block *tendermint.Block) error
type MantlemintCallbackAfter func(block *tendermint.Block, events *event.EventCollector) error

// --- internal types
type BroadcastTxCommitHandler func(context.Context, tendermint.Tx) (*coretypes.ResultBroadcastTxCommit, error)
type BroadcastTxAsyncHandler func(context.Context, tendermint.Tx) (*coretypes.ResultBroadcastTx, error)
type BroadcastTxSyncHandler func(context.Context, tendermint.Tx) (*coretypes.ResultBroadcastTx, error)
