package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/arterynetwork/artr/util"
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
	k.scheduleKeeper.DeleteAll(ctx, *proposal.EndTime, types.VoteHookName)

	store := ctx.KVStore(k.storeKey)

	// Save proposal data to history
	k.SaveProposalToHistory(ctx, store)

	// Delete all proposal info
	store.Delete(types.KeyCurrentVote)
	store.Delete(types.KeyAgreedMembers)
	store.Delete(types.KeyDisagreedMembers)
	store.Delete(types.KeyStartBlock)

	util.EmitEvent(ctx,
		&types.EventVotingFinished{
			Name:   proposal.Name,
			Agreed: agreed,
		},
	)

	if agreed {
		var err error
		switch proposal.Type {
		case types.PROPOSAL_TYPE_ENTER_PRICE:
			p := k.profileKeeper.GetParams(ctx)
			p.SubscriptionPrice = proposal.GetPrice().Price
			k.profileKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_DELEGATION_AWARD:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_DELEGATION_NETWORK_AWARD:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_PRODUCT_NETWORK_AWARD:
			err = errors.New("parameter is deprecated")
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
			plan := upgrade.Plan{
				Name: p.Name,
				Info: p.Info,
			}
			if p.Time != nil {
				plan.Time = *p.Time
			}
			err = k.upgradeKeeper.ScheduleUpgrade(ctx, plan)
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
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_VALIDATOR_MINIMAL_CRITERIA:
			p := k.nodingKeeper.GetParams(ctx)
			p.MinCriteria = *proposal.GetMinCriteria().MinCriteria
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_JAIL_AFTER:
			p := k.nodingKeeper.GetParams(ctx)
			p.JailAfter = proposal.GetCount().Count
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_REVOKE_PERIOD:
			p := k.delegatingKeeper.GetParams(ctx)
			p.RevokePeriod = proposal.GetPeriod().Days
			k.delegatingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_DUST_DELEGATION:
			p := k.bankKeeper.GetParams(ctx)
			p.DustDelegation = proposal.GetMinAmount().MinAmount
			k.bankKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_VOTING_POWER:
			p := k.nodingKeeper.GetParams(ctx)
			p.VotingPower = *proposal.GetVotingPower()
			k.nodingKeeper.SetParams(ctx, p)
		case types.PROPOSAL_TYPE_VALIDATOR_BONUS:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_SUBSCRIPTION_BONUS:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_VPN_BONUS:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_STORAGE_BONUS:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_VALIDATOR:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_TRANSACTION_FEE:
			p := k.bankKeeper.GetParams(ctx)
			p.TransactionFee = proposal.GetPortion().Fraction
			if err = p.Validate(); err == nil {
				k.bankKeeper.SetParams(ctx, p)
			}
		case types.PROPOSAL_TYPE_BURN_ON_REVOKE:
			p := k.delegatingKeeper.GetParams(ctx)
			p.BurnOnRevoke = proposal.GetPortion().Fraction
			if err = p.Validate(); err == nil {
				k.delegatingKeeper.SetParams(ctx, p)
			}
		case types.PROPOSAL_TYPE_MAX_TRANSACTION_FEE:
			p := k.bankKeeper.GetParams(ctx)
			p.MaxTransactionFee = proposal.GetMinAmount().MinAmount
			if err = p.Validate(); err == nil {
				k.bankKeeper.SetParams(ctx, p)
			}
		case types.PROPOSAL_TYPE_TRANSACTION_FEE_SPLIT_RATIOS:
			p := k.bankKeeper.GetParams(ctx)
			p.TransactionFeeSplitRatios.ForProposer = proposal.GetPortions().Fractions[0]
			p.TransactionFeeSplitRatios.ForCompany = proposal.GetPortions().Fractions[1]
			if err = p.Validate(); err == nil {
				k.bankKeeper.SetParams(ctx, p)
			}
		case types.PROPOSAL_TYPE_ACCRUE_PERCENTAGE_RANGES:
			err = errors.New("parameter is deprecated")
		case types.PROPOSAL_TYPE_ACCRUE_PERCENTAGE_TABLE:
			p := k.delegatingKeeper.GetParams(ctx)
			p.AccruePercentageTable = proposal.GetAccruePercentageTable().AccruePercentageTable
			if err = p.Validate(); err == nil {
				k.delegatingKeeper.SetParams(ctx, p)
			}
		case types.PROPOSAL_TYPE_BLOCKED_SENDER_ADD:
			k.bankKeeper.AddBlockedSender(ctx, proposal.GetAddress().GetAddress())
		case types.PROPOSAL_TYPE_BLOCKED_SENDER_REMOVE:
			k.bankKeeper.RemoveBlockedSender(ctx, proposal.GetAddress().GetAddress())
		default:
			err = errors.Errorf("unknown proposal type %d", proposal.Type)
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
	k.scheduleKeeper.ScheduleTask(ctx, time, types.VoteHookName, nil)
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

	util.EmitEvent(ctx,
		&types.EventProposalCreated{
			Name:   proposal.Name,
			Author: proposal.Author,
			Type:   proposal.Type,
		},
	)

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

	util.EmitEvent(ctx,
		&types.EventProposalVote{
			Voter:  voter.String(),
			Agreed: agree,
		},
	)

	if complete, agree := k.Validate(gov, agreed, disagreed); complete {
		k.EndProposal(ctx, *proposal, agree)
	}
	return nil
}

func (k Keeper) GetCurrentPoll(ctx sdk.Context) (poll types.Poll, ok bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)
	bz := store.Get(types.KeyPollCurrent)

	if bz == nil {
		return types.Poll{}, false
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &poll)
	return poll, true
}

