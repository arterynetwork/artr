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
	ProposeConst    = "propose"
	VoteConst       = "vote"
	StartPollConst  = "start_poll"
	AnswerPollConst = "answer_poll"
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

func (MsgStartPoll) Route() string { return RouterKey }

func (MsgStartPoll) Type() string { return StartPollConst }

func (msg MsgStartPoll) ValidateBasic() error {
	if msg.Poll.StartTime != nil {
		return errors.New("start_time should be empty")
	}
	if msg.Poll.EndTime != nil {
		return errors.New("end_time should be empty")
	}
	return msg.Poll.Validate()
}

func (msg *MsgStartPoll) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgStartPoll) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetAuthor()}
}

func (msg MsgStartPoll) GetAuthor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Poll.Author)
	if err != nil {
		panic(err)
	}
	return addr
}

func (MsgAnswerPoll) Route() string { return RouterKey }

func (MsgAnswerPoll) Type() string { return AnswerPollConst }

func (msg MsgAnswerPoll) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Respondent); err != nil {
		return errors.Wrap(err, "cannot parse respondent")
	}
	return nil
}

func (msg *MsgAnswerPoll) GetSignBytes() []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgAnswerPoll) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetRespondent()}
}

func (msg MsgAnswerPoll) GetRespondent() sdk.AccAddress {
	res, err := sdk.AccAddressFromBech32(msg.Respondent)
	if err != nil {
		panic(err)
	}
	return res
}
