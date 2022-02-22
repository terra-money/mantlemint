package block_feed

import (
	tendermint "github.com/tendermint/tendermint/types"
)

// BlockFeed is a standard interface to provide subscription over blocks
// There is only one method OnBlockFound and it gives you access to the
// BlockFeed channel
type BlockFeed interface {
	// Close closes underlying subscriber
	Close() error

	// Subscribe starts subscription to the block source
	Subscribe() (chan *BlockResult, error)

	// IsSynced reports whether feeder has caught up with the most recent block (i.e. running on ws)
	IsSynced() bool

	// Inject allows force injection of block
	Inject(*BlockResult)

	GetBlockFeedChannel() chan *BlockResult
}

type BlockResult struct {
	BlockID *tendermint.BlockID `json:"block_id"`
	Block   *tendermint.Block   `json:"block"`
}