func (k Keeper) GetPollStatus(ctx sdk.Context) (yes, no uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)
	if bz := store.Get(types.KeyPollYesCount); bz != nil {
		yes = binary.BigEndian.Uint64(bz)
	}
	if bz := store.Get(types.KeyPollNoCount); bz != nil {
		no = binary.BigEndian.Uint64(bz)
	}
	return
}

func (k Keeper) StartPoll(ctx sdk.Context, poll types.Poll) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)

	if store.Has(types.KeyPollCurrent) {
		return types.ErrOtherActive
	}
	if !util.ContainsString(k.GetGovernment(ctx).Strings(), poll.Author) {
		return types.ErrSignerNotAllowed
	}

	start := ctx.BlockTime()
	end := start.Add(time.Duration(k.GetParams(ctx).PollPeriod) * time.Hour)
	k.scheduleKeeper.ScheduleTask(ctx, end, types.PollHookName, nil)
	poll.StartTime = &start
	poll.EndTime = &end

	store.Set(types.KeyPollCurrent, k.cdc.MustMarshalBinaryBare(&poll))
	return nil
}

func (k Keeper) Answer(ctx sdk.Context, acc string, yes bool) error {
	poll, ok := k.GetCurrentPoll(ctx)
	if !ok {
		return types.ErrNoActivePoll
	}

	addr, err := sdk.AccAddressFromBech32(acc)
	if err != nil {
		panic(errors.Wrap(err, "cannot parse acc address"))
	}

	switch r := poll.Requirements.(type) {
	case *types.Poll_CanValidate:
		q, _, _, err := k.nodingKeeper.IsQualified(ctx, addr)
		if err != nil {
			panic(errors.Wrap(err, "cannot check for qualification"))
		}
		if !q {
			return types.ErrRespondentNotAllowed
		}
	case *types.Poll_MinStatus:
		info, err := k.referralKeeper.Get(ctx, acc)
		if err != nil {
			panic(errors.Wrap(err, "cannot obtain referral info"))
		}
		if info.Status < r.MinStatus {
			return types.ErrRespondentNotAllowed
		}
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)
	ansStore := prefix.NewStore(store, types.KeyPollAnswers)
	key := []byte(acc)
	if ansStore.Has(key) {
		return types.ErrAlreadyVoted
	}

	var ans, countKey []byte
	if yes {
		ans = types.ValueYes
		countKey = types.KeyPollYesCount
	} else {
		ans = types.ValueNo
		countKey = types.KeyPollNoCount
	}
	ansStore.Set(key, ans)

	var (
		bz    []byte
		value uint64
	)
	if bz = store.Get(countKey); bz != nil {
		value = binary.BigEndian.Uint64(bz)
	} else {
		bz = make([]byte, 8)
	}
	value += 1
	binary.BigEndian.PutUint64(bz, value)
	store.Set(countKey, bz)

	return nil
}

