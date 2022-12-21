package proposal

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

var (
	lastIndexedHeightKey = []byte("proposal_last_height:")
	proposalKey          = []byte("proposal:")
)

type Proposal struct {
	Id     uint64                    `json:"id"`
	Status govtypesv1.ProposalStatus `json:"status"`
	Votes  []Vote                    `json:"votes"`
}

type Vote struct {
	Voter   string               `json:"voter"`
	Options []WeightedVoteOption `json:"options"`
}

type WeightedVoteOption struct {
	Weight string `json:"weight""`
	Option string `json:"option"`
}

func NewWeightedVoteOptions(options govtypesv1.WeightedVoteOptions) (wvo []WeightedVoteOption) {
	wvo = []WeightedVoteOption{}
	for _, option := range options {
		wvo = append(wvo, WeightedVoteOption{
			Weight: option.Weight,
			Option: govtypesv1.VoteOption_name[int32(option.Option)],
		})
	}
	return wvo
}

func getProposalKey(proposalId uint64) (key []byte) {
	key = append(key, proposalKey...)
	return append(key, sdk.Uint64ToBigEndian(proposalId)...)
}

func NewProposal(id uint64, status govtypesv1.ProposalStatus) Proposal {
	return Proposal{
		Id:     id,
		Status: status,
		Votes:  []Vote{},
	}
}
