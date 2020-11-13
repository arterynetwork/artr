package app

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func CliWarningUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {
	fmt.Println(`
╔═════════════════════════════════════════════════════════════╗
║ PLEASE MAKE YOU SURE YOU HAVE UPGRADED CLI CLIENT AS WELL ! ║
╚═════════════════════════════════════════════════════════════╝`,
	)
}
