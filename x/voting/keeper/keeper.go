package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/x/voting/types"
)

// Keeper of the voting store
type Keeper struct {
	storeKey         sdk.StoreKey
	cdc              codec.BinaryMarshaler
	paramspace       types.ParamSubspace
	scheduleKeeper   types.ScheduleKeeper
	upgradeKeeper    types.UprgadeKeeper
	nodingKeeper     types.NodingKeeper
	delegatingKeeper types.DelegatingKeeper
	referralKeeper   types.ReferralKeeper
	profileKeeper    types.ProfileKeeper
	earningKeeper    types.EarningKeeper
	bankKeeper       types.BankKeeper
}

// NewKeeper creates a voting keeper
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramspace types.ParamSubspace,
	scheduleKeeper types.ScheduleKeeper,
	upgradeKeeper types.UprgadeKeeper,
	nodingKeeper types.NodingKeeper,
	delegatingKeeper types.DelegatingKeeper,
	referralKeeper types.ReferralKeeper,
	profileKeeper types.ProfileKeeper,
	earningKeeper types.EarningKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	keeper := Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramspace:       paramspace.WithKeyTable(types.ParamKeyTable()),
		scheduleKeeper:   scheduleKeeper,
		upgradeKeeper:    upgradeKeeper,
		nodingKeeper:     nodingKeeper,
		delegatingKeeper: delegatingKeeper,
		referralKeeper:   referralKeeper,
		profileKeeper:    profileKeeper,
		earningKeeper:    earningKeeper,
		bankKeeper:       bankKeeper,
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
	err := proto.Unmarshal(bz, &proposal)

	if err != nil {
		panic(err)
	}

	return &proposal
}

func (k Keeper) SetCurrentProposal(ctx sdk.Context, proposal types.Proposal) {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&proposal)
	if err != nil {
		panic(err)
	}
	store.Set(types.KeyCurrentVote, bz)
}

func (k Keeper) GetAgreed(ctx sdk.Context) (gov types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyAgreedMembers)

	if bz == nil {
		return types.Government{}
	}

	if err := proto.Unmarshal(bz, &gov); err != nil {
		panic(err)
	}
	return gov
}

func (k Keeper) SetAgreed(ctx sdk.Context, agreed types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&agreed)
	if err != nil {
		panic(err)
	}
	store.Set(types.KeyAgreedMembers, bz)
}

func (k Keeper) GetDisagreed(ctx sdk.Context) (gov types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyDisagreedMembers)

	if bz == nil {
		return types.Government{}
	}

	if err := proto.Unmarshal(bz, &gov); err != nil {
		panic(err)
	}
	return gov
}

func (k Keeper) SetDisagreed(ctx sdk.Context, disagreed types.Government) {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&disagreed)
	if err != nil {
		panic(err)
	}
	store.Set(types.KeyDisagreedMembers, bz)
}

func (k Keeper) Validate(gov types.Government,
	aGov types.Government,
	dGov types.Government,
) (complete bool, agreed bool) {

	if len(gov.Members) == (len(aGov.Members) + len(dGov.Members)) {
		complete = true
		agreed = len(aGov.Members)*3 >= len(gov.Members)*2
	}

	return complete, agreed
}

func (k Keeper) SaveProposalToHistory(ctx sdk.Context, store sdk.KVStore) {
	history := types.ProposalHistoryRecord{
		Proposal:   *k.GetCurrentProposal(ctx),
		Government: k.GetGovernment(ctx).Members,
		Agreed:     k.GetAgreed(ctx).Members,
		Disagreed:  k.GetDisagreed(ctx).Members,
		Started:    k.GetStartBlock(ctx),
		Finished:   ctx.BlockHeight(),
	}

	historyBz, err := proto.Marshal(&history)
	if err != nil {
		panic(err)
	}
	height := make([]byte, 8)
	binary.BigEndian.PutUint64(height, uint64(ctx.BlockHeight()))
	key := append(types.KeyHistoryPrefix, height...)
	store.Set(key, historyBz)
}

