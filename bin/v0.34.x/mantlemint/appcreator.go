package mantlemint

import (
	abcicli "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/types"
	tmsync "github.com/tendermint/tendermint/libs/sync"
	"github.com/tendermint/tendermint/proxy"
)

type localClientCreator struct {
	mtx *tmsync.RWMutex
	app types.Application
}

func NewConcurrentQueryClientCreator(app types.Application) proxy.ClientCreator {
	return &localClientCreator{
		mtx: new(tmsync.RWMutex),
		app: app,
	}
}

func (l *localClientCreator) NewABCIClient() (abcicli.Client, error) {
	return NewConcurrentQueryClient(l.mtx, l.app), nil
}