package mantlemint

import (
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/types"
)

func NewMantlemintGenesisProvider(doc *types.GenesisDoc) node.GenesisDocProvider {
	return func() (*types.GenesisDoc, error) {
		return doc, nil
	}
}
