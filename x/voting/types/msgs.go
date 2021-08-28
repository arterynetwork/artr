package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgPropose{}
	_ sdk.Msg = &MsgVote{}
)

const (
	ProposeConst = "propose"
	VoteConst    = "vote"
)

func (MsgPropose) Route() string { return RouterKey }
func (MsgPropose) Type() string  { return ProposeConst }
func (msg MsgPropose) ValidateBasic() error {
	if msg.Proposal.EndTime != nil {
		return errors.New("end_time should be empty")
	}
	return msg.Proposal.Validate()
}

func (msg *MsgPropose) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgPropose) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposal.GetAuthor()}
}

func (MsgVote) Route() string { return RouterKey }

func (MsgVote) Type() string { return VoteConst }

func (msg MsgVote) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return errors.Wrap(err, "invalid voter")
	}
	return nil
}

func (msg *MsgVote) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetVoter()}
}

func (msg MsgVote) GetVoter() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		panic(err)
	}
	return addr
}
