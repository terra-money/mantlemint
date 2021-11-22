package mantlemint

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/service"
	tmsync "github.com/tendermint/tendermint/libs/sync"
)

var _ abcicli.Client = (*UnmutexedClient)(nil)

type UnmutexedClient struct {
	service.BaseService
	localClient

	mtx *tmsync.RWMutex
	app types.Application
	abcicli.Callback
}

func NewConcurrentQueryClient(mtx *tmsync.RWMutex, app types.Application) abcicli.Client {
	cli := &UnmutexedClient{
		mtx: mtx,
		app: app,
	}

	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

func (uc *UnmutexedClient) QueryAsync(req types.RequestQuery) *abcicli.ReqRes {
	res := uc.app.Query(req)
	return uc.callback(
		types.ToRequestQuery(req),
		types.ToResponseQuery(res),
	)
}

func (uc *UnmutexedClient) QuerySync(req types.RequestQuery) (*types.ResponseQuery, error) {
	res := uc.app.Query(req)
	return &res, nil
}
