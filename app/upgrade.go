package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/arterynetwork/artr/x/referral"
)

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func CliWarningUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {
	fmt.Println(`
╔═════════════════════════════════════════════════════════════╗
║ PLEASE MAKE YOU SURE YOU HAVE UPGRADED CLI CLIENT AS WELL ! ║
╚═════════════════════════════════════════════════════════════╝`,
	)
}

func RefreshStatus(k referral.Keeper, status referral.Status) func(ctx sdk.Context, _ upgrade.Plan) {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		k.Iterate(ctx, func(r referral.DataRecord)(checkForStatusUpdate bool) {
			return r.Status == status
		})
	}
}

func Chain(handlers... upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}