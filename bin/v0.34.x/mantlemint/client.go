package mantlemint

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/types"
	tmsync "github.com/tendermint/tendermint/libs/sync"
)

var _ abcicli.Client = (*UnmutexedClient)(nil)

type UnmutexedClient struct {
	abcicli.Client
	mtx *tmsync.RWMutex
	app types.Application
	abcicli.Callback
}

func NewConcurrentQueryClient(mtx *tmsync.RWMutex, app types.Application) abcicli.Client {
	return &UnmutexedClient{
		mtx: mtx,
		app: app,
	}
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

//-------------------------------------------------------
func (uc *UnmutexedClient) callback(req *types.Request, res *types.Response) *abcicli.ReqRes {
	uc.Callback(req, res)
	return newLocalReqRes(req, res)
}

func newLocalReqRes(req *types.Request, res *types.Response) *abcicli.ReqRes {
	reqRes := abcicli.NewReqRes(req)
	reqRes.Response = res
	reqRes.SetDone()
	return reqRes
}
