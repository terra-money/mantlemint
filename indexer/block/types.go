package block

import (
	tm "github.com/tendermint/tendermint/types"
	"github.com/terra-money/mantlemint/lib"
)

var (
	prefix = []byte("block/height:")
	getKey = func(height uint64) []byte {
		return lib.ConcatBytes(prefix, lib.UintToBigEndian(height))
	}
)

type BlockRecord struct {
	BlockID *tm.BlockID `json:"block_id"`
	Block   *tm.Block   `json:"block"`
}
