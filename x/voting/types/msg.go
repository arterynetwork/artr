package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateProposal{}

const CreateProposalConst = "create_proposal"

type MsgCreateProposal struct {
	Author   sdk.AccAddress `json:"author" yaml:"author"`
	Name     string         `json:"name" yaml:"name"`
	TypeCode uint8          `json:"type_code" yaml:"type_code"`
	Params   ProposalParams `json:"params" yaml:"params"`
}

func NewMsgCreateProposal(author sdk.AccAddress, name string, typeCode uint8, params ProposalParams) MsgCreateProposal {
	return MsgCreateProposal{
		Author:   author,
		Name:     name,
		TypeCode: typeCode,
		Params:   params,
	}
}

func (msg MsgCreateProposal) Route() string { return RouterKey }
func (msg MsgCreateProposal) Type() string  { return CreateProposalConst }
func (msg MsgCreateProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Author}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgCreateProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgCreateProposal) ValidateBasic() error {
	if msg.Author.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing author address")
	}

	return nil
}

var _ sdk.Msg = &MsgProposalVote{}

const ProposalVoteConst = "proposal_vote"

type MsgProposalVote struct {
	Voter sdk.AccAddress `json:"voter" yaml:"voter"`
	Agree bool           `json:"agree" yaml:"agree"`
}

func NewMsgProposalVote(voter sdk.AccAddress, agree bool) MsgProposalVote {
	return MsgProposalVote{
		Voter: voter,
		Agree: agree,
	}
}

func (msg MsgProposalVote) Route() string { return RouterKey }
func (msg MsgProposalVote) Type() string  { return ProposalVoteConst }
func (msg MsgProposalVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgProposalVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgProposalVote) ValidateBasic() error {
	if msg.Voter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing voter address")
	}

	return nil
}
