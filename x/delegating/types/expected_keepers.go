package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/arterynetwork/artr/x/bank"
	profile "github.com/arterynetwork/artr/x/profile/types"
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

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.Account
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, block uint64, event string, data *[]byte) error
	GetParams(ctx sdk.Context) schedule.Params
}

type SupplyKeeper interface {
	GetSupply(ctx sdk.Context) supply.SupplyI
	SetSupply(ctx sdk.Context, supply supply.SupplyI)

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetDustDelegation(ctx sdk.Context) int64

	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)
	InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)
}

type ProfileKeeper interface {
	GetProfile(ctx sdk.Context, addr sdk.AccAddress) *profile.Profile
}

type ReferralKeeper interface {
	GetReferralFeesForDelegating(ctx sdk.Context, acc sdk.AccAddress) ([]referral.ReferralFee, error)
}
