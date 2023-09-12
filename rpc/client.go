package rpc

import (
	"context"
	"fmt"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	tmlog "github.com/cometbft/cometbft/libs/log"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cometbft/cometbft/rpc/core"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cometbfttypes "github.com/cometbft/cometbft/types"
)

var _ rpcclient.Client = (*MantlemintRPCClient)(nil)

type MantlemintRPCClient struct {
	client abcicli.Client
}

func NewRpcClient(client abcicli.Client) rpcclient.Client {
	return &MantlemintRPCClient{client: client}
}

func (m *MantlemintRPCClient) ABCIInfo(ctx context.Context) (*coretypes.ResultABCIInfo, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
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

func (m *MantlemintRPCClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
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

func (m *MantlemintRPCClient) Start() error {
	panic("implement me")
}

func (m *MantlemintRPCClient) OnStart() error {
	panic("implement me")
}

func (m *MantlemintRPCClient) Stop() error {
	panic("implement me")
}

func (m *MantlemintRPCClient) OnStop() {
	panic("implement me")
}

func (m *MantlemintRPCClient) Reset() error {
	panic("implement me")
}

func (m *MantlemintRPCClient) OnReset() error {
	panic("implement me")
}

func (m *MantlemintRPCClient) IsRunning() bool {
	return m.client.IsRunning()
}

func (m *MantlemintRPCClient) Quit() <-chan struct{} {
	panic("implement me")
}

func (m *MantlemintRPCClient) String() string {
	return m.client.String()
}

func (m *MantlemintRPCClient) SetLogger(logger tmlog.Logger) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BroadcastTxCommit(ctx context.Context, tx cometbfttypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BroadcastTxAsync(ctx context.Context, tx cometbfttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BroadcastTxSync(ctx context.Context, tx cometbfttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Subscribe(ctx context.Context, subscriber, query string, outCapacity ...int) (out <-chan coretypes.ResultEvent, err error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Unsubscribe(ctx context.Context, subscriber, query string) error {
	panic("implement me")
}

func (m *MantlemintRPCClient) UnsubscribeAll(ctx context.Context, subscriber string) error {
	panic("implement me")
}

func (m *MantlemintRPCClient) Genesis(ctx context.Context) (*coretypes.ResultGenesis, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) GenesisChunked(ctx context.Context, u uint) (*coretypes.ResultGenesisChunk, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*coretypes.ResultBlockchainInfo, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) NetInfo(ctx context.Context) (*coretypes.ResultNetInfo, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) DumpConsensusState(ctx context.Context) (*coretypes.ResultDumpConsensusState, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) ConsensusState(ctx context.Context) (*coretypes.ResultConsensusState, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) ConsensusParams(ctx context.Context, height *int64) (*coretypes.ResultConsensusParams, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Health(ctx context.Context) (*coretypes.ResultHealth, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	return core.Block(nil, height)
	fmt.Println("is it you")

	panic("implement me")
}

func (m *MantlemintRPCClient) BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Commit(ctx context.Context, height *int64) (*coretypes.ResultCommit, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*coretypes.ResultValidators, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) TxSearch(ctx context.Context, query string, prove bool, page, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BlockSearch(ctx context.Context, query string, page, perPage *int, orderBy string) (*coretypes.ResultBlockSearch, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Status(ctx context.Context) (*coretypes.ResultStatus, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) BroadcastEvidence(ctx context.Context, evidence cometbfttypes.Evidence) (*coretypes.ResultBroadcastEvidence, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) UnconfirmedTxs(ctx context.Context, limit *int) (*coretypes.ResultUnconfirmedTxs, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) NumUnconfirmedTxs(ctx context.Context) (*coretypes.ResultUnconfirmedTxs, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) CheckTx(ctx context.Context, tx cometbfttypes.Tx) (*coretypes.ResultCheckTx, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) Header(ctx context.Context, height *int64) (*coretypes.ResultHeader, error) {
	panic("implement me")
}

func (m *MantlemintRPCClient) HeaderByHash(ctx context.Context, hash bytes.HexBytes) (*coretypes.ResultHeader, error) {
	panic("implement me")
}
