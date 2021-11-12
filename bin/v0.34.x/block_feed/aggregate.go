package block_feed

import (
	"fmt"
	"log"
	"time"
)

var _ BlockFeed = (*AggregateSubscription)(nil)

type AggregateSubscription struct {
	ws                    *WSSubscription
	rpc                   *RPCSubscription
	lastKnownBlock        int64
	lastKnownEndpointIdx  int
	aggregateBlockChannel chan *BlockResult
	wsEndpointsLength     int
	isSynced              bool
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
		ws:                    ws,
		rpc:                   rpc,
		lastKnownBlock:        currentBlock,
		lastKnownEndpointIdx:  0,
		aggregateBlockChannel: make(chan *BlockResult),
		wsEndpointsLength:     len(wsEndpoints),
		isSynced:              false,
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
	ags.setSyncState(false)

	// read the first cWS
	<-cWS

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
					ags.lastKnownBlock = r.Block.Height
				} else {
					break
				}
			}

			log.Printf("[block_feed/aggregate] switching to ws...")

			// patch ws to aggregate
			for {
				r := <-cWS

				// gracefully handle done signal; in whatever case received is nil,
				// handle reconnection here
				if r == done {
					log.Printf("[block_feed/aggregate] websocket done signal received, reconnecting...")
					ags.setSyncState(false)
					ags.Close()
					ags.Reconnect()
					break
				} else {
					// if block feeder got upto this point,
					// it is relatively safe that mantle is synced
					ags.setSyncState(true)
					ags.aggregateBlockChannel <- r
					ags.lastKnownBlock = r.Block.Height
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
	endpointIndex := ags.nextWSEndpoint()
	time.Sleep(time.Second)

	log.Printf("[block_feed/aggregate] reconnecting with rpcIndex of %d\n", endpointIndex)
	if _, err := ags.Subscribe(endpointIndex); err != nil {
		ags.Reconnect()
	}
}

func (ags *AggregateSubscription) IsSynced() bool {
	return ags.isSynced
}

func (ags *AggregateSubscription) setSyncState(state bool) {
	ags.isSynced = state
}

func (ags *AggregateSubscription) nextWSEndpoint() int {
	ags.lastKnownEndpointIdx++
	ags.lastKnownEndpointIdx = ags.lastKnownEndpointIdx % ags.wsEndpointsLength

	return ags.lastKnownEndpointIdx
}
