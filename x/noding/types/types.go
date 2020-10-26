package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type D struct {
	// Power - voting power (depends on delegated funds)
	Power             int64                    `json:"power"`
	// Status - if the validator is on (true) or off (false)
	Status            bool                     `json:"status"`
	// LastPower - voting power that the validator had during the last block signing.
	// It must be 0 if the validator was not chosen for signing.
	LastPower         int64                    `json:"last_power"`
	// PubKey - consensus public key of assigned node (bech32)
	PubKey            string                   `json:"pub_key"`
	// LastPubKey - last known to TM consensus public key of assigned node (bech32)
	LastPubKey        string                   `json:"last_pub_key"`
	// Strokes - how many times has that validator missed a block
	Strokes           int64                    `json:"strokes"`
	// OkBlocksInRow - how many blocks the validator successfully signed (in row, i.e. without being missing).
	// It must be 0 if the validator missed the last block. It must not be reset if the validator was not chosen for
	// signing a block.
	OkBlocksInRow     int64                    `json:"ok_blocks_in_row"`
	// MissedBlocksInRow - how many blocks the validator missed in row (i.e. without successful signing).
	// It must be 0 if the validator successfully signed the last block. It must not reset if the validator wasn't
	// chosen for signing a block. But it must be reset if the validator is jailed.
	MissedBlocksInRow int64                    `json:"missed_blocks_in_row"`
	// Jailed - if the validator is jailed for missing blocks
	Jailed            bool                     `json:"jailed"`
	// UnjailAt - block height after which the validator can unjail
	UnjailAt          int64                    `json:"unjail_at"`
	// Infractions - evidences of byzantine behavior (from Tendermint)
	Infractions       []abci.Evidence          `json:"infractions"`
	// BannedForLife - is the validator permanently banned
	BannedForLife     bool                     `json:"banned_for_life"`
	// Staff nodes are allowed to be validators even if they are not qualified by status/stake
	Staff             bool                     `json:"staff"`
	// ProposedCount - how many blocks was proposed (successfully) by a validator for the all time
	ProposedCount     int64                    `json:"proposed_count"`
	// JailCount - how many times a validator was jailed for the all time
	JailCount         int64                    `json:"jail_count"`
}

func NewD(power int64, pubKey string) D {
	return D {
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
