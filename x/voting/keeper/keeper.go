package keeper

import (
	"github.com/arterynetwork/artr/x/delegating"
	"encoding/binary"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/tendermint/tendermint/libs/log"
	"time"

	"github.com/arterynetwork/artr/x/voting/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the voting store
type Keeper struct {
	storeKey           sdk.StoreKey
	cdc                *codec.Codec
	paramspace         types.ParamSubspace
	scheduleKeeper     types.ScheduleKeeper
	upgradeKeeper      types.UprgadeKeeper
	nodingKeeper       types.NodingKeeper
	delegatingKeeper   types.DelegatingKeeper
	referralKeeper     types.ReferralKeeper
	subscriptionKeeper types.SubscriptionKeeper
	profileKeeper      types.ProfileKeeper
	earningKeeper      types.EarningKeeper
	vpnKeeper          types.VpnKeeper
	bankKeeper         types.BankKeeper
}

// NewKeeper creates a voting keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramspace types.ParamSubspace,
	scheduleKeeper types.ScheduleKeeper,
	upgradeKeeper types.UprgadeKeeper,
	nodingKeeper types.NodingKeeper,
	delegatingKeeper types.DelegatingKeeper,
	referralKeeper types.ReferralKeeper,
	subscriptionKeeper types.SubscriptionKeeper,
	profileKeeper types.ProfileKeeper,
	earningKeeper types.EarningKeeper,
	vpnKeeper types.VpnKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	keeper := Keeper{
		storeKey:           key,
		cdc:                cdc,
		paramspace:         paramspace.WithKeyTable(types.ParamKeyTable()),
		scheduleKeeper:     scheduleKeeper,
		upgradeKeeper:      upgradeKeeper,
		nodingKeeper:       nodingKeeper,
		delegatingKeeper:   delegatingKeeper,
		referralKeeper:     referralKeeper,
		subscriptionKeeper: subscriptionKeeper,
		profileKeeper:      profileKeeper,
		earningKeeper:      earningKeeper,
		vpnKeeper:          vpnKeeper,
		bankKeeper:         bankKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetCurrentProposal(ctx sdk.Context) *types.Proposal {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyCurrentVote)

	if bz == nil {
		return nil
	}

	var proposal types.Proposal
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &proposal)

	if err != nil {
		panic(err)
	}

	return &proposal
}

func (k Keeper) SetCurrentProposal(ctx sdk.Context, proposal types.Proposal) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(proposal)
	store.Set(types.KeyCurrentVote, bz)
}

func (k Keeper) GetAgreed(ctx sdk.Context) (gov types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyAgreedMembers)

	if bz == nil {
		return types.NewEmptyGovernment()
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &gov)
	return gov
}

func (k Keeper) SetAgreed(ctx sdk.Context, agreed types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(agreed)
	store.Set(types.KeyAgreedMembers, bz)
}

func (k Keeper) GetDisagreed(ctx sdk.Context) (gov types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyDisagreedMembers)

	if bz == nil {
		return types.NewEmptyGovernment()
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &gov)
	return gov
}

func (k Keeper) SetDisagreed(ctx sdk.Context, disagreed types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(disagreed)
	store.Set(types.KeyDisagreedMembers, bz)
}

func (k Keeper) Validate(gov types.Government,
	aGov types.Government,
	dGov types.Government,
) (complete bool, agreed bool) {

	if len(gov) == (len(aGov) + len(dGov)) {
		complete = true

		if len(aGov) == len(gov) || len(aGov) >= len(gov)*2/3 {
			agreed = true
		}
	}

	return complete, agreed
}

