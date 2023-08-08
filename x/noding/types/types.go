package types

import (
	"math"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
)

func NewInfo(pubKey string, delegation int64) *Info {
	val := Info{
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
	val.UpdateScore(delegation)
	return &val
}

func (x Info) IsActive() bool {
	return x.Status && !x.Jailed && !x.BannedForLife
}

func (x *Info) UpdateScore(stake int64) (changed bool) {

	if stake < 1 {
		stake = 1
	}
	score := int64(math.Log(float64(stake))) - x.Strokes + x.OkBlocksInRow/100
	if score == x.Score {
		return false
	}
	x.Score = score
	return true
}

type InfoWithAccount struct {
	Info
	Account sdk.AccAddress
}

func NewInfoWithAccount(acc sdk.AccAddress, info Info) InfoWithAccount {
	return InfoWithAccount{
		Info:    info,
		Account: acc,
	}
}

var (
	ErrVotingPowerNonPositive = errors.New("voting power must be positive")
	ErrNoSlices               = errors.New("at least one slice is required")
	ErrFractionNonPositive    = errors.New("part must be positive")
	ErrWrongPartsTotal        = errors.New("parts total must be equal 100%")
)

func (x Distribution) Validate() error {
	if x.LuckiesVotingPower <= 0 {
		return errors.Wrap(ErrVotingPowerNonPositive, "invalid luckies_voting_power")
	}
	if len(x.Slices) == 0 {
		return ErrNoSlices
	}
	sumP := util.FractionZero()
	for i, slice := range x.Slices {
		if slice.VotingPower <= 0 {
			return errors.Wrapf(ErrVotingPowerNonPositive, "invalid slice #%d", i)
		}
		if !slice.Part.IsPositive() {
			return errors.Wrapf(ErrFractionNonPositive, "invalid slice #%d", i)
		}
		sumP = sumP.Add(slice.Part)
	}
	if !sumP.Equal(util.FractionInt(1)) {
		return ErrWrongPartsTotal
	}
	return nil
}
