package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}
