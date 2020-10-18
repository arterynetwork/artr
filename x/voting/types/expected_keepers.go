package types

import (
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/subscription"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	WithKeyTable(table params.KeyTable) params.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps params.ParamSet)
	SetParamSet(ctx sdk.Context, ps params.ParamSet)
}

type ScheduleKeeper interface {
	ScheduleTask(ctx sdk.Context, block uint64, event string, data *[]byte) error
	DeleteAllTasksOnBlock(ctx sdk.Context, block uint64, event string)
}

type UprgadeKeeper interface {
	ScheduleUpgrade(ctx sdk.Context, plan upgrade.Plan) error
	ClearUpgradePlan(ctx sdk.Context)
}

type NodingKeeper interface {
	AddToStaff(ctx sdk.Context, acc sdk.AccAddress) error
	RemoveFromStaff(ctx sdk.Context, acc sdk.AccAddress) error
}

type DelegatingKeeper interface {
	GetParams(ctx sdk.Context) (params delegating.Params)
	SetParams(ctx sdk.Context, params delegating.Params)
}

type ReferralKeeper interface {
	GetParams(ctx sdk.Context) (params referral.Params)
	SetParams(ctx sdk.Context, params referral.Params)
}

type SubscriptionKeeper interface {
	GetParams(ctx sdk.Context) (params subscription.Params)
	SetParams(ctx sdk.Context, params subscription.Params)
	AddCourseChangeSigner(ctx sdk.Context, address sdk.AccAddress)
	RemoveCourseChangeSigner(ctx sdk.Context, address sdk.AccAddress)
}

type ProfileKeeper interface {
	AddFreeCreator(ctx sdk.Context, creator sdk.AccAddress)
	RemoveFreeCreator(ctx sdk.Context, creator sdk.AccAddress)
}

type signersKeeper interface {
	AddSigner(ctx sdk.Context, address sdk.AccAddress)
	RemoveSigner(ctx sdk.Context, address sdk.AccAddress)
}
type EarningKeeper signersKeeper
type VpnKeeper     signersKeeper
