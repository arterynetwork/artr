package types

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (v Validator) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(v.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

func (v Validator) ToInfo(stake int64) Info {
	res := Info{
		PubKey:            v.PubKey,
		Strokes:           v.Strokes,
		OkBlocksInRow:     v.OkBlocksInRow,
		MissedBlocksInRow: v.MissedBlocksInRow,
		Jailed:            v.Jailed,
		Status:            v.Jailed && v.SwitchedOn,
		UnjailAt:          v.UnjailAt,
		Infractions:       v.Infractions,
		BannedForLife:     v.Banned,
		Staff:             v.Staff,
		ProposedCount:     v.ProposedCount,
		JailCount:         v.JailCount,
	}
	res.UpdateScore(stake)
	return res
}

func GenesisValidatorFromD(acc sdk.AccAddress, info Info, proposedBlocks []uint64) Validator {
	return Validator{
		Account:           acc.String(),
		PubKey:            info.PubKey,
		Strokes:           info.Strokes,
		OkBlocksInRow:     info.OkBlocksInRow,
		MissedBlocksInRow: info.MissedBlocksInRow,
		Jailed:            info.Jailed,
		UnjailAt:          info.UnjailAt,
		Infractions:       info.Infractions,
		Banned:            info.BannedForLife,
		Staff:             info.Staff,
		ProposedCount:     info.ProposedCount,
		JailCount:         info.JailCount,
		SwitchedOn:        info.Jailed && info.Status,
		ProposedBlocks:    proposedBlocks,
	}
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, active []Validator, nonactive []Validator) *GenesisState {
	return &GenesisState{
		Params:    params,
		Active:    active,
		NonActive: nonactive,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the noding genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return errors.Wrap(err, "invalid params")
	}
	if err := validateActiveValidators(data.Active); err != nil {
		return errors.Wrap(err, "invalid active")
	}
	if err := validateNonActiveValidators(data.NonActive); err != nil {
		return errors.Wrap(err, "invalid non_active")
	}
	return nil
}

func validateActiveValidators(v []Validator) error {
	if len(v) == 0 {
		return errors.New("empty list")
	}
	for i, val := range v {
		if _, err := sdk.AccAddressFromBech32(val.Account); err != nil {
			return errors.Wrapf(err, "invalid validator #%d: invalid account", i)
		}
		if _, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, val.PubKey); err != nil {
			return errors.Wrapf(err, "invalid validator #%d: invalid pub_key", i)
		}
		if val.Jailed {
			return errors.Errorf("invalid validator #%d: jailed validator cannot be active", i)
		}
		if val.Banned {
			return errors.Errorf("invalid validator #%d: banned validator cannot be active", i)
		}
		if val.ProposedCount < 0 {
			return errors.Errorf("invalid validator #%d: proposed_count must be non-negative", i)
		}
		if val.JailCount < 0 {
			return errors.Errorf("invalid validator #%d: jailed_count must be non-negative", i)
		}
		if val.OkBlocksInRow < 0 {
			return errors.Errorf("invalid validator #%d: ok_blocks_in_row must be non-negative", i)
		}
		if val.MissedBlocksInRow < 0 {
			return errors.Errorf("invalid validator #%d: missed_blocks_in_row must be non-negative", i)
		}
		if val.OkBlocksInRow > 0 && val.MissedBlocksInRow > 0 {
			return errors.Errorf("invalid validator #%d: either OK or missed block counter can be non-zero, not both of them", i)
		}
	}
	return nil
}

func validateNonActiveValidators(v []Validator) error {
	for i, val := range v {
		if _, err := sdk.AccAddressFromBech32(val.Account); err != nil {
			return errors.Wrapf(err, "invalid validator #%d: invalid account", i)
		}
		if len(val.PubKey) > 0 {
			if _, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, val.PubKey); err != nil {
				return errors.Wrapf(err, "invalid validator #%d: invalid pub_key", i)
			}
		}
		if val.ProposedCount < 0 {
			return errors.Errorf("invalid validator #%d: proposed_count must be non-negative", i)
		}
		if val.JailCount < 0 {
			return errors.Errorf("invalid validator #%d: jail_count must be non-negative", i)
		}
		if val.OkBlocksInRow < 0 {
			return errors.Errorf("invalid validator #%d: ok_blocks_in_row must be non-negative", i)
		}
		if val.MissedBlocksInRow < 0 {
			return errors.Errorf("invalid validator #%d: missed_blocks_in_row must be non-negative", i)
		}
		if val.OkBlocksInRow > 0 && val.MissedBlocksInRow > 0 {
			return errors.Errorf("invalid validator #%d: either OK or missed block counter can be non-zero, not both of them", i)
		}
	}
	return nil
}
