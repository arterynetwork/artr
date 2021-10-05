package types

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func NewGenesisState(params Params, gov Government, current *Proposal, start int64, agreed, disagreed Government, history []ProposalHistoryRecord) *GenesisState {
	var currentProposal Proposal
	if current != nil {
		currentProposal = *current
	}
	return &GenesisState{
		Params:          params,
		Government:      gov.Members,
		CurrentProposal: currentProposal,
		StartBlock:      start,
		Agreed:          agreed.Members,
		Disagreed:       disagreed.Members,
		History:         history,
	}
}

// ValidateGenesis validates the voting genesis parameters
func ValidateGenesis(data GenesisState) error {
	if len(data.Government) == 0 {
		return errors.New("invalid government: empty list")
	}
	for i, bech32 := range data.Government {
		if _, err := sdk.AccAddressFromBech32(bech32); err != nil {
			return errors.Wrapf(err, "invalid government (item #%d)", i)
		}
	}
	if err := data.Params.Validate(); err != nil {
		return errors.Wrap(err, "invalid params")
	}
	if data.CurrentProposal.Equal(Proposal{}) {
		if data.StartBlock != 0 {
			return errors.New("invalid start_block: must be zero unless current_proposal is set")
		}
		if data.Agreed != nil {
			return errors.New("invalid agreed: must be empty unless current_proposal is set")
		}
		if data.Disagreed != nil {
			return errors.New("invalid disagreed: must be empty unless current_proposal is set")
		}
	} else {
		if err := data.CurrentProposal.Validate(); err != nil {
			return errors.Wrap(err, "invalid current_proposal")
		}
		if data.StartBlock <= 0 {
			return errors.New("invalid start_block: must be positive as current_proposal is set")
		}
		for i, bech32 := range data.Agreed {
			if _, err := sdk.AccAddressFromBech32(bech32); err != nil {
				return errors.Wrapf(err, "invalid agreed (item #%d)", i)
			}
		}
		for i, bech32 := range data.Disagreed {
			if _, err := sdk.AccAddressFromBech32(bech32); err != nil {
				return errors.Wrapf(err, "invalid disagreed (item #%d)", i)
			}
		}
	}
	for i, r := range data.History {
		if err := r.Validate(); err != nil {
			return errors.Wrapf(err, "invalid history (item #%d)", i)
		}
	}
	return nil
}
