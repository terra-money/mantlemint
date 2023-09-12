package block_feed

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var _ BlockFeed = (*RPCSubscription)(nil)

type RPCSubscription struct {
	rpcEndpoints []string
	cSub         chan *BlockResult
}

func NewRpcSubscription(rpcEndpoints []string) (*RPCSubscription, error) {
	return &RPCSubscription{
		rpcEndpoints: rpcEndpoints,
		cSub:         make(chan *BlockResult),
	}, nil
}

func (rpc *RPCSubscription) SyncFromUntil(from int64, to int64, rpcIndex int) {
	cSub := rpc.cSub

	log.Printf("[block_feed/rpc] subscription started, from=%d, to=%d\n", from, to)

	// is a blocking operation
	for i := from; i <= to; i++ {
		log.Printf("[block_feed/rpc] receiving block %d...\n", i)
		url := fmt.Sprintf("%s/block?height=%d", rpc.rpcEndpoints[rpcIndex], i)
		res, err := http.Get(url)
		if err != nil {
			log.Fatalf("block request failed, %v", err)
		}

		resBytes, err := io.ReadAll(res.Body)
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

func (rpc *RPCSubscription) Subscribe(_ int) (chan *BlockResult, error) {
	return rpc.cSub, nil
}

func (rpc *RPCSubscription) Close() error {
	return nil
}
