package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/schedule"
)

// ParamSubspace defines the expected Subspace interface
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.Account
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, block uint64, event string, data *[]byte) error
	GetParams(cts sdk.Context) schedule.Params
}

type BankKeeper interface {
	InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}
