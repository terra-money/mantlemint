package mantlemint

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/service"
	tmsync "github.com/tendermint/tendermint/libs/sync"
)

var _ abcicli.Client = (*ConcurrentQueryClient)(nil)

type ConcurrentQueryClient struct {
	*localClient
}

func NewConcurrentQueryClient(mtx *tmsync.RWMutex, app types.Application) abcicli.Client {
	if mtx == nil {
		mtx = &tmsync.RWMutex{}
	}

	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}

	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)

	return &ConcurrentQueryClient{
		localClient: cli,
	}
}

func (uc *ConcurrentQueryClient) QueryAsync(req types.RequestQuery) *abcicli.ReqRes {
	res := uc.Application.Query(req)
	return uc.callback(
		types.ToRequestQuery(req),
		types.ToResponseQuery(res),
	)
}

func (uc *ConcurrentQueryClient) QuerySync(req types.RequestQuery) (*types.ResponseQuery, error) {
	res := uc.Application.Query(req)
	return &res, nil
}
