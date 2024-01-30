package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/x/bank"
	ref "github.com/arterynetwork/artr/x/referral/types"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Has(ctx sdk.Context, key []byte) bool
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.AccountI
	SetAccount(ctx sdk.Context, acc auth.AccountI)
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) auth.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type BankKeeper interface {
	GetParams(ctx sdk.Context) bank.Params

	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	PayTxFee(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) (fee sdk.Coins, err error)

	InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error
}

type ReferralKeeper interface {
	GetParams(ctx sdk.Context) (params ref.Params)
	Get(ctx sdk.Context, acc string) (ref.Info, error)
	AppendChild(ctx sdk.Context, parentAcc string, childAcc string) error
	ScheduleCompression(ctx sdk.Context, acc string, compressionAt time.Time)
	MustSetActive(ctx sdk.Context, acc string, value bool)
	MustSetActiveWithoutStatusUpdate(ctx sdk.Context, acc string, value bool)
	Iterate(ctx sdk.Context, callback func(acc string, r *ref.Info) (changed, checkForStatusUpdate bool))
	CompressionPeriod(ctx sdk.Context) time.Duration
	ComeBack(ctx sdk.Context, acc string) error
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, time time.Time, event string, data []byte)
	Delete(ctx sdk.Context, time time.Time, event string, payload []byte)

	OneMonth(ctx sdk.Context) time.Duration
}
