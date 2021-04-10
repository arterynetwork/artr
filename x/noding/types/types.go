package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type D struct {
	// Power - voting power (depends on delegated funds)
	Power int64 `json:"power"`
	// Status - if the validator is on (true) or off (false)
	Status bool `json:"status"`
	// LastPower - voting power that the validator had during the last block signing.
	// It must be 0 if the validator was not chosen for signing.
	LastPower int64 `json:"last_power"`
	// PubKey - consensus public key of assigned node (bech32)
	PubKey string `json:"pub_key"`
	// LastPubKey - last known to TM consensus public key of assigned node (bech32)
	LastPubKey string `json:"last_pub_key"`
	// Strokes - how many times has that validator missed a block
	Strokes int64 `json:"strokes"`
	// OkBlocksInRow - how many blocks the validator successfully signed (in row, i.e. without being missing).
	// It must be 0 if the validator missed the last block. It must not be reset if the validator was not chosen for
	// signing a block.
	OkBlocksInRow int64 `json:"ok_blocks_in_row,omitempty" yaml:"ok_blocks_in_row,omitempty"`
	// MissedBlocksInRow - how many blocks the validator missed in row (i.e. without successful signing).
	// It must be 0 if the validator successfully signed the last block. It must not reset if the validator wasn't
	// chosen for signing a block. But it must be reset if the validator is jailed.
	MissedBlocksInRow int64 `json:"missed_blocks_in_row"`
	// Jailed - if the validator is jailed for missing blocks
	Jailed bool `json:"jailed"`
	// UnjailAt - block height after which the validator can unjail
	UnjailAt int64 `json:"unjail_at"`
	// Infractions - evidences of byzantine behavior (from Tendermint)
	Infractions []abci.Evidence `json:"infractions"`
	// BannedForLife - is the validator permanently banned
	BannedForLife bool `json:"banned_for_life"`
	// Staff nodes are allowed to be validators even if they are not qualified by status/stake
	Staff bool `json:"staff"`
	// ProposedCount - how many blocks was proposed (successfully) by a validator for the all time
	ProposedCount int64 `json:"proposed_count"`
	// JailCount - how many times a validator was jailed for the all time
	JailCount int64 `json:"jail_count"`
	// LotteryNo - account's number in the lottery validators' queue
	LotteryNo uint64 `json:"lottery_no,omitempty" yaml:"lottery_no,omitempty"`
}

func NewD(power int64, pubKey string) D {
	return D{
		Power:             power,
		Status:            true,
		LastPower:         0,
		PubKey:            pubKey,
		Strokes:           0,
		OkBlocksInRow:     0,
		MissedBlocksInRow: 0,
		Jailed:            false,
		UnjailAt:          0,
		Infractions:       nil,
		BannedForLife:     false,
		Staff:             false,
		ProposedCount:     0,
		JailCount:         0,
	}
}

func (d D) IsActive() bool {
	return d.Status && !d.Jailed && !d.BannedForLife && d.Power != 0
}

type KeyedD struct {
	D
	Account sdk.AccAddress
}

func NewKeyedD(acc sdk.AccAddress, d D) KeyedD {
	return KeyedD{
		D:       d,
		Account: acc,
	}
}

// ValidatorState - enum of all possible account's validation states:
//
// ValidatorStateOff - validation's being switched off and can be switched on if stake and status are enough.
//
// ValidatorStateBan - the account is banned for life for malevolent behavior and validation can never been switched on.
//
// ValidatorStateJail - validation's suspended and can be resumed after jail period is over.
//
// ValidatorStateSpare - validator is in reserve now, it  can be entitled to block signing as soon as somebody frees
// a position.
//
// ValidatorStateLucky - validator takes one of "lucky" slots; it can sign a block, but will be moved to the reserve
// after it did so or failed in any way.
//
// ValidatorStateTop - validator takes one of "top" slots; it can sign blocks while its rating is high enough to keep
// the position.
type ValidatorState byte

const (
	ValidatorStateOff   ValidatorState = iota // ValidatorStateOff - validation's being switched off and can be switched on if stake and status are enough.
	ValidatorStateBan                         // ValidatorStateBan - the account is banned for life for malevolent behavior and validation can never been switched on.
	ValidatorStateJail                        // ValidatorStateJail - validation's suspended and can be resumed after jail period is over.
	ValidatorStateSpare                       // ValidatorStateSpare - validator is in reserve now, it  can be entitled to block signing as soon as somebody frees a position.
	ValidatorStateLucky                       // ValidatorStateLucky - validator takes one of "lucky" slots; it can sign a block, but will be moved to the reserve after it did so or failed in any way.
	ValidatorStateTop                         // ValidatorStateTop - validator takes one of "top" slots; it can sign blocks while its rating is high enough to keep the position.
)

func (x ValidatorState) String() string {
	switch x {
	case ValidatorStateOff:
		return "off"
	case ValidatorStateBan:
		return "ban"
	case ValidatorStateJail:
		return "jail"
	case ValidatorStateSpare:
		return "spare"
	case ValidatorStateLucky:
		return "lucky"
	case ValidatorStateTop:
		return "top"
	default:
		return fmt.Sprintf("0x%2X", byte(x))
	}
}
