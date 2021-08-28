package types

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
)

// verify interface at compile time
var _ MsgEarningCommandI = &MsgListEarners{}
var _ MsgEarningCommandI = &MsgRun{}
var _ MsgEarningCommandI = &MsgReset{}

type MsgEarningCommandI interface {
	sdk.Msg
	GetSigner() sdk.AccAddress
}

// NewMsg<Action> creates a new Msg<Action> instance
func NewMsgListEarners(sender sdk.AccAddress, earners []Earner) *MsgListEarners {
	return &MsgListEarners{
		Earners: earners,
		Signer:  sender.String(),
	}
}

func NewMsgRun(
	sender sdk.AccAddress,
	fundPart util.Fraction, perBlock uint32,
	totalVpn int64, totalStorage int64,
	time time.Time,
) *MsgRun {
	return &MsgRun{
		FundPart:     fundPart,
		PerBlock:     perBlock,
		TotalVpn:     totalVpn,
		TotalStorage: totalStorage,
		Time:         time,
		Signer:       sender.String(),
	}
}

func NewMsgReset(sender sdk.AccAddress) *MsgReset {
	return &MsgReset{Signer: sender.String()}
}

const ListEarnersConst = "list-earners"
const RunConst = "run"
const ResetConst = "reset"

func (msg MsgListEarners) GetSigner() sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return acc
}

func (msg MsgListEarners) Route() string {
	return RouterKey
}

func (msg MsgListEarners) Type() string {
	return ListEarnersConst
}

func (msg MsgListEarners) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSigner()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgListEarners) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgListEarners) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errors.Wrap(err, "invalid signer")
	}
	if len(msg.Earners) == 0 {
		return errors.New("missing earners list")
	}
	for i, earner := range msg.Earners {
		if _, err := sdk.AccAddressFromBech32(earner.Account); err != nil {
			return errors.Wrapf(err, "invalid earner #%d acc address", i)
		}
		if earner.Vpn < 0 {
			return errors.Errorf("vpn points must be non-negative (#%d)", i)
		}
		if earner.Storage < 0 {
			return errors.Errorf("storage points must be non-negative (#%d)", i)
		}
	}
	return nil
}

func (msg MsgRun) GetSigner() sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return acc
}

func (msg MsgRun) Route() string {
	return RouterKey
}

func (msg MsgRun) Type() string {
	return RunConst
}

func (msg MsgRun) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSigner()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgRun) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgRun) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errors.Wrap(err, "invalid signer")
	}
	if !msg.FundPart.IsPositive() {
		return errors.New("fund_part must be positive")
	}
	if msg.FundPart.GT(util.FractionInt(1)) {
		return errors.New("fund_part must be less than or equal 1")
	}
	if msg.PerBlock <= 0 {
		return errors.New("per_block must be positive")
	}
	if msg.TotalVpn < 0 {
		return errors.New("total_vpn must be non-negative")
	}
	if msg.TotalStorage < 0 {
		return errors.New("total_storage must be non-negative")
	}
	return nil
}

func (msg MsgReset) GetSigner() sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return acc
}

func (MsgReset) Route() string {
	return RouterKey
}

func (MsgReset) Type() string {
	return ResetConst
}

func (msg MsgReset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.GetSigner()}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgReset) GetSignBytes() []byte {
	bz, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgReset) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errors.Wrap(err, "invalid signer")
	}
	return nil
}
