package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}
