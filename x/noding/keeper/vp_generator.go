package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/noding/types"
)

type VpGenerator struct {
	count, maxValidators uint32
	lastScore            int64
	ptr                  int
	partAccum            util.Fraction
	distr                types.Distribution
}

func NewVotingPowerGenerator(k Keeper, ctx sdk.Context) VpGenerator {
	p := k.GetParams(ctx)
	return VpGenerator{
		maxValidators: p.MaxValidators,
		distr:         p.VotingPower,
		partAccum:     util.FractionZero(),
	}
}

func (g *VpGenerator) GetVotingPower(score int64) int64 {
	g.count++
	defer func() {
		g.lastScore = score
	}()

	if g.count > g.maxValidators {
		return g.distr.LuckiesVotingPower
	}
	if g.count == 1 {
		g.partAccum = g.distr.Slices[0].Part
		return g.distr.Slices[0].VotingPower
	}
	if util.NewFraction(int64(g.count), int64(g.maxValidators)).LTE(g.partAccum) || score == g.lastScore {
		return g.distr.Slices[g.ptr].VotingPower
	}

	g.ptr++
	g.partAccum = g.partAccum.Add(g.distr.Slices[g.ptr].Part)
	return g.distr.Slices[g.ptr].VotingPower
}
