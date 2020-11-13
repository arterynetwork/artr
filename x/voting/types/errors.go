package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// Signer not in government list
	ErrSignerNotAllowed          = sdkerrors.Register(ModuleName, 1, "signer not in government list")
	ErrOtherActive               = sdkerrors.Register(ModuleName, 2, "other proposal is active")
	ErrAlreadyVoted              = sdkerrors.Register(ModuleName, 3, "already voted")
	ErrNoActiveProposals         = sdkerrors.Register(ModuleName, 4, "no active proposals to vote")
	ErrProposalGovernorExists    = sdkerrors.Register(ModuleName, 5, "candidate already in government list")
	ErrProposalGovernorNotExists = sdkerrors.Register(ModuleName, 6, "candidate not in government list")
	ErrProposalGovernorLast      = sdkerrors.Register(ModuleName, 7, "cannot remove the last governor")
)
