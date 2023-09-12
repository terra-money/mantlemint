package mantlemint

import (
	abci "github.com/cometbft/cometbft/abci/types"
	tm "github.com/cometbft/cometbft/types"
)

type EventCollector struct {
	Height             int64
	Block              *tm.Block
	ResponseBeginBlock *abci.ResponseBeginBlock
	ResponseEndBlock   *abci.ResponseEndBlock
	ResponseDeliverTxs []*abci.ResponseDeliverTx
}

func NewMantlemintEventCollector() *EventCollector {
	return &EventCollector{}
}

// PublishEventNewBlock collects block, ResponseBeginBlock, ResponseEndBlock
func (ev *EventCollector) PublishEventNewBlock(
	block tm.EventDataNewBlock,
) error {
	ev.Height = block.Block.Height
	ev.Block = block.Block
	ev.ResponseBeginBlock = &block.ResultBeginBlock
	ev.ResponseEndBlock = &block.ResultEndBlock

	return nil
}

// PublishEventTx collect txResult in order
func (ev *EventCollector) PublishEventTx(
	txEvent tm.EventDataTx,
) error {
	ev.ResponseDeliverTxs = append(ev.ResponseDeliverTxs, &txEvent.Result)
	return nil
}

// PublishEventNewBlockHeader unused
func (ev *EventCollector) PublishEventNewBlockHeader(
	_ tm.EventDataNewBlockHeader,
) error {
	return nil
}

// PublishEventValidatorSetUpdates unused
func (ev *EventCollector) PublishEventValidatorSetUpdates(
	_ tm.EventDataValidatorSetUpdates,
) error {
	return nil
}

// PublishEventNewEvidence unused
func (ev *EventCollector) PublishEventNewEvidence(_ tm.EventDataNewEvidence) error {
	return nil
}
