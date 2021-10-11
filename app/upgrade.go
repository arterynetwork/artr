package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	referralK "github.com/arterynetwork/artr/x/referral/keeper"
	referralT "github.com/arterynetwork/artr/x/referral/types"
)

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func RecalculateActiveReferrals(k referralK.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RecalculateActiveReferrals ...")
		k.Iterate(ctx, func(acc string, info *referralT.Info) (changed, checkForStatusUpdate bool) {
			arl := make([]string, 0, len(info.Referrals))
			ars := make(map[string]bool, len(info.Referrals))
			for _, rAddr := range info.Referrals {
				rInfo, err := k.Get(ctx, rAddr)
				if err != nil {
					logger.Error("Account %s not found", rAddr)
					panic(err)
				}
				if rInfo.Active {
					arl = append(arl, rAddr)
					ars[rAddr] = true
				}
			}
			oars := make(map[string]bool, len(info.ActiveReferrals))
			for _, addr := range info.ActiveReferrals {
				if !ars[addr] {
					changed = true
					break
				}
				oars[addr] = true
			}
			if !changed {
				for _, addr := range arl {
					if !oars[addr] {
						changed = true
						break
					}
				}
			}
			if changed {
				logger.Debug("... %s: %v → %v", acc, info.ActiveReferrals, arl)
				info.ActiveReferrals = arl
				return true, true
			}
			return false, false
		})
		logger.Info("... done")
	}
}