func (k Keeper) SaveProposalToHistory(ctx sdk.Context, store sdk.KVStore) {
	history := types.ProposalHistoryRecord{
		Proposal:   *k.GetCurrentProposal(ctx),
		Government: k.GetGovernment(ctx),
		Agreed:     k.GetAgreed(ctx),
		Disagreed:  k.GetDisagreed(ctx),
		Started:    k.GetStartBlock(ctx),
		Ended:      ctx.BlockHeight(),
	}

	historyBz := k.cdc.MustMarshalBinaryLengthPrefixed(history)
	height := make([]byte, 8)
	binary.BigEndian.PutUint64(height, uint64(ctx.BlockHeight()))
	key := append(types.KeyHistoryPrefix, height...)
	store.Set(key, historyBz)
}

func (k Keeper) AddProposalHistoryRecord(ctx sdk.Context, record types.ProposalHistoryRecord) {
	store := ctx.KVStore(k.storeKey)
	historyBz := k.cdc.MustMarshalBinaryLengthPrefixed(record)
	height := make([]byte, 8)
	binary.BigEndian.PutUint64(height, uint64(record.Ended))
	key := append(types.KeyHistoryPrefix, height...)
	store.Set(key, historyBz)
}

func (k Keeper) SetStartBlock(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, (uint64(ctx.BlockHeight())))
	store.Set(types.KeyStartBlock, bz)
}

