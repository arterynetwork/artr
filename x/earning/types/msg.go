package types

import (
	"github.com/arterynetwork/artr/util"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// verify interface at compile time
var _ MsgEarningCommandI = &MsgListEarners{}
var _ MsgEarningCommandI = &MsgRun{}
var _ MsgEarningCommandI = &MsgReset{}

type MsgEarningCommandI interface {
	sdk.Msg
	GetSender() sdk.AccAddress
}

type MsgListEarners struct {
	Sender  sdk.AccAddress `json:"sender"`
	Earners []Earner       `json:"earners"`
}

type MsgRun struct {
	Sender             sdk.AccAddress `json:"sender"`
	FundPart           util.Fraction  `json:"fund_part"`
	AccountPerBlock    uint16         `json:"per_block"`
	TotalVpnPoints     int64          `json:"total_vpn"`
	TotalStoragePoints int64          `json:"total_storage"`
	Height             int64          `json:"height"`
}

type MsgReset struct {
	Sender sdk.AccAddress `json:"sender"`
}

// NewMsg<Action> creates a new Msg<Action> instance
func NewMsgListEarners(sender sdk.AccAddress, earners []Earner) MsgListEarners {
	return MsgListEarners{
		Earners: earners[:],
		Sender:  sender,
	}
}

func NewMsgRun(sender sdk.AccAddress, fundPart util.Fraction, perBlock uint16, totalVpn int64, totalStorage int64, height int64) MsgRun {
	return MsgRun{
		FundPart:           fundPart,
		AccountPerBlock:    perBlock,
		TotalVpnPoints:     totalVpn,
		TotalStoragePoints: totalStorage,
		Height:             height,
		Sender:             sender,
	}
}

func NewMsgReset(sender sdk.AccAddress) MsgReset {
	return MsgReset{sender}
}

const ListEarnersConst = "list-earners"
const RunConst         = "run"
const ResetConst       = "reset"

// nolint
func (msg MsgListEarners) GetSender() sdk.AccAddress { return msg.Sender }
func (msg MsgListEarners) Route() string { return RouterKey }
func (msg MsgListEarners) Type() string  { return ListEarnersConst }
func (msg MsgListEarners) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgListEarners) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgListEarners) ValidateBasic() error {
	if msg.Sender.Empty() { return fmt.Errorf("missing sender") }
	if len(msg.Earners) == 0 {
		return fmt.Errorf( "missing earners list")
	}
	for i, earner := range msg.Earners {
		if earner.Account.Empty() {
			return fmt.Errorf("missing address (#%d)", i)
		}
		if earner.Vpn < 0 {
			return fmt.Errorf("vpn points must be non-negative (#%d)", i)
		}
		if earner.Storage < 0 {
			return fmt.Errorf("storage points must be non-negative (#%d)", i)
		}
	}
	return nil
}

// nolint
func (msg MsgRun) GetSender() sdk.AccAddress { return msg.Sender }
func (msg MsgRun) Route() string { return RouterKey }
func (msg MsgRun) Type() string  { return RunConst }
func (msg MsgRun) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgRun) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgRun) ValidateBasic() error {
	if msg.Sender.Empty() { return fmt.Errorf("missing sender") }
	if !msg.FundPart.IsPositive() {
		return fmt.Errorf("fund part must be positive")
	}
	if msg.FundPart.GT(util.FractionFromInt64(1)) {
		return fmt.Errorf("fund part must be less than or equal 1")
	}
	if msg.AccountPerBlock <= 0 {
		return fmt.Errorf("account per block must be positive")
	}
	if msg.TotalVpnPoints < 0 {
		return fmt.Errorf("total VPN points must be non-negative")
	}
	if msg.TotalStoragePoints < 0 {
		return fmt.Errorf("total storage points must be non-negative")
	}
	return nil
}

// nolint
func (msg MsgReset) GetSender() sdk.AccAddress { return msg.Sender }
func (msg MsgReset) Route() string { return RouterKey }
func (msg MsgReset) Type() string  { return ResetConst }
func (msg MsgReset) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgReset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgReset) ValidateBasic() error {
	if msg.Sender.Empty() { return fmt.Errorf("missing sender") }
	return nil
}
