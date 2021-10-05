package types

import (
	"time"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewCompression(acc string, at time.Time) *Compression {
	return &Compression{
		Account: acc,
		Time:    at,
	}
}

func NewDowngrade(acc string, current Status, at time.Time) *Downgrade {
	return &Downgrade{
		Account: acc,
		Current: current,
		Time:    at,
	}
}

func (d Downgrade) GetAccount() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(d.Account)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewTransition creates and fully initializes a new Transition instance.
func NewTransition(subject, destination sdk.AccAddress) *Transition {
	return &Transition{
		Subject:     subject.String(),
		Destination: destination.String(),
	}
}

func NewRefs(parent string, children []string) *Refs {
	return &Refs{
		Referrer:  parent,
		Referrals: children,
	}
}

// Validate performs very basic sanity check.
func (t Transition) Validate() error {
	if _, err := sdk.AccAddressFromBech32(t.Subject); err != nil {
		return errors.Wrap(err, "invalid subject address")
	}
	if _, err := sdk.AccAddressFromBech32(t.Destination); err != nil {
		return errors.Wrap(err, "invalid destination address")
	}
	return nil
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	topLevelAccounts []string,
	otherAccounts []Refs,
	banished []Banished,
	neverPaid []string,
	compressions []Compression,
	downgrades []Downgrade,
	transitions []Transition,
) *GenesisState {
	return &GenesisState{
		Params:           params,
		TopLevelAccounts: topLevelAccounts,
		OtherAccounts:    otherAccounts,
		BanishedAccounts: banished,
		NeverPaid:        neverPaid,
		Compressions:     compressions,
		Downgrades:       downgrades,
		Transitions:      transitions,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the referral genesis parameters
func ValidateGenesis(data GenesisState) error {
	if data.TopLevelAccounts == nil {
		return errors.New("empty top level accounts set")
	}
	if err := data.Params.Validate(); err != nil {
		return errors.Wrap(err, "invalid params")
	}
	for i, t := range data.Transitions {
		if err := t.Validate(); err != nil {
			return errors.Wrapf(err, "invalid transition #%d", i)
		}
	}
	return nil
}
