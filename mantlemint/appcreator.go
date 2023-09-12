package mantlemint

import (
	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/types"
	tmsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/proxy"
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