func (k Keeper) EndPollHandler(ctx sdk.Context, _ []byte, _ time.Time) { k.EndPoll(ctx) }
func (k Keeper) EndPoll(ctx sdk.Context) {
	store := cachekv.NewStore(prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix))

	var (
		poll     types.Poll
		yes, no  uint64
		decision types.Decision
	)
	if bz := store.Get(types.KeyPollCurrent); bz != nil {
		k.cdc.MustUnmarshalBinaryBare(bz, &poll)
	} else {
		panic(types.ErrNoActivePoll)
	}
	if bz := store.Get(types.KeyPollYesCount); bz != nil {
		yes = binary.BigEndian.Uint64(bz)
	}
	if bz := store.Get(types.KeyPollNoCount); bz != nil {
		no = binary.BigEndian.Uint64(bz)
	}
	if poll.Quorum != nil {
		if yes != 0 && util.FractionInt(int64(yes)).GTE(poll.Quorum.Mul(util.FractionInt(int64(yes+no)))) {
			decision = types.DECISION_POSITIVE
		} else {
			decision = types.DECISION_NEGATIVE
		}
	}

	util.EmitEvent(ctx,
		&types.EventPollFinished{
			Name:     poll.Name,
			Yes:      yes,
			No:       no,
			Decision: decision,
		},
	)

	historyKey := make([]byte, len(types.KeyPollHistory)+8)
	copy(historyKey, types.KeyPollHistory)
	binary.BigEndian.PutUint64(historyKey[len(types.KeyPollHistory):], uint64(poll.EndTime.Unix()))
	store.Set(historyKey, k.cdc.MustMarshalBinaryBare(&types.PollHistoryItem{
		Poll:     poll,
		Yes:      yes,
		No:       no,
		Decision: decision,
	}))

	store.Delete(types.KeyPollCurrent)
	store.Delete(types.KeyPollYesCount)
	store.Delete(types.KeyPollNoCount)
	it := sdk.KVStorePrefixIterator(store, types.KeyPollAnswers)
	for ; it.Valid(); it.Next() {
		store.Delete(it.Key())
	}
	it.Close()

	store.Write()
}

func (k Keeper) GetPollHistoryAll(ctx sdk.Context) []types.PollHistoryItem {
	return k.GetPollHistory(ctx, 0, 0)
}
func (k Keeper) GetPollHistory(ctx sdk.Context, limit int32, page int32) []types.PollHistoryItem {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)
	var (
		it  sdk.Iterator
		res []types.PollHistoryItem
	)
	if limit > 0 {
		it = sdk.KVStorePrefixIteratorPaginated(store, types.KeyPollHistory, uint(page), uint(limit))
		res = make([]types.PollHistoryItem, 0, limit)
	} else {
		it = sdk.KVStorePrefixIterator(store, types.KeyPollHistory)
	}
	for ; it.Valid(); it.Next() {
		var item types.PollHistoryItem
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &item)
		res = append(res, item)
	}
	it.Close()
	return res
}

func (k Keeper) IterateThroughCurrentPollAnswers(ctx sdk.Context, callback func(acc string, ans bool) (stop bool)) (err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)
	if !store.Has(types.KeyPollCurrent) {
		return types.ErrNoActivePoll
	}

	it := sdk.KVStorePrefixIterator(store, types.KeyPollAnswers)
	defer func() {
		it.Close()
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = errors.Wrap(er, "callback paniced")
			} else {
				err = errors.Errorf("callback paniced: %s", er)
			}
		}
	}()

	for ; it.Valid(); it.Next() {
		acc := string(it.Key()[len(types.KeyPollAnswers):])
		ans := bytes.Equal(it.Value(), types.ValueYes)
		if stop := callback(acc, ans); stop {
			return nil
		}
	}
	return nil
}

func (k Keeper) LoadPolls(ctx sdk.Context, state types.GenesisState) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPollPrefix)

	if state.CurrentPoll != nil {
		store.Set(types.KeyPollCurrent, k.cdc.MustMarshalBinaryBare(state.CurrentPoll))
		for _, ans := range state.PollAnswers {
			if err := k.Answer(ctx, ans.Acc, ans.Ans); err != nil {
				panic(err)
			}
		}
	}

	key := make([]byte, len(types.KeyPollHistory)+8)
	copy(key, types.KeyPollHistory)
	for _, item := range state.PollHistory {
		binary.BigEndian.PutUint64(key[len(types.KeyPollHistory):], uint64(item.Poll.EndTime.Unix()))
		store.Set(key, k.cdc.MustMarshalBinaryBare(&item))
	}
}
