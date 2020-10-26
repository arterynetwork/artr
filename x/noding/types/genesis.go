package types

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Validator struct {
	Account           sdk.AccAddress  `json:"account"`
	Pubkey            string          `json:"pubkey,omitempty"`
	Strokes           int64           `json:"strokes,omitempty"`
	OkBlocksInRow     int64           `json:"ok_blocks_in_row,omitempty"`
	MissedBlocksInRow int64           `json:"missed_blocks_in_row,omitempty"`
	Jailed            bool            `json:"jailed,omitempty"`
	UnjailAt          int64           `json:"unjail_at,omitempty"`
	Infractions       []abci.Evidence `json:"infractions,omitempty"`
	Banned            bool            `json:"banned,omitempty"`
	Staff             bool            `json:"staff,omitempty"`
	ProposedCount     int64           `json:"proposed_count,omitempty"`
	JailCount         int64           `json:"jail_count,omitempty"`
	SwitchedOn        bool            `json:"switched_on,omitempty"`
	ProposedBlocks    []uint64        `json:"proposed_blocks,omitempty"`
}

func (v Validator) ToD() D {
	return D{
		PubKey:            v.Pubkey,
		Strokes:           v.Strokes,
		OkBlocksInRow:     v.OkBlocksInRow,
		MissedBlocksInRow: v.MissedBlocksInRow,
		Jailed:            v.Jailed,
		Status:            v.Jailed && v.SwitchedOn,
		UnjailAt:          v.UnjailAt,
		Infractions:       v.Infractions[:],
		BannedForLife:     v.Banned,
		Staff:             v.Staff,
		ProposedCount:     v.ProposedCount,
		JailCount:         v.JailCount,
	}
}

func GenesisValidatorFromD(acc sdk.AccAddress, d D, proposedBlocks []uint64) Validator {
	return Validator{
		Account:           acc,
		Pubkey:            d.PubKey,
		Strokes:           d.Strokes,
		OkBlocksInRow:     d.OkBlocksInRow,
		MissedBlocksInRow: d.MissedBlocksInRow,
		Jailed:            d.Jailed,
		UnjailAt:          d.UnjailAt,
		Infractions:       d.Infractions[:],
		Banned:            d.BannedForLife,
		Staff:             d.Staff,
		ProposedCount:     d.ProposedCount,
		JailCount:         d.JailCount,
		SwitchedOn:        d.Jailed && d.Status,
		ProposedBlocks:    proposedBlocks,
	}
}

// GenesisState - all noding state that must be provided at genesis
type GenesisState struct {
	Params              Params      `json:"params"`
	ActiveValidators    []Validator `json:"active"`
	NonActiveValidators []Validator `json:"non_active"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, active []Validator, nonactive []Validator) GenesisState {
	return GenesisState{
		Params:              params,
		ActiveValidators:    active,
		NonActiveValidators: nonactive,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the noding genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil { return err }
	if err := validateActiveValidators(data.ActiveValidators); err != nil { return err }
	if err := validateNonActiveValidators(data.NonActiveValidators); err != nil { return err }
	return nil
}

func validateActiveValidators(v []Validator) error {
	if len(v) == 0 { return fmt.Errorf("empty validator set") }
	for i, val := range v {
		if val.Account.Empty() {
			return fmt.Errorf("empty account address (#%d)", i)
		}
		if _, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, val.Pubkey); err != nil {
			return sdkerrors.Wrapf(err, "invalid pubkey (#%d)", i)
		}
		if val.Jailed {
			return fmt.Errorf("jailed validator cannot be active (#%d)", i)
		}
		if val.Banned {
			return fmt.Errorf("banned validator cannot be active (#%d)", i)
		}
		if val.ProposedCount < 0 {
			return fmt.Errorf("proposed block count mustbe non-negative (#%d)", i)
		}
		if val.JailCount < 0 {
			return fmt.Errorf("jail count mustbe non-negative (#%d)", i)
		}
		if val.OkBlocksInRow < 0 {
			return fmt.Errorf("OK block counter must be non-negative (#%d)", i)
		}
		if val.MissedBlocksInRow < 0 {
			return fmt.Errorf("missed block counter must be non-negative (#%d)", i)
		}
		if val.OkBlocksInRow > 0 && val.MissedBlocksInRow > 0 {
			return fmt.Errorf("either OK or missed block counter can be non-zero, not both of them (#%d)", i)
		}
	}
	return nil
}

func validateNonActiveValidators(v []Validator) error {
	for i, val := range v {
		if val.Account.Empty() {
			return fmt.Errorf("empty account address (#%d)", i)
		}
		if len(val.Pubkey) > 0 {
			if _, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, val.Pubkey); err != nil {
				return sdkerrors.Wrapf(err, "invalid pubkey (#%d)", i)
			}
		}
		if val.ProposedCount < 0 {
			return fmt.Errorf("proposed block count mustbe non-negative (#%d)", i)
		}
		if val.JailCount < 0 {
			return fmt.Errorf("jail count mustbe non-negative (#%d)", i)
		}
		if val.OkBlocksInRow < 0 {
			return fmt.Errorf("OK block counter must be non-negative (#%d)", i)
		}
		if val.MissedBlocksInRow < 0 {
			return fmt.Errorf("missed block counter must be non-negative (#%d)", i)
		}
		if val.OkBlocksInRow > 0 && val.MissedBlocksInRow > 0 {
			return fmt.Errorf("either OK or missed block counter can be non-zero, not both of them (#%d)", i)
		}
	}
	return nil
}
