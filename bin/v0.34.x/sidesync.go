package main

import (
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer/block"
	"github.com/terra-money/mantlemint-provider-v0.34.x/indexer/tx"
)

func main() {
	go block.SidesyncBlock()
	go tx.SidesyncTx()
}
