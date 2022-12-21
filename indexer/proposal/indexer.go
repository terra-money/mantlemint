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
	"github.com/tendermint/tendermint/proto/tendermint/types"
	tm "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	tapp "github.com/terra-money/alliance/app"
	"github.com/terra-money/mantlemint/db/safe_batch"
	"github.com/terra-money/mantlemint/indexer"
	blockindexer "github.com/terra-money/mantlemint/indexer/block"
	txindexer "github.com/terra-money/mantlemint/indexer/tx"
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
	lastHeight, err := getLastIndexedHeight(db)
	fmt.Printf("[indexer/proposals] last height: %d current height %d\n", lastHeight, block.Height)
	if err != nil {
		return err
	}

	// if last indexed height is not the last block, perform historical indexing by replaying old blocks
	if block.Height-lastHeight != 1 {
		err = indexHistoricalProposals(db, app, lastHeight, block.Height)
	} else {
		err = indexNewBlock(db, app, evc, block)
	}
	if err != nil {
		panic(err)
	}

	err = db.Set(lastIndexedHeightKey, sdk.Uint64ToBigEndian(uint64(block.Height)))
	if err != nil {
		return err
	}

	count := 0
	IterateProposals(db, 0, func(proposal Proposal) bool {
		count += 1
		return false
	})
	fmt.Printf("[indexer/proposals] indexed %d proposals\n", count)
	return nil
}

func indexHistoricalProposals(db safe_batch.SafeBatchDB, app *cosmoscmd.App, lastHeight int64, currentHeight int64) (err error) {
	fmt.Printf("[indexer/proposals] indexing historical from %d\n", lastHeight)
	var proposalsAdded []*Proposal

	// Go through all blocks
	err = blockindexer.IterateBlocks(db, lastHeight, currentHeight, func(block *tm.Block) bool {
		fmt.Println("[indexer/proposals] block", block.Height, "has", len(block.Txs), "txs")
		fmt.Println("[indexer/proposals] processing block", block.Height)
		// For each block, get all transactions to look for proposal creation txns
		for _, txByte := range block.Txs {
			hash := fmt.Sprintf("%X", txByte.Hash())
			var txRecord txindexer.TxRecord
			var txResponse txindexer.ResponseDeliverTx
			txRecord, err = txindexer.GetTxByHash(db, hash)
			if err != nil {
				return true
			}
			err = tmjson.Unmarshal(txRecord.TxResponse, &txResponse)
			if err != nil {
				return true
			}
			for _, event := range txResponse.Events {
				switch event.Type {
				case govtypes.EventTypeProposalDeposit, govtypes.EventTypeSubmitProposal:
					proposalIdStr, found := findAttributeValue(event.Attributes, govtypes.AttributeKeyVotingPeriodStart)
					var proposalId uint64
					if found {
						proposalId, err = strconv.ParseUint(proposalIdStr, 10, 64)
						if err != nil {
							return true
						}
						proposal := NewProposal(proposalId, govtypesv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)
						proposalsAdded = append(proposalsAdded, &proposal)
					}
				}
			}
		}
		if err != nil {
			panic(err)
		}

		// Find all non-completed proposals and get current voter counts
		for _, proposal := range proposalsAdded {
			fmt.Println("[indexer/proposals] Processing proposal", proposal.Id)
			if proposal.Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
				appProposal, err := getProposalFromApp(block.Height, app, proposal.Id)
				if err != nil {
					panic(err)
				}
				if appProposal.Status != govtypesv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
					proposal.Status = appProposal.Status
				} else {
					var votes []Vote
					votes, err = getVoters(block.Height, app, proposal.Id)
					if err != nil {
						panic(err)
					}
					proposal.Votes = votes
				}
			}
		}
		return false
	})
	if err != nil {
		panic(err)
	}

	// Update all new proposals
	for _, proposal := range proposalsAdded {
		err = upsertProposal(db, proposal.Id, *proposal)
		if err != nil {
			return err
		}
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

func getProposalFromApp(height int64, app *cosmoscmd.App, proposalId uint64) (proposal govtypesv1.Proposal, err error) {
	cms, err := (*app).CommitMultiStore().CacheMultiStoreWithVersion(height)
	if err != nil {
		return proposal, err
	}
	ctx := sdk.NewContext(cms, types.Header{
		Height: 0,
	}, false, logger)

	tApp, ok := (*app).(*tapp.App)
	if !ok {
		return proposal, fmt.Errorf("invalid app expect: %T got %T", tapp.App{}, tApp)
	}

	proposal, found := tApp.GovKeeper.GetProposal(ctx, proposalId)
	if !found {
		return proposal, fmt.Errorf("proposal with id %d not found", proposalId)
	}
	return proposal, nil
}

func getVoters(height int64, app *cosmoscmd.App, proposalId uint64) (voters []Vote, err error) {
	cms, err := (*app).CommitMultiStore().CacheMultiStoreWithVersion(height)
	if err != nil {
		return voters, err
	}
	ctx := sdk.NewContext(cms, types.Header{
		Height: 0,
	}, false, logger)

	tApp, ok := (*app).(*tapp.App)
	if !ok {
		return voters, fmt.Errorf("invalid app expect: %T got %T", tapp.App{}, tApp)
	}
	voters = []Vote{}

	tApp.GovKeeper.IterateVotes(ctx, proposalId, func(vote govtypesv1.Vote) bool {
		voters = append(voters, Vote{
			Voter:   vote.Voter,
			Options: NewWeightedVoteOptions(vote.Options),
		})
		return false
	})
	return voters, err
}

func getLastIndexedHeight(db safe_batch.SafeBatchDB) (height int64, err error) {
	bz, err := db.Get(lastIndexedHeightKey)
	if err != nil {
		return height, err
	}
	if bz == nil {
		return 1, err
	}
	return int64(sdk.BigEndianToUint64(bz)), nil
}

func findAttributeValue(attrs []txindexer.EventAttribute, key string) (value string, found bool) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value, true
		}
	}
	return value, false
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

func IterateProposals(db safe_batch.SafeBatchDB, fromId uint64, cb func(proposal Proposal) bool) (err error) {
	iter, err := db.Iterator(getProposalKey(fromId), sdk.PrefixEndBytes(proposalKey))
	//iter, err := db.Iterator(getProposalKey(fromId), getProposalKey(1000))
	if err != nil {
		return err
	}
	for iter.Valid() {
		b := iter.Value()
		if b == nil {
			iter.Next()
			continue
		}
		var proposal Proposal
		err = tmjson.Unmarshal(b, &proposal)
		if err != nil {
			return err
		}
		stop := cb(proposal)
		if stop {
			return nil
		}
		iter.Next()
	}
	return nil
}

func indexVotesByVoter(votes []Vote) (m map[string]Vote) {
	m = make(map[string]Vote)
	for _, vote := range votes {
		m[vote.Voter] = vote
	}
	return m
}
