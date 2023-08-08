package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/x/bank"
	referral "github.com/arterynetwork/artr/x/referral/types"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type ReferralKeeper interface {
	GetStatus(ctx sdk.Context, acc string) (referral.Status, error)
	GetDelegatedInNetwork(ctx sdk.Context, acc string, maxDepth int) (sdk.Int, error)
}

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type BankKeeper interface {
	GetParams(ctx sdk.Context) bank.Params
	GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnAccCoins(ctx sdk.Context, acc sdk.AccAddress, amt sdk.Coins) error
}
