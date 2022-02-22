package mantlemint

import (
	"context"
	"fmt"
	abcicli "github.com/tendermint/tendermint/abci/client"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tendermint "github.com/tendermint/tendermint/types"
)

var _ rpcclient.Client = (*RPCClient)(nil)

type RPCClient struct {
	client                   abcicli.Client
	broadcastTxCommitHandler BroadcastTxCommitHandler
	broadcastTxSyncHandler   BroadcastTxSyncHandler
	broadcastTxAsyncHandler  BroadcastTxAsyncHandler
}

func NewRPCClient(
	client abcicli.Client,
	broadcastTxCommitHandler BroadcastTxCommitHandler,
	broadcastTxSyncHandler BroadcastTxSyncHandler,
	broadcastTxAsyncHandler BroadcastTxAsyncHandler,
) rpcclient.Client {
	return &RPCClient{
		client:                   client,
		broadcastTxCommitHandler: broadcastTxCommitHandler,
		broadcastTxSyncHandler:   broadcastTxSyncHandler,
		broadcastTxAsyncHandler:  broadcastTxAsyncHandler,
	}
}

func (m *RPCClient) ABCIInfo(ctx context.Context) (*coretypes.ResultABCIInfo, error) {
	panic("implement me")
}

func (m *RPCClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
	if resp, err := m.client.QuerySync(abci.RequestQuery{
		Data:   data,
		Path:   path,
		Height: 0,
		Prove:  false,
	}); err != nil {
		return nil, err
	} else {
		return &coretypes.ResultABCIQuery{
			Response: *resp,
		}, nil
	}
}

func (m *RPCClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
	if resp, err := m.client.QuerySync(abci.RequestQuery{
		Data:   data,
		Path:   path,
		Height: opts.Height,
		Prove:  opts.Prove,
	}); err != nil {
		return nil, err
	} else {
		return &coretypes.ResultABCIQuery{
			Response: *resp,
		}, nil
	}
}

func (m *RPCClient) BroadcastTxCommit(ctx context.Context, tx tendermint.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	if invariant(m.broadcastTxCommitHandler, "unset broadcastTxCommitHandler") {
		return m.broadcastTxCommitHandler(ctx, tx)
	} else {
		return nil, fmt.Errorf("invalid access")
	}
}

func (m *RPCClient) BroadcastTxAsync(ctx context.Context, tx tendermint.Tx) (*coretypes.ResultBroadcastTx, error) {
	if invariant(m.broadcastTxAsyncHandler, "invalid broadcastTxAsync") {
		return m.broadcastTxAsyncHandler(ctx, tx)
	} else {
		return nil, fmt.Errorf("invalid access")
	}
}

func (m *RPCClient) BroadcastTxSync(ctx context.Context, tx tendermint.Tx) (*coretypes.ResultBroadcastTx, error) {
	if invariant(m.broadcastTxSyncHandler, "invalid broadcastTxSync") {
		return m.broadcastTxSyncHandler(ctx, tx)
	} else {
		return nil, fmt.Errorf("invalid access")
	}
}

func (m *RPCClient) Start() error {
	panic("implement me")
}

func (m *RPCClient) OnStart() error {
	panic("implement me")
}

func (m *RPCClient) Stop() error {
	panic("implement me")
}

func (m *RPCClient) OnStop() {
	panic("implement me")
}

func (m *RPCClient) Reset() error {
	panic("implement me")
}

func (m *RPCClient) OnReset() error {
	panic("implement me")
}

func (m *RPCClient) IsRunning() bool {
	return m.client.IsRunning()
}

func (m *RPCClient) Quit() <-chan struct{} {
	panic("implement me")
}

func (m *RPCClient) String() string {
	return m.client.String()
}

func (m *RPCClient) SetLogger(logger tmlog.Logger) {
	panic("implement me")
}

func (m *RPCClient) Subscribe(ctx context.Context, subscriber, query string, outCapacity ...int) (out <-chan coretypes.ResultEvent, err error) {
	panic("implement me")
}

func (m *RPCClient) Unsubscribe(ctx context.Context, subscriber, query string) error {
	panic("implement me")
}

func (m *RPCClient) UnsubscribeAll(ctx context.Context, subscriber string) error {
	panic("implement me")
}

func (m *RPCClient) Genesis(ctx context.Context) (*coretypes.ResultGenesis, error) {
	panic("implement me")
}

func (m *RPCClient) GenesisChunked(ctx context.Context, u uint) (*coretypes.ResultGenesisChunk, error) {
	panic("implement me")
}

func (m *RPCClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*coretypes.ResultBlockchainInfo, error) {
	panic("implement me")
}

func (m *RPCClient) NetInfo(ctx context.Context) (*coretypes.ResultNetInfo, error) {
	panic("implement me")
}

func (m *RPCClient) DumpConsensusState(ctx context.Context) (*coretypes.ResultDumpConsensusState, error) {
	panic("implement me")
}

func (m *RPCClient) ConsensusState(ctx context.Context) (*coretypes.ResultConsensusState, error) {
	panic("implement me")
}

func (m *RPCClient) ConsensusParams(ctx context.Context, height *int64) (*coretypes.ResultConsensusParams, error) {
	panic("implement me")
}

func (m *RPCClient) Health(ctx context.Context) (*coretypes.ResultHealth, error) {
	panic("implement me")
}

func (m *RPCClient) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	return core.Block(nil, height)
}

func (m *RPCClient) BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error) {
	panic("implement me")
}

func (m *RPCClient) BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	panic("implement me")
}

func (m *RPCClient) Commit(ctx context.Context, height *int64) (*coretypes.ResultCommit, error) {
	panic("implement me")
}

func (m *RPCClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*coretypes.ResultValidators, error) {
	panic("implement me")
}

func (m *RPCClient) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	panic("implement me")
}

func (m *RPCClient) TxSearch(ctx context.Context, query string, prove bool, page, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	panic("implement me")
}

func (m *RPCClient) BlockSearch(ctx context.Context, query string, page, perPage *int, orderBy string) (*coretypes.ResultBlockSearch, error) {
	panic("implement me")
}

func (m *RPCClient) Status(ctx context.Context) (*coretypes.ResultStatus, error) {
	panic("implement me")
}

func (m *RPCClient) BroadcastEvidence(ctx context.Context, evidence tendermint.Evidence) (*coretypes.ResultBroadcastEvidence, error) {
	panic("implement me")
}

func (m *RPCClient) UnconfirmedTxs(ctx context.Context, limit *int) (*coretypes.ResultUnconfirmedTxs, error) {
	panic("implement me")
}

func (m *RPCClient) NumUnconfirmedTxs(ctx context.Context) (*coretypes.ResultUnconfirmedTxs, error) {
	panic("implement me")
}

func (m *RPCClient) CheckTx(ctx context.Context, tx tendermint.Tx) (*coretypes.ResultCheckTx, error) {
	panic("implement me")
}

func invariant(handler interface{}, err string) bool {
	if handler != nil {
		return true
	} else {
		panic(err)
	}
}