func (k Keeper) GetStartBlock(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyStartBlock)
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) EndProposal(ctx sdk.Context, proposal types.Proposal, agreed bool) {
	// Delete scheduled completion
	k.scheduleKeeper.DeleteAllTasksOnBlock(ctx, uint64(proposal.EndBlock), types.HookName)

	store := ctx.KVStore(k.storeKey)

	// Save proposal data to history
	k.SaveProposalToHistory(ctx, store)

	// Delete all proposal info
	store.Delete(types.KeyCurrentVote)
	store.Delete(types.KeyAgreedMembers)
	store.Delete(types.KeyDisagreedMembers)
	store.Delete(types.KeyStartBlock)

	agreedText := "no"

	if agreed {
		agreedText = "yes"
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeProposalEnd,
		sdk.NewAttribute(types.AttributeKeyAgree, agreedText),
	))

	if agreed {
		var err error
		switch proposal.TypeCode {
		case types.ProposalTypeEnterPrice:
			p := k.subscriptionKeeper.GetParams(ctx)
			p.SubscriptionPrice = proposal.Params.(types.PriceProposalParams).Price
			k.subscriptionKeeper.SetParams(ctx, p)
		case types.ProposalTypeDelegationAward:
			pp := proposal.Params.(types.DelegationAwardProposalParams)
			val := delegating.NewPercentage(int(pp.Minimal), int(pp.ThousandPlus), int(pp.TenKPlus), int(pp.HundredKPlus))
			p := k.delegatingKeeper.GetParams(ctx)
			p.Percentage = val
			k.delegatingKeeper.SetParams(ctx, p)
		case types.ProposalTypeDelegationNetworkAward:
			p := k.referralKeeper.GetParams(ctx)
			p.DelegatingAward = proposal.Params.(types.NetworkAwardProposalParams).Award
			k.referralKeeper.SetParams(ctx, p)
		case types.ProposalTypeProductNetworkAward:
			p := k.referralKeeper.GetParams(ctx)
			p.SubscriptionAward = proposal.Params.(types.NetworkAwardProposalParams).Award
			k.referralKeeper.SetParams(ctx, p)
		case types.ProposalTypeGovernmentAdd:
			k.AddGovernor(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeGovernmentRemove:
			k.RemoveGovernor(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeProductVpnBasePrice:
			p := k.subscriptionKeeper.GetParams(ctx)
			p.VPNGBPrice = proposal.Params.(types.PriceProposalParams).Price
			k.subscriptionKeeper.SetParams(ctx, p)
		case types.ProposalTypeProductStorageBasePrice:
			p := k.subscriptionKeeper.GetParams(ctx)
			p.StorageGBPrice = proposal.Params.(types.PriceProposalParams).Price
			k.subscriptionKeeper.SetParams(ctx, p)
		case types.ProposalTypeAddFreeCreator:
			k.profileKeeper.AddFreeCreator(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeRemoveFreeCreator:
			k.profileKeeper.RemoveFreeCreator(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeSoftwareUpgrade:
			p := proposal.Params.(types.SoftwareUpgradeProposalParams)
			err = k.upgradeKeeper.ScheduleUpgrade(ctx, upgrade.Plan{
				Name:   p.Name,
				Time:   time.Time{},
				Height: p.Height,
				Info:   p.Info,
			})
		case types.ProposalTypeCancelSoftwareUpgrade:
			k.upgradeKeeper.ClearUpgradePlan(ctx)
		case types.ProposalTypeStaffValidatorAdd:
			err = k.nodingKeeper.AddToStaff(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeStaffValidatorRemove:
			err = k.nodingKeeper.RemoveFromStaff(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeEarningSignerAdd:
			k.earningKeeper.AddSigner(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeEarningSignerRemove:
			k.earningKeeper.RemoveSigner(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeRateChangeSignerAdd:
			k.subscriptionKeeper.AddCourseChangeSigner(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeRateChangeSignerRemove:
			k.subscriptionKeeper.RemoveCourseChangeSigner(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeVpnCurrentSignerAdd:
			k.vpnKeeper.AddSigner(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeVpnCurrentSignerRemove:
			k.vpnKeeper.RemoveSigner(ctx, proposal.Params.(types.AddressProposalParams).Address)
		case types.ProposalTypeTransitionCost:
			p := k.referralKeeper.GetParams(ctx)
			p.TransitionCost = uint64(proposal.Params.(types.PriceProposalParams).Price)
			k.referralKeeper.SetParams(ctx, p)
		case types.ProposalTypeMinSend:
			k.bankKeeper.SetMinSend(ctx, proposal.Params.(types.MinAmountProposalParams).MinAmount)
		case types.ProposalTypeMinDelegate:
			p := k.delegatingKeeper.GetParams(ctx)
			p.MinDelegate = proposal.Params.(types.MinAmountProposalParams).MinAmount
			k.delegatingKeeper.SetParams(ctx, p)
		case types.ProposalTypeMaxValidators:
			p := k.nodingKeeper.GetParams(ctx)
			p.MaxValidators = proposal.Params.(types.ShortCountProposalParams).Count
			k.nodingKeeper.SetParams(ctx, p)
		case types.ProposalTypeGeneralAmnesty:
			k.nodingKeeper.GeneralAmnesty(ctx)
		case types.ProposalTypeLotteryValidators:
			p := k.nodingKeeper.GetParams(ctx)
			p.LotteryValidators = proposal.Params.(types.ShortCountProposalParams).Count
			k.nodingKeeper.SetParams(ctx, p)
		}
		if err != nil {
			k.Logger(ctx).Error("could not apply voting result due to error",
				"name", proposal.Name,
				"error", err,
			)
		}
	}
}

func (k Keeper) ScheduleEnding(ctx sdk.Context, block int64) {
	k.scheduleKeeper.ScheduleTask(ctx, uint64(block), types.HookName, &[]byte{0x0})
}

func (k Keeper) ProcessSchedule(ctx sdk.Context, data []byte) {
	proposal := k.GetCurrentProposal(ctx)

	if proposal != nil {
		_, agree := k.Validate(
			k.GetGovernment(ctx),
			k.GetAgreed(ctx),
			k.GetDisagreed(ctx),
		)

		k.EndProposal(ctx, *proposal, agree)
	}
}

func (k Keeper) GetHistory(ctx sdk.Context, limit int32, page int32) []types.ProposalHistoryRecord {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyHistoryPrefix)
	defer iterator.Close()

	records := make([]types.ProposalHistoryRecord, 0)
	start := limit * (page - 1)
	end := limit * page

	for current := int32(0); iterator.Valid() && (current < end); iterator.Next() {
		if current < start {
			current++
			continue
		}
		current++
		var record types.ProposalHistoryRecord
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &record)
		records = append(records, record)
	}

	return records
}
