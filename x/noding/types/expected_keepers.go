package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"

	referral "github.com/arterynetwork/artr/x/referral/types"
	"github.com/arterynetwork/artr/x/schedule"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type ReferralKeeper interface {
	GetStatus(ctx sdk.Context, acc sdk.AccAddress) (referral.Status, error)
	GetDelegatedInNetwork(ctx sdk.Context, acc sdk.AccAddress) (sdk.Int, error)
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, block uint64, event string, data *[]byte) error
	GetParams(ctx sdk.Context) schedule.Params
}

type SupplyKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) supply.ModuleAccountI
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
