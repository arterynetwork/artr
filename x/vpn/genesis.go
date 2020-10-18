package vpn

import (
	"github.com/arterynetwork/artr/x/vpn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.SetParams(ctx, data.Params)
	for _, vpnStatus := range data.VpnStatus {
		k.SetInfo(ctx, vpnStatus.Address, vpnStatus.VpnInfo)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	var vpnStatus []types.GenesisVpnInfo

	k.IterateInfo(ctx, func(info types.VpnInfo, addr sdk.AccAddress) (stop bool) {
		vpnStatus = append(vpnStatus, types.GenesisVpnInfo{
			VpnInfo: info,
			Address: addr,
		})

		return false
	})

	return NewGenesisState(k.GetParams(ctx), vpnStatus)
}
