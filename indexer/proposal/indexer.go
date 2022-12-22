package proposal

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	tapp "github.com/terra-money/alliance/app"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/indexer"
	"github.com/terra-money/mantlemint/mantlemint"
	"golang.org/x/exp/maps"
	"os"
	"strconv"
)

var IndexProposals = indexer.CreateIndexer(proposalIndexerFunc)
var cdc = cosmoscmd.MakeEncodingConfig(tapp.ModuleBasics)
var logger log.Logger

func init() {
	govtypesv1.RegisterInterfaces(cdc.InterfaceRegistry)
	logger = log.NewTMLogger(os.Stdout)
}

func proposalIndexerFunc(db safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector, app *cosmoscmd.App) (err error) {
	if err != nil {
		return err
	}
	err = indexNewBlock(db, app, evc, block)
	if err != nil {
		panic(err)
	}
	return nil
}

func indexNewBlock(db safe_batch.SafeBatchDB, app *cosmoscmd.App, evc *mantlemint.EventCollector, block *tm.Block) (err error) {
	proposalsToUpdate := make(map[uint64]*Proposal)
	// parse tx events for proposal creation or deposit messages
	// then find vote messages
	for i, txByte := range block.Txs {
		txResponse := evc.ResponseDeliverTxs[i]
		if txResponse.IsErr() {
			continue
		}

		tx, err := cdc.TxConfig.TxDecoder()(txByte)
		if err != nil {
			return err
		}

		for _, msg := range tx.GetMsgs() {
			_, isMsgSubmitProposal := msg.(*govtypesv1b1.MsgSubmitProposal)
			if isMsgSubmitProposal {
				for _, event := range txResponse.Events {
					switch event.Type {
					case govtypes.EventTypeSubmitProposal:
						proposalIdStr, found := findAbciAttributeValue(event.Attributes, govtypes.AttributeKeyProposalID)
						var proposalId uint64
						if found {
							proposalId, err = strconv.ParseUint(proposalIdStr, 10, 64)
							if err != nil {
								return err
							}
							// proposal status here doesn't really matter since we are only interested in voters
							proposal := NewProposal(proposalId, govtypesv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
							proposalsToUpdate[proposalId] = &proposal
						}
					}
				}
			}

			// handle votes
			msgVote, ok := msg.(*govtypesv1.MsgVote)
			if ok {
				var proposal *Proposal
				// If proposal was already updated in the same block, use it instead of fetching from db again
				if proposal, ok = proposalsToUpdate[msgVote.ProposalId]; !ok {
					fetchedProposal, err := GetProposal(&db, msgVote.ProposalId)
					if err != nil {
						return err
					}
					proposal = &fetchedProposal
					proposalsToUpdate[msgVote.ProposalId] = &fetchedProposal
				}
				votes := indexVotesByVoter(proposal.Votes)
				// we can directly replace old votes with new votes since new votes will fully replace the previous vote
				// voters cannot delete their vote, they can only change it to yes/no/abstain
				votes[msgVote.Voter] = Vote{
					Voter: msgVote.Voter,
					Options: []WeightedVoteOption{
						{
							Weight: sdk.OneDec().String(),
							Option: msgVote.Option.String(),
						},
					},
				}
				proposal.Votes = maps.Values(votes)
			}
		}
	}

	// check for proposals that are either in voting or completed
	for _, event := range evc.ResponseEndBlock.Events {
		fmt.Println(event.Type, event.Attributes)
		switch event.Type {
		case govtypes.EventTypeActiveProposal, govtypes.EventTypeInactiveProposal:
			for _, attr := range event.Attributes {
				switch string(attr.Key) {
				case govtypes.AttributeKeyProposalID:
					proposalId, err := strconv.ParseUint(string(attr.Value), 10, 64)
					if err != nil {
						return err
					}
					proposal, err := GetProposal(&db, proposalId)
					if err != nil {
						return err
					}
					result, ok := findAbciAttributeValue(event.Attributes, govtypes.AttributeKeyProposalResult)
					if !ok {
						panic(fmt.Errorf("wrongly formatted event, missing %s", govtypes.AttributeKeyProposalResult))
					}
					switch result {
					case govtypes.AttributeValueProposalDropped:
						proposal.Status = govtypesv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
					case govtypes.AttributeValueProposalPassed:
						proposal.Status = govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED
					case govtypes.AttributeValueProposalFailed:
						proposal.Status = govtypesv1.ProposalStatus_PROPOSAL_STATUS_FAILED
					case govtypes.AttributeValueProposalRejected:
						proposal.Status = govtypesv1.ProposalStatus_PROPOSAL_STATUS_REJECTED
					}
					proposalsToUpdate[proposalId] = &proposal
				}
			}
		}
	}
	// update all proposals at once
	for _, proposal := range proposalsToUpdate {
		err = upsertProposal(db, proposal.Id, *proposal)
		if err != nil {
			return err
		}
	}
	return nil
}

func findAbciAttributeValue(attrs []abci.EventAttribute, key string) (value string, found bool) {
	for _, attr := range attrs {
		fmt.Println(string(attr.GetKey()), string(attr.GetValue()))
		attrStr := string(attr.Key)
		if attrStr == key {
			return string(attr.GetValue()), true
		}
	}
	return value, false
}

func upsertProposal(db safe_batch.SafeBatchDB, proposalId uint64, proposal Proposal) (err error) {
	key := getProposalKey(proposalId)
	b, err := tmjson.Marshal(proposal)
	if err != nil {
		return err
	}
	err = db.Set(key, b)
	if err != nil {
		return err
	}
	return nil
}

func GetProposal(db tmdb.DB, proposalId uint64) (proposal Proposal, err error) {
	b, err := db.Get(getProposalKey(proposalId))
	if err != nil {
		return proposal, err
	}
	if b == nil {
		return proposal, fmt.Errorf("proposal with id %d not found", proposalId)
	}
	err = tmjson.Unmarshal(b, &proposal)
	if err != nil {
		return proposal, err
	}
	if proposal.Votes == nil {
		proposal.Votes = []Vote{}
	}
	return proposal, err
}

func indexVotesByVoter(votes []Vote) (m map[string]Vote) {
	m = make(map[string]Vote)
	for _, vote := range votes {
		m[vote.Voter] = vote
	}
	return m
}
