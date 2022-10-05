package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/x/bank"
	profile "github.com/arterynetwork/artr/x/profile/types"
	referral "github.com/arterynetwork/artr/x/referral/types"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress

}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, time time.Time, event string, data []byte)
	Delete(ctx sdk.Context, time time.Time, event string, payload []byte)

	OneDay(ctx sdk.Context) time.Duration
	OneWeek(ctx sdk.Context) time.Duration
}

type BankKeeper interface {
	GetParams(ctx sdk.Context) bank.Params

	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error

	GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx sdk.Context) bank.Supply
	SetSupply(ctx sdk.Context, supply bank.Supply)

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	PayTxFee(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) (fee sdk.Coins, err error)
}

type ProfileKeeper interface {
	GetProfile(ctx sdk.Context, addr sdk.AccAddress) *profile.Profile
}

type ReferralKeeper interface {
	GetReferralFeesForDelegating(ctx sdk.Context, acc string) ([]referral.ReferralFee, error)
}

type NodingKeeper interface {
	IsActiveValidator(ctx sdk.Context, accAddr sdk.AccAddress) (bool, error)
}
