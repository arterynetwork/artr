package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/x/bank"
)

// ParamSubspace defines the expected Subspace interface
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.AccountI
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, time time.Time, event string, data []byte)
	Delete(ctx sdk.Context, time time.Time, event string, payload []byte)

	OneDay(ctx sdk.Context) time.Duration
	OneMonth(ctx sdk.Context) time.Duration
}

type BankKeeper interface {
	GetParams(ctx sdk.Context) bank.Params

	InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
