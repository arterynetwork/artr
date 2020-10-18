package types

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all voting state that must be provided at genesis
type GenesisState struct {
	Government      []sdk.AccAddress        `json:"government" yaml:"government"`
	Params          Params                  `json:"params" yaml:"params"`
	CurrentProposal Proposal                `json:"current_proposal,omitempty" yaml:"current_proposal,omitempty"`
	StartBlock      int64                   `json:"start_block,omitempty" yaml:"start_block,omitempty"`
	Agreed          []sdk.AccAddress        `json:"agreed,omitempty" yaml:"agreed,omitempty"`
	Disagreed       []sdk.AccAddress        `json:"disagreed,omitempty" yaml:"disagreed,omitempty"`
	History         []ProposalHistoryRecord `json:"history,omitempty" yaml:"history,omitempty"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params          Params,
	gov             Government,
	currentProposal Proposal,
	startBlock      int64,
	agreed          Government,
	disagreed       Government,
	history         []ProposalHistoryRecord,
) GenesisState {
	return GenesisState{
		Params:          params,
		Government:      gov,
		CurrentProposal: currentProposal,
		StartBlock:      startBlock,
		Agreed:          agreed,
		Disagreed:       disagreed,
		History:         history[:],
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the voting genesis parameters
func ValidateGenesis(data GenesisState) error {
	if data.Government == nil || len(data.Government) == 0 {
		return errors.New("no Government accounts")
	}
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return nil
}
