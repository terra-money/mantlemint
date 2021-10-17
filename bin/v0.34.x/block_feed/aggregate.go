package block_feed

import (
	"fmt"
	"log"
	"time"
)

var _ BlockFeed = (*AggregateSubscription)(nil)

type AggregateSubscription struct {
	lastKnownBlock            int64
	ws                        *WSSubscription
	wsEndpoints               []string
	rpc                       *RPCSubscription
	aggregateBlockChannel     chan *BlockResult
	lastKnownWSEndpointsIndex int
	isSynced                  bool
}

var done *BlockResult = nil

func NewAggregateBlockFeed(
	currentBlock int64,
	rpcEndpoints []string,
	wsEndpoints []string,
) *AggregateSubscription {
	var rpc, rpcErr = NewRpcSubscription(rpcEndpoints)
	if rpcErr != nil {
		panic(rpcErr)
	}

	// ws starts with 1st occurrence of ws endpoints
	var ws, wsErr = NewWSSubscription(wsEndpoints)
	if wsErr != nil {
		panic(wsErr)
	}

	return &AggregateSubscription{
		ws:                        ws,
		rpc:                       rpc,
		lastKnownBlock:            currentBlock,
		lastKnownWSEndpointsIndex: 0,
	}
}

func (ags *AggregateSubscription) Subscribe(rpcIndex int) (chan *BlockResult, error) {
	// create rpc subscriber
	cRpc, cRpcErr := ags.rpc.Subscribe(rpcIndex)
	if cRpcErr != nil {
		return nil, cRpcErr
	}

	// create websocket subscriber
	cWS, cWSErr := ags.ws.Subscribe(rpcIndex)
	if cWSErr != nil {
		return nil, cWSErr
	}

	// start with isSynced flag false
	ags.isSynced = false

	// check if the first block received from ws is the right block (currentHeight + 1)
	// if not, the local blockchain is behind, in such case we would need to sync from Rpc.
	if firstBlock := <-cWS; firstBlock.Block.Header.Height != ags.lastKnownBlock+1 {
		log.Printf("[block_feed/aggregate] received the first block(%d), but local blockchain is at (%d)\n", firstBlock.Block.Header.Height, ags.lastKnownBlock)
		go func() {
			go ags.rpc.SyncFromUntil(ags.lastKnownBlock+1, firstBlock.Block.Header.Height, rpcIndex)
			for {
				r := <-cRpc
				if r != done {
					ags.aggregateBlockChannel <- r
				} else {
					break
				}
			}

			// patch ws to aggregate
			for {
				r := <-cWS

				// gracefully handle done signal; in whatever case received is nil,
				// handle reconnection here
				if r == done {
					log.Printf("[block_feed/aggregate] websocket done signal received, reconnecting...")
					ags.Reconnect()
					break
				} else {
					// if block feeder got upto this point,
					// it is relatively safe that mantle is synced
					ags.isSynced = true
					ags.aggregateBlockChannel <- <-cWS
				}
			}
		}()
	}

	return ags.aggregateBlockChannel, nil
}

func (ags *AggregateSubscription) Close() error {
	rpcCloseErr := ags.rpc.Close()
	wsCloseErr := ags.ws.Close()

	return fmt.Errorf("error during aggregate subscription close: %s, %s", rpcCloseErr, wsCloseErr)
}

// Reconnect reestablishes all underlying connections
// On any reconnection, it is likely that the underlying RPC is having some problem.
// To mitigate this,
func (ags *AggregateSubscription) Reconnect() {
	ags.isSynced = false
	ags.lastKnownWSEndpointsIndex++
	ags.lastKnownWSEndpointsIndex = ags.lastKnownWSEndpointsIndex % len(ags.wsEndpoints)
	time.Sleep(time.Second)

	log.Printf("[block_feed/aggregate] reconnecting with rpcIndex of %d\n", ags.lastKnownWSEndpointsIndex)
	if _, err := ags.Subscribe(ags.lastKnownWSEndpointsIndex); err != nil {
		ags.Reconnect()
	}
}

func (ags *AggregateSubscription) IsSynced() bool {
	return ags.isSynced
}