func (k Keeper) AddProposalHistoryRecord(ctx sdk.Context, record types.ProposalHistoryRecord) {
	store := ctx.KVStore(k.storeKey)
	historyBz, err := proto.Marshal(&record)
	if err != nil {
		panic(err)
	}
	height := make([]byte, 8)
	binary.BigEndian.PutUint64(height, uint64(record.Finished))
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
	k.Logger(ctx).Debug("EndProposal", "proposal", proposal, "agreed", agreed)
	// Delete scheduled completion
	k.scheduleKeeper.DeleteAll(ctx, *proposal.EndTime, types.HookName)

	store := ctx.KVStore(k.storeKey)

	// Save proposal data to history
	k.SaveProposalToHistory(ctx, store)

	// Delete all proposal info
	store.Delete(types.KeyCurrentVote)
	store.Delete(types.KeyAgreedMembers)
	store.Delete(types.KeyDisagreedMembers)
	store.Delete(types.KeyStartBlock)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventVotingFinished{
			Name:   proposal.Name,
			Agreed: agreed,
		},
	); err != nil { panic(err) }

	if agreed {
		var err error
		switch proposal.Type {
		case types.PROPOSAL_TYPE_ENTER_PRICE:
			p := k.profileKeeper.GetParams(ctx)
			p.SubscriptionPrice = proposal.GetPrice().Price
			k.profileKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_DELEGATION_AWARD:
			p := k.delegatingKeeper.GetParams(ctx)
			p.Percentage = proposal.GetDelegationAward().Award
			k.delegatingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_DELEGATION_NETWORK_AWARD:
			p := k.referralKeeper.GetParams(ctx)
			p.DelegatingAward = proposal.GetNetworkAward().Award
			k.referralKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_PRODUCT_NETWORK_AWARD:
			p := k.referralKeeper.GetParams(ctx)
			p.SubscriptionAward = proposal.GetNetworkAward().Award
			k.referralKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_GOVERNMENT_ADD:
			k.AddGovernor(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_GOVERNMENT_REMOVE:
			k.RemoveGovernor(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_PRODUCT_VPN_BASE_PRICE:
			p := k.profileKeeper.GetParams(ctx)
			p.VpnGbPrice = proposal.GetPrice().Price
			k.profileKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_PRODUCT_STORAGE_BASE_PRICE:
			p := k.profileKeeper.GetParams(ctx)
			p.StorageGbPrice = proposal.GetPrice().Price
			k.profileKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_FREE_CREATOR_ADD:
			k.profileKeeper.AddFreeCreator(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_FREE_CREATOR_REMOVE:
			k.profileKeeper.RemoveFreeCreator(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_SOFTWARE_UPGRADE:
			p := proposal.GetSoftwareUpgrade()
			err = k.upgradeKeeper.ScheduleUpgrade(ctx, upgrade.Plan{
				Name:   p.Name,
				Time:   time.Time{},
				Height: p.Height,
				Info:   p.Info,
			})
		case types.PROPOSAL_TYPE_CANCEL_SOFTWARE_UPGRADE:
			k.upgradeKeeper.ClearUpgradePlan(ctx)
		case types.PROPOSAL_TYPE_STAFF_VALIDATOR_ADD:
			err = k.nodingKeeper.AddToStaff(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_STAFF_VALIDATOR_REMOVE:
			err = k.nodingKeeper.RemoveFromStaff(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_EARNING_SIGNER_ADD:
			k.earningKeeper.AddSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_EARNING_SIGNER_REMOVE:
			k.earningKeeper.RemoveSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_TOKEN_RATE_SIGNER_ADD:
			k.profileKeeper.AddTokenRateSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_TOKEN_RATE_SIGNER_REMOVE:
			k.profileKeeper.RemoveTokenRateSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_VPN_SIGNER_ADD:
			k.profileKeeper.AddVpnCurrentSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_VPN_SIGNER_REMOVE:
			k.profileKeeper.RemoveVpnCurrentSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_STORAGE_SIGNER_ADD:
			k.profileKeeper.AddStorageCurrentSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_STORAGE_SIGNER_REMOVE:
			k.profileKeeper.RemoveStorageCurrentSigner(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_TRANSITION_PRICE:
			p := k.referralKeeper.GetParams(ctx)
			p.TransitionPrice = uint64(proposal.GetPrice().Price)
			k.referralKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_MIN_SEND:
			p := k.bankKeeper.GetParams(ctx)
			p.MinSend = proposal.GetMinAmount().MinAmount
			k.bankKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_MIN_DELEGATE:
			p := k.delegatingKeeper.GetParams(ctx)
			p.MinDelegate = proposal.GetMinAmount().MinAmount
			k.delegatingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_MAX_VALIDATORS:
			p := k.nodingKeeper.GetParams(ctx)
			p.MaxValidators = proposal.GetCount().Count
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_GENERAL_AMNESTY:
			k.nodingKeeper.GeneralAmnesty(ctx)
		case types.PROPOSAL_TYPE_LUCKY_VALIDATORS:
			p := k.nodingKeeper.GetParams(ctx)
			p.LotteryValidators = proposal.GetCount().Count
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_VALIDATOR_MINIMAL_STATUS:
			p := k.nodingKeeper.GetParams(ctx)
			p.MinStatus = proposal.GetStatus().Status
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_JAIL_AFTER:
			p := k.nodingKeeper.GetParams(ctx)
			p.JailAfter = proposal.GetCount().Count
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_REVOKE_PERIOD:
			p := k.delegatingKeeper.GetParams(ctx)
			p.RevokePeriod = proposal.GetPeriod().Days
			k.delegatingKeeper.SetParams(ctx, p)
		}
		if err != nil {
			k.Logger(ctx).Error("could not apply voting result due to error",
				"name", proposal.Name,
				"error", err,
			)
		}
	}
}

func (k Keeper) ScheduleEnding(ctx sdk.Context, time time.Time) {
	k.scheduleKeeper.ScheduleTask(ctx, time, types.HookName, nil)
}

func (k Keeper) ProcessSchedule(ctx sdk.Context, _ []byte, _ time.Time) {
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
		if err := proto.Unmarshal(iterator.Value(), &record); err != nil {
			panic(err)
		}
		records = append(records, record)
	}

	return records
}

func (k Keeper) Propose(ctx sdk.Context, msg types.MsgPropose) error {
	if k.GetCurrentProposal(ctx) != nil {
		return types.ErrOtherActive
	}

	var (
		proposal = msg.Proposal
		gov      = k.GetGovernment(ctx)
	)
	if !gov.Contains(proposal.GetAuthor()) {
		return errors.Wrap(types.ErrSignerNotAllowed, msg.Proposal.Author)
	}

	params := k.GetParams(ctx)
	endTime := ctx.BlockTime().Add(time.Duration(params.VotingPeriod) * time.Hour)
	proposal.EndTime = &endTime

	// Set proposal
	k.SetCurrentProposal(ctx, proposal)

	// Set empty lists of voters
	agreed, disagreed := types.Government{Members: []string{proposal.Author}}, types.Government{}
	k.SetAgreed(ctx, agreed)
	k.SetDisagreed(ctx, disagreed)
	k.ScheduleEnding(ctx, endTime)
	k.SetStartBlock(ctx)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventProposalCreated{
			Name:   proposal.Name,
			Author: proposal.Author,
			Type:   proposal.Type,
		},
	); err != nil { panic(err) }

	if complete, agree := k.Validate(gov, agreed, disagreed); complete {
		k.EndProposal(ctx, proposal, agree)
	}
	return nil
}

func (k Keeper) Vote(ctx sdk.Context, voter sdk.AccAddress, agree bool) error {
	proposal := k.GetCurrentProposal(ctx)
	if proposal == nil {
		return types.ErrNoActiveProposals
	}

	gov := k.GetGovernment(ctx)
	if !gov.Contains(voter) {
		return errors.Wrap(types.ErrSignerNotAllowed, voter.String())
	}

	agreed := k.GetAgreed(ctx)
	if agreed.Contains(voter) {
		return errors.Wrap(types.ErrAlreadyVoted, voter.String())
	}

	disagreed := k.GetDisagreed(ctx)
	if disagreed.Contains(voter) {
		return sdkerrors.Wrap(types.ErrAlreadyVoted, voter.String())
	}

	if agree {
		agreed.Append(voter)
		k.SetAgreed(ctx, agreed)
	} else {
		disagreed.Append(voter)
		k.SetDisagreed(ctx, disagreed)
	}

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventProposalVote{
			Voter:  voter.String(),
			Agreed: agree,
		},
	); err != nil { panic(err) }

	if complete, agree := k.Validate(gov, agreed, disagreed); complete {
		k.EndProposal(ctx, *proposal, agree)
	}
	return nil
}
