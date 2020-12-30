package app

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	refTypes "github.com/arterynetwork/artr/x/referral/types"
	"github.com/arterynetwork/artr/x/storage"
)

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func CliWarningUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {
	fmt.Println(`
╔═════════════════════════════════════════════════════════════╗
║ PLEASE MAKE YOU SURE YOU HAVE UPGRADED CLI CLIENT AS WELL ! ║
╚═════════════════════════════════════════════════════════════╝`,
	)
}

func RefreshStatus(k referral.Keeper, status referral.Status) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, func(_ sdk.AccAddress, r *referral.DataRecord) (changed, checkForStatusUpdate bool) {
			return false, r.Status == status
		})
	}
}

func RestoreTrafficLimit(k storage.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, []byte{0x01} /* limitPrefix */, func(key sdk.AccAddress, value *uint64) (changed bool) {
			if *value < 5*util.GBSize {
				*value = 5 * util.GBSize
				return true
			} else {
				return false
			}
		})
	}
}

func ScheduleCompression(k referral.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, func(acc sdk.AccAddress, r *referral.DataRecord) (changed, checkForStatusUpdate bool) {
			var compressionAt int64
			if !r.Active && r.CompressionAt < ctx.BlockHeight() {
				if r.CompressionAt == -1 {
					createdAt, ok := util.CreatedAt[acc.String()]
					if ok {
						compressionAt = createdAt + referral.CompressionPeriod*(1+(ctx.BlockHeight()-createdAt)/referral.CompressionPeriod)
					} else {
						compressionAt = ctx.BlockHeight() + referral.CompressionPeriod
					}
				} else {
					compressionAt = r.CompressionAt + referral.CompressionPeriod*(1+(ctx.BlockHeight()-r.CompressionAt)/referral.CompressionPeriod)
				}
				r.CompressionAt = compressionAt
				changed = true
				if err := k.ScheduleCompression(ctx, acc, compressionAt); err != nil {
					panic(err)
				}
			}
			return changed, false
		})
	}
}

func CountRevoking(ak auth.AccountKeeper, rk referral.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		ak.IterateAccounts(ctx, func(account authTypes.Account) (stop bool) {
			if account.GetCoins().AmountOf(util.ConfigRevokingDenom).IsPositive() {
				if err := rk.OnBalanceChanged(ctx, account.GetAddress()); err != nil {
					panic(err)
				}
			}
			return false
		})
	}
}

func InitializeTransitionCost(k referral.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Debug("Starting InitializeTransitionCost...")
		var pz referral.Params
		for _, pair := range pz.ParamSetPairs() {
			if bytes.Equal(pair.Key, refTypes.KeyTransitionCost) {
				pz.TransitionCost = refTypes.DefaultTransitionCost
			} else {
				paramspace.Get(ctx, pair.Key, pair.Value)
			}
		}
		logger.Debug("Finished InitializeTransitionCost", "params", pz)
		k.SetParams(ctx, pz)
	}
}

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}
