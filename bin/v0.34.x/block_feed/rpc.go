package block_feed

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var _ BlockFeed = (*RPCSubscription)(nil)

type RPCSubscription struct {
	rpcEndpoints []string
	cSub         chan *BlockResult
	connIdx      int
}

func (rpc *RPCSubscription) GetBlockFeedChannel() chan *BlockResult {
	return rpc.cSub
}

func (rpc *RPCSubscription) IsSynced() bool {
	//TODO implement me
	panic("implement me")
}

func (rpc *RPCSubscription) Inject(result *BlockResult) {
	rpc.cSub <- result
}

func NewRpcSubscription(rpcEndpoints []string) (*RPCSubscription, error) {
	return &RPCSubscription{
		rpcEndpoints: rpcEndpoints,
		cSub:         make(chan *BlockResult),
		connIdx:      0,
	}, nil
}

func (rpc *RPCSubscription) SyncFromUntil(from int64, to int64) {
	if len(rpc.rpcEndpoints) == 0 {
		return
	}

	var cSub = rpc.cSub
	rpcIndex := rpc.ConnIndex()

	log.Printf("[block_feed/rpc] subscription started, from=%d, to=%d\n", from, to)

	// is a blocking operation
	for i := from; i <= to; i++ {
		log.Printf("[block_feed/rpc] receiving block %d...\n", i)
		url := fmt.Sprintf("%s/block?height=%d", rpc.rpcEndpoints[rpcIndex], i)
		res, err := http.Get(url)
		if err != nil {
			log.Fatalf("block request failed, %v", err)
		}

		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("block request failed, %v", err)
		}

		if block, blockParseErr := ExtractBlockFromRPCResponse(resBytes); blockParseErr != nil {
			log.Fatalf("block parse failed, %v", err)
		} else {
			cSub <- block
		}
	}

	cSub <- nil
}

func (rpc *RPCSubscription) Next() {
	rpc.connIdx = rpc.connIdx + 1%len(rpc.rpcEndpoints)
}

func (rpc *RPCSubscription) ConnIndex() int {
	return rpc.connIdx
}

func (rpc *RPCSubscription) Subscribe() (chan *BlockResult, error) {
	return rpc.cSub, nil
}

func (rpc *RPCSubscription) Close() error {
	return nil
}
