package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bank "github.com/arterynetwork/artr/x/bank/types"
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/noding"
	profile "github.com/arterynetwork/artr/x/profile/types"
	"github.com/arterynetwork/artr/x/referral"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, time time.Time, event string, data []byte)
	DeleteAll(ctx sdk.Context, time time.Time, event string)
}

type UprgadeKeeper interface {
	ScheduleUpgrade(ctx sdk.Context, plan upgrade.Plan) error
	ClearUpgradePlan(ctx sdk.Context)
}

type NodingKeeper interface {
	AddToStaff(ctx sdk.Context, acc sdk.AccAddress) error
	RemoveFromStaff(ctx sdk.Context, acc sdk.AccAddress) error

	GetParams(ctx sdk.Context) (params noding.Params)
	SetParams(ctx sdk.Context, params noding.Params)

	GeneralAmnesty(ctx sdk.Context)
}

type DelegatingKeeper interface {
	GetParams(ctx sdk.Context) (params delegating.Params)
	SetParams(ctx sdk.Context, params delegating.Params)
}

type ReferralKeeper interface {
	GetParams(ctx sdk.Context) (params referral.Params)
	SetParams(ctx sdk.Context, params referral.Params)
}

type ProfileKeeper interface {
	GetParams(ctx sdk.Context) profile.Params
	SetParams(ctx sdk.Context, params profile.Params)

	AddFreeCreator(ctx sdk.Context, creator sdk.AccAddress)
	RemoveFreeCreator(ctx sdk.Context, creator sdk.AccAddress)
	AddTokenRateSigner(ctx sdk.Context, address sdk.AccAddress)
	RemoveTokenRateSigner(ctx sdk.Context, address sdk.AccAddress)
	AddVpnCurrentSigner(ctx sdk.Context, address sdk.AccAddress)
	RemoveVpnCurrentSigner(ctx sdk.Context, address sdk.AccAddress)
	AddStorageCurrentSigner(ctx sdk.Context, address sdk.AccAddress)
	RemoveStorageCurrentSigner(ctx sdk.Context, address sdk.AccAddress)
}

type signersKeeper interface {
	AddSigner(ctx sdk.Context, address sdk.AccAddress)
	RemoveSigner(ctx sdk.Context, address sdk.AccAddress)
}
type EarningKeeper signersKeeper

type BankKeeper interface {
	GetParams(ctx sdk.Context) bank.Params
	SetParams(ctx sdk.Context, params bank.Params)
}
