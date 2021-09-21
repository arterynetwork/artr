package app

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/arterynetwork/artr/util"
	nodingK "github.com/arterynetwork/artr/x/noding/keeper"
	noding "github.com/arterynetwork/artr/x/noding/types"
)

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func InitializeVotingPower(k nodingK.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitializeVotingPower...")

		var pz noding.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, noding.KeyVotingPower) {
				pz.VotingPower = noding.Distribution{
					Slices: []noding.Distribution_Slice{
						{
							Part:        util.Percent(15),
							VotingPower: 15,
						},
						{
							Part:        util.Percent(85),
							VotingPower: 10,
						},
					},
					LuckiesVotingPower: 10,
				}
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		k.SetParams(ctx, pz)

		logger.Info("... done!")
	}
}
