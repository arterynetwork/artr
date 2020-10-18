package types

import (
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/profile/types"
	referral "github.com/arterynetwork/artr/x/referral/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type BankKeeper interface {
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
}

type ReferralKeeper interface {
	GetReferralFeesForSubscription(ctx sdk.Context, acc sdk.AccAddress) ([]referral.ReferralFee, error)
	SetActive(ctx sdk.Context, acc sdk.AccAddress, value bool) error
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, block uint64, event string, data *[]byte) error
}

type VPNKeeper interface {
	SetLimit(ctx sdk.Context, addr sdk.AccAddress, value int64)
	SetCurrent(ctx sdk.Context, addr sdk.AccAddress, value int64)
	AddLimit(ctx sdk.Context, addr sdk.AccAddress, value int64) (int64, error)
}

type StorageKeeper interface {
	GetLimit(ctx sdk.Context, addr sdk.AccAddress) int64
	GetCurrent(ctx sdk.Context, addr sdk.AccAddress) int64
	SetCurrent(ctx sdk.Context, addr sdk.AccAddress, volume int64)
	SetLimit(ctx sdk.Context, addr sdk.AccAddress, volume int64)
	AddLimit(ctx sdk.Context, addr sdk.AccAddress, volume int64) (int64, error)
}

type SupplyKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	SendCoinsFromAccountToModule(
		ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
	) error
}

type ProfileKeeper interface {
	GetProfile(ctx sdk.Context, addr sdk.AccAddress) *types.Profile
}
