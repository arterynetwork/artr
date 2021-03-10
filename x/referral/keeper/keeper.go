package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/referral/types"
)

const (
	// CompressionPeriod - amount of time between switching status off and account compression
	CompressionPeriod = 2 * util.BlocksOneMonth

	minIndexedStatus = types.Businessman
)

// Keeper of the referral store
type Keeper struct {
	storeKey       sdk.StoreKey
	indexStoreKey  sdk.StoreKey
	cdc            *codec.Codec
	paramspace     types.ParamSubspace
	accKeeper      types.AccountKeeper
	scheduleKeeper types.ScheduleKeeper
	bankKeeper     types.BankKeeper
	supplyKeeper   types.SupplyKeeper
	eventHooks     map[string][]func(ctx sdk.Context, acc sdk.AccAddress) error
}

// NewKeeper creates a referral keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, idxKey sdk.StoreKey, paramspace types.ParamSubspace,
	accKeeper types.AccountKeeper, scheduleKeeper types.ScheduleKeeper, bankKeeper types.BankKeeper,
	supplyKeeper types.SupplyKeeper,
) Keeper {
	keeper := Keeper{
		storeKey:       key,
		indexStoreKey:  idxKey,
		cdc:            cdc,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		accKeeper:      accKeeper,
		scheduleKeeper: scheduleKeeper,
		bankKeeper:     bankKeeper,
		supplyKeeper:   supplyKeeper,
		eventHooks:     make(map[string][]func(ctx sdk.Context, acc sdk.AccAddress) error),
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetStatus returns a status for an account (i.e. lvl 1 "Lucky", lvl 2 "Leader", lvl 3 "Master" or so on)
func (k Keeper) GetStatus(ctx sdk.Context, acc sdk.AccAddress) (types.Status, error) {
	data, err := k.get(ctx, acc)
	if err != nil {
		return 0, err
	}
	return data.Status, nil
}

// GetParent returns a parent for an account
func (k Keeper) GetParent(ctx sdk.Context, acc sdk.AccAddress) (sdk.AccAddress, error) {
	data, err := k.get(ctx, acc)
	if err != nil {
		return nil, err
	}
	return data.Referrer, nil
}

// GetChildren returns children (1st line only) for an account
func (k Keeper) GetChildren(ctx sdk.Context, acc sdk.AccAddress) ([]sdk.AccAddress, error) {
	data, err := k.get(ctx, acc)
	if err != nil {
		return nil, err
	}
	result := make([]sdk.AccAddress, 0, len(data.Referrals))
	for _, child := range data.Referrals {
		result = append(result, child)
	}
	return result, nil
}

// GetReferralFeesForSubscription returns a set of account-ratio pairs, describing what part of monthly subscription
// should go to what wallet. 0.85 total. The rest goes for validator and leader bonuses.
func (k Keeper) GetReferralFeesForSubscription(ctx sdk.Context, acc sdk.AccAddress) ([]types.ReferralFee, error) {
	var params types.Params
	k.paramspace.GetParamSet(ctx, &params)
	ca := params.CompanyAccounts

	fees, err := k.getReferralFeesCore(
		ctx,
		acc,
		ca.ForSubscription,
		params.SubscriptionAward.Company,
		params.SubscriptionAward.Network,
		ca.TopReferrer,
	)
	return append(fees,
		types.ReferralFee{
			Beneficiary: ca.PromoBonuses,
			Ratio:       util.Percent(5),
		},
		types.ReferralFee{
			Beneficiary: ca.StatusBonuses,
			Ratio:       util.Percent(5),
		},
		types.ReferralFee{
			Beneficiary: ca.LeaderBonuses,
			Ratio:       util.Percent(5),
		},
	), err
}

// GetReferralFeesForDelegating returns a set of account-ratio pairs, describing what part of being delegated funds
// should go to what wallet. 0.15 total. The rest should be frozen at the account's special wallet.
func (k Keeper) GetReferralFeesForDelegating(ctx sdk.Context, acc sdk.AccAddress) ([]types.ReferralFee, error) {
	var params types.Params
	k.paramspace.GetParamSet(ctx, &params)
	ca := params.CompanyAccounts

	return k.getReferralFeesCore(
		ctx,
		acc,
		ca.ForDelegating,
		params.DelegatingAward.Company,
		params.DelegatingAward.Network,
		ca.TopReferrer,
	)
}

// AreStatusRequirementsFulfilled validates if the account suffices the status requirement.
// The actual account status doesn't matter and won't be updated.
func (k Keeper) AreStatusRequirementsFulfilled(ctx sdk.Context, acc sdk.AccAddress, s types.Status) (types.StatusCheckResult, error) {
	if s < types.MinimumStatus || s > types.MaximumStatus {
		return types.StatusCheckResult{Overall: false}, fmt.Errorf("there is no such status: %d", s)
	}
	data, err := k.get(ctx, acc)
	if err != nil {
		return types.StatusCheckResult{Overall: false}, err
	}
	return statusRequirements[s](data, newBunchUpdater(k, ctx))
}

// AddTopLevelAccount adds accounts without parent and is supposed to be used during genesis
func (k Keeper) AddTopLevelAccount(ctx sdk.Context, acc sdk.AccAddress) error {
	if k.exists(ctx, acc) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("account %s already exists", acc.String()),
		)
	}
	var (
		bu        = newBunchUpdater(k, ctx)
		coins     = k.getBalance(ctx, acc)
		delegated = k.getDelegated(ctx, acc)
	)
	newItem := types.NewR(nil, coins, delegated)
	if err := bu.set(acc, newItem); err != nil {
		return err
	}
	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

// GetTopLevelAccounts returns all accounts without parents and is supposed to be used during genesis export
func (k Keeper) GetTopLevelAccounts(ctx sdk.Context) ([]sdk.AccAddress, error) {
	var res []sdk.AccAddress
	store := ctx.KVStore(k.storeKey)
	itr := store.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		v := itr.Value()
		var record types.R
		err := k.cdc.UnmarshalBinaryLengthPrefixed(v, &record)
		if err != nil {
			return nil, err
		}
		if record.Referrer == nil {
			res = append(res, sdk.AccAddress(itr.Key()))
		}
	}
	return res, nil
}

// AppendChild adds a new account to the referral structure. The parent account should already exist and the child one
// should not.
func (k Keeper) AppendChild(ctx sdk.Context, parentAcc sdk.AccAddress, childAcc sdk.AccAddress) error {
	if parentAcc == nil {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"parentAcc cannot be nil",
		)
	}
	if k.exists(ctx, childAcc) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("account %s already exists", childAcc.String()),
		)
	}
	var (
		bu        = newBunchUpdater(k, ctx)
		anc       = parentAcc
		coins     = k.getBalance(ctx, childAcc)
		delegated = k.getDelegated(ctx, childAcc)
	)
	newItem := types.NewR(parentAcc, coins, delegated)
	newItem.CompressionAt = ctx.BlockHeight() + CompressionPeriod
	err := bu.set(childAcc, newItem)
	if err != nil {
		return sdkerrors.Wrap(err, "cannot set "+childAcc.String())
	}
	for i := 0; i < 10; i++ {
		if anc == nil {
			break
		}
		err = bu.update(anc, true, func(value *types.R) {
			value.Coins[i+1] = value.Coins[i+1].Add(coins)
			value.Delegated[i+1] = value.Delegated[i+1].Add(delegated)
			bu.addCallback(StakeChangedCallback, anc)
			if i == 0 {
				value.Referrals = append(value.Referrals, childAcc)
			}
			anc = value.Referrer
		})
		if err != nil {
			return sdkerrors.Wrap(err, "cannot update "+anc.String())
		}
	}

	if err := bu.commit(); err != nil {
		return sdkerrors.Wrap(err, "cannot commit")
	}
	return nil
}

// Compress relocates all account's children under its parent, so the account looses its entire network.
func (k Keeper) Compress(ctx sdk.Context, acc sdk.AccAddress) error {
	var (
		bu         = newBunchUpdater(k, ctx)
		childrenSb = strings.Builder{}

		coins     [11]sdk.Int
		delegated [11]sdk.Int
		children  []sdk.AccAddress
		refsCount [11]int
		parent    sdk.AccAddress
	)
	// Compressed account itself:
	//   * no referrals
	//   * no coins (neither delegated nor free)
	//   * status dump
	//   * shorten legs
	//   * no own children
	//   * new compression in a time
	compressionAt := ctx.BlockHeight() + CompressionPeriod

	err := bu.update(acc, false, func(value *types.R) {
		children = value.Referrals
		coins = value.Coins
		delegated = value.Delegated
		parent = value.Referrer
		refsCount = value.ActiveReferralsCount

		value.Referrals = nil
		value.ActiveReferralsCount = [11]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		value.Coins = [11]sdk.Int{
			coins[0],
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
		}
		value.Delegated = [11]sdk.Int{
			delegated[0],
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
		}
		value.CompressionAt = compressionAt
		bu.addCallback(StakeChangedCallback, acc)
		k.setStatus(ctx, value, types.Lucky, acc)
		bu.addCallback(StatusUpdatedCallback, acc)
	})
	if err != nil {
		return err
	}

	err = k.ScheduleCompression(ctx, acc, compressionAt)
	if err != nil {
		return err
	}

	// Children: just new referrer
	for _, acc := range children {
		err = bu.update(acc, false, func(value *types.R) {
			value.Referrer = parent
		})
		if err != nil {
			return err
		}
		if _, err = childrenSb.WriteString(acc.String() + ","); err != nil {
			return err
		}
	}
	childrenStr := childrenSb.String()
	if len(childrenStr) > 0 {
		childrenStr = childrenStr[:len(childrenStr)-1]
	}

	// Ancestors (level k, 1 <= k <= 10):
	//   * coins[i] pop from level k+i to level k+i-1 (, for 0 < i < 11-k)
	//   * coins[11-k] appears at level 10
	//   * extend leg (as a distance shrinks, new nodes might appear in 10-lvl-radius)
	// Parent (k = 1) only:
	//   * new referrals
	for k, ancestor := 1, parent; k <= 10 && ancestor != nil; k++ {
		err = bu.update(ancestor, true, func(value *types.R) {
			bu.addCallback(StakeChangedCallback, ancestor)
			ancestor = value.Referrer
			value.Coins[k] = value.Coins[k].Add(coins[1])
			value.Delegated[k] = value.Delegated[k].Add(delegated[1])
			value.ActiveReferralsCount[k] += refsCount[1]
			for i := 1; i < 10-k; i++ {
				value.Coins[k+i] = value.Coins[k+i].Add(coins[i+1]).Sub(coins[i])
				value.Delegated[k+i] = value.Delegated[k+i].Add(delegated[i+1]).Sub(delegated[i])
				value.ActiveReferralsCount[k+i] += refsCount[i+1] - refsCount[i]
			}
			if k == 1 {
				value.Referrals = append(value.Referrals, children...)
			}
		})
		if err != nil {
			return err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompression,
			sdk.NewAttribute(types.AttributeKeyAddress, acc.String()),
			sdk.NewAttribute(types.AttributeKeyReferrer, parent.String()),
			sdk.NewAttribute(types.AttributeKeyReferrals, childrenStr),
		),
	)

	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

// GetCoinsInNetwork returns total amount of coins (delegated and not) in a person's network
// (at levels that are open according the person's current status, but no deeper than `maxDepth` levels down).
// Own coins inclusive.
func (k Keeper) GetCoinsInNetwork(ctx sdk.Context, acc sdk.AccAddress, maxDepth int) (sdk.Int, error) {
	data, err := k.get(ctx, acc)
	if err != nil {
		return sdk.Int{}, err
	}
	d := data.Status.LinesOpened()
	if d > maxDepth {
		d = maxDepth
	}
	return data.CoinsAtLevelsUpTo(d), nil
}

// GetDelegatedInNetwork returns total amount of delegated coins in a person's network
// (at levels that are open according the person's current status, but no deeper than `maxDepth` levels down).
// Own coins inclusive.
func (k Keeper) GetDelegatedInNetwork(ctx sdk.Context, acc sdk.AccAddress, maxDepth int) (sdk.Int, error) {
	data, err := k.get(ctx, acc)
	if err != nil {
		return sdk.Int{}, err
	}
	d := data.Status.LinesOpened()
	if d > maxDepth {
		d = maxDepth
	}
	return data.DelegatedAtLevelsUpTo(d), nil
}

func (k Keeper) OnBalanceChanged(ctx sdk.Context, acc sdk.AccAddress) error {
	k.Logger(ctx).Debug("OnBalanceChanged", "acc", acc)
	var (
		bu = newBunchUpdater(k, ctx)

		dc, dd sdk.Int
		node   sdk.AccAddress
	)
	err := bu.update(acc, true, func(value *types.R) {
		newBalance := k.getBalance(ctx, acc)
		newDelegated := k.getDelegated(ctx, acc)

		dc = newBalance.Sub(value.Coins[0])
		dd = newDelegated.Sub(value.Delegated[0])
		if !dd.IsZero() {
			bu.addCallback(StakeChangedCallback, acc)
		}
		node = value.Referrer

		value.Coins[0] = newBalance
		value.Delegated[0] = newDelegated
	})
	if err != nil {
		k.Logger(ctx).Error("OnBalanceChanged hook failed", "step", 0, "error", err)
		return err
	}

	for i := 1; i <= 10; i++ {
		if node == nil {
			break
		}
		err = bu.update(node, true, func(value *types.R) {
			value.Coins[i] = value.Coins[i].Add(dc)
			value.Delegated[i] = value.Delegated[i].Add(dd)
			if !dd.IsZero() {
				bu.addCallback(StakeChangedCallback, node)
			}

			node = value.Referrer
		})
		if err != nil {
			k.Logger(ctx).Error("OnBalanceChanged hook failed", "step", i, "error", err)
			return err
		}
	}

	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

func (k Keeper) SetActive(ctx sdk.Context, acc sdk.AccAddress, value bool) error {
	var (
		bu                = newBunchUpdater(k, ctx)
		valueIsAlreadySet = false

		parent        sdk.AccAddress
		d             int
		compressionAt int64
	)
	if value {
		d = 1
		compressionAt = -1
	} else {
		d = -1
		compressionAt = ctx.BlockHeight() + CompressionPeriod
	}

	err := bu.update(acc, false, func(x *types.R) {
		if x.Active == value {
			valueIsAlreadySet = true
		} else {
			x.Active = value
			x.ActiveReferralsCount[0] += d
			x.CompressionAt = compressionAt
			parent = x.Referrer
		}
	})
	if err != nil {
		return err
	} else if valueIsAlreadySet {
		return nil
	}

	for i := 0; i < 10; i++ {
		if parent == nil {
			break
		}
		err = bu.update(parent, true, func(x *types.R) {
			x.ActiveReferralsCount[i+1] += d
			parent = x.Referrer
		})
		if err != nil {
			return err
		}
	}

	if !value && !valueIsAlreadySet {
		if err = k.ScheduleCompression(ctx, acc, ctx.BlockHeight()+CompressionPeriod); err != nil {
			return err
		}
	}

	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

func (k Keeper) PayStatusBonus(ctx sdk.Context) error {
	if ctx.BlockHeight() <= k.scheduleKeeper.GetParams(ctx).InitialHeight {
		return nil
	}
	var (
		ca     = k.GetParams(ctx).CompanyAccounts
		sender = ca.StatusBonuses
		amt    = k.accKeeper.GetAccount(ctx, sender).GetCoins().AmountOf(util.ConfigMainDenom).Int64() / 5
	)
	if amt == 0 {
		return nil
	}
	var (
		store           = ctx.KVStore(k.indexStoreKey)
		receivers       = make([]sdk.AccAddress, 0)
		outMap          = make(map[string]bank.Output)
		total     int64 = 0
	)

	for status := types.AbsoluteChampion; status >= types.Businessman; status-- {
		it := sdk.KVStorePrefixIterator(store, []byte{uint8(status)})
		for ; it.Valid(); it.Next() {
			receivers = append(receivers, sdk.AccAddress(it.Key()[1:]))
		}
		it.Close()
		if len(receivers) == 0 {
			setOrUpdate(outMap, ca.TopReferrer, amt)
			total += amt
		} else {
			n := int64(len(receivers))
			each := amt / n
			if each == 0 {
				break
			}
			total += each * n
			for _, r := range receivers {
				setOrUpdate(outMap, r, each)
			}
		}
	}
	if len(outMap) == 0 {
		return nil
	}

	outputs := make([]bank.Output, 0, len(outMap))
	for _, output := range outMap {
		outputs = append(outputs, output)
	}
	// Map iteration order is not determined :-(
	sort.Slice(outputs, func(i, j int) bool { return bytes.Compare(outputs[i].Address, outputs[j].Address) < 0 })
	for _, out := range outputs {
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeStatusBonus,
			sdk.NewAttribute(types.AttributeKeyAddress, out.Address.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, out.Coins.String()),
		))
	}

	return k.bankKeeper.InputOutputCoins(ctx, []bank.Input{bank.NewInput(sender, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(total))))}, outputs)
}

func (k Keeper) Iterate(ctx sdk.Context, callback func(acc sdk.AccAddress, r *types.R) (changed, checkForStatusUpdate bool)) {
	bu := newBunchUpdater(k, ctx)
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer func() {
		if it != nil {
			it.Close()
		}
	}()
	for ; it.Valid(); it.Next() {
		var acc sdk.AccAddress = it.Key()
		var item types.R
		k.cdc.MustUnmarshalBinaryLengthPrefixed(it.Value(), &item)
		if changed, checkForStatusUpdate := callback(acc, &item); changed || checkForStatusUpdate {
			var f func(r *types.R)
			if changed {
				f = func(r *types.R) { *r = item }
			} else {
				f = func(_ *types.R) {}
			}
			err := bu.update(acc, checkForStatusUpdate, f)
			if err != nil {
				panic(err)
			}
		}
	}
	it.Close()
	it = nil
	err := bu.commit()
	if err != nil {
		panic(err)
	}
}

func (k Keeper) GetCompressionBlockHeight(ctx sdk.Context, acc sdk.AccAddress) (int64, error) {
	info, err := k.get(ctx, acc)
	if err != nil {
		return -1, sdkerrors.Wrap(err, "account not found")
	}
	return info.CompressionAt, nil
}

// RequestTransaction is supposed to be called when a user wants to be moved under another referrer. If the current
// referrer do not approve this operation in a day, it will be cancelled.
func (k Keeper) RequestTransition(ctx sdk.Context, subject, newParent sdk.AccAddress) error {
	var (
		r   types.R
		err error
	)

	if r, err = k.get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}
	if err = k.validateTransition(ctx, subject, newParent, true); err != nil {
		return errors.Wrap(err, "transition is invalid")
	}

	params := k.GetParams(ctx)
	if params.TransitionCost > 0 {
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, subject, auth.FeeCollectorName, util.UartrsUint64(params.TransitionCost))
		if err != nil {
			return errors.Wrap(err, "cannot pay commission")
		}
	}

	r.Transition = newParent
	if err = k.set(ctx, subject, r); err != nil {
		panic(errors.Wrap(err, "cannot write to KVStore"))
	}

	var data []byte = subject
	err = k.scheduleKeeper.ScheduleTask(ctx, uint64(ctx.BlockHeight()+util.BlocksOneDay), TransitionTimeoutHookName, &data)
	if err != nil {
		panic(errors.Wrap(err, "cannot schedule transition timeout"))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransitionRequested,
		sdk.NewAttribute(types.AttributeKeyAddress, subject.String()),
		sdk.NewAttribute(types.AttributeKeyReferrerBefore, r.Referrer.String()),
		sdk.NewAttribute(types.AttributeKeyReferrerAfter, newParent.String()),
	))
	return nil
}

// CancelTransition is supposed to be called when either a current referrer declines a referral transition or this
// transition timeout occurs. See also RequestTransition method.
func (k Keeper) CancelTransition(ctx sdk.Context, subject sdk.AccAddress, timeout bool) error {
	var (
		r   types.R
		err error
	)
	if r, err = k.get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}
	value := r.Transition
	r.Transition = nil
	if err = k.set(ctx, subject, r); err != nil {
		panic(errors.Wrap(err, "cannot write to KVStore"))
	}

	var reason string
	if timeout {
		reason = types.AttributeValueTimeout
	} else {
		reason = types.AttributeValueDeclined
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransitionDeclined,
		sdk.NewAttribute(types.AttributeKeyAddress, subject.String()),
		sdk.NewAttribute(types.AttributeKeyReferrerBefore, r.Referrer.String()),
		sdk.NewAttribute(types.AttributeKeyReferrerAfter, value.String()),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	))
	return nil
}

// AffirmTransition is supposed to be called when a current referrer approves a referral transition. Actual subtree
// relocation and all the according recalculations and updates are done here.
func (k Keeper) AffirmTransition(ctx sdk.Context, subject sdk.AccAddress) error {
	var (
		r   types.R
		err error
	)

	if r, err = k.get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}

	// we should double-check, just in case something has changed
	if err = k.validateTransition(ctx, subject, r.Transition, false); err != nil {
		return errors.Wrap(err, "transition is invalid")
	}

	oldParent, newParent := r.Referrer, r.Transition
	r.Referrer, r.Transition = newParent, nil
	if err = k.set(ctx, subject, r); err != nil {
		panic(errors.Wrap(err, "cannot write to KVStore"))
	}

	var (
		bu                       = newBunchUpdater(k, ctx)
		oldAncestor, newAncestor sdk.AccAddress
	)

	if err = bu.update(oldParent, true, func(value *types.R) {
		idx := 0
		for ; idx < len(value.Referrals) && !value.Referrals[idx].Equals(subject); idx++ {
		}
		value.Referrals[idx] = value.Referrals[len(value.Referrals)-1]
		value.Referrals = value.Referrals[:len(value.Referrals)-1]

		for i := 1; i <= 10; i++ {
			value.Coins[i] = value.Coins[i].Sub(r.Coins[i-1])
			value.Delegated[i] = value.Delegated[i].Sub(r.Delegated[i-1])
			value.ActiveReferralsCount[i] -= r.ActiveReferralsCount[i-1]
		}

		oldAncestor = value.Referrer
	}); err != nil {
		panic(errors.Wrap(err, "cannot update old referrer data"))
	}

	if err = bu.update(newParent, true, func(value *types.R) {
		value.Referrals = append(value.Referrals, subject)

		for i := 1; i <= 10; i++ {
			value.Coins[i] = value.Coins[i].Add(r.Coins[i-1])
			value.Delegated[i] = value.Delegated[i].Add(r.Delegated[i-1])
			value.ActiveReferralsCount[i] += r.ActiveReferralsCount[i-1]
		}

		newAncestor = value.Referrer
	}); err != nil {
		panic(errors.Wrap(err, "cannot update new referrer data"))
	}

	for level := 2; level <= 10; level++ {
		if oldAncestor.Equals(newAncestor) {
			break
		}
		if !oldAncestor.Empty() {
			if err = bu.update(oldAncestor, true, func(value *types.R) {
				for i := level; i <= 10; i++ {
					value.Coins[i] = value.Coins[i].Sub(r.Coins[i-level])
					value.Delegated[i] = value.Delegated[i].Sub(r.Delegated[i-level])
					value.ActiveReferralsCount[i] -= r.ActiveReferralsCount[i-level]
				}
				oldAncestor = value.Referrer
			}); err != nil {
				panic(errors.Wrapf(err, "cannot update old level-%d ancestor data", level))
			}
		}
		if !newAncestor.Empty() {
			if err = bu.update(newAncestor, true, func(value *types.R) {
				for i := level; i <= 10; i++ {
					value.Coins[i] = value.Coins[i].Add(r.Coins[i-level])
					value.Delegated[i] = value.Delegated[i].Add(r.Delegated[i-level])
					value.ActiveReferralsCount[i] += r.ActiveReferralsCount[i-level]
				}
				newAncestor = value.Referrer
			}); err != nil {
				panic(errors.Wrapf(err, "cannot update new level-%d ancestor data", level))
			}
		}
	}

	if err = bu.commit(); err != nil {
		panic(errors.Wrap(err, "cannot commit changes"))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransitionPerformed,
		sdk.NewAttribute(types.AttributeKeyAddress, subject.String()),
		sdk.NewAttribute(types.AttributeKeyReferrerBefore, oldParent.String()),
		sdk.NewAttribute(types.AttributeKeyReferrerAfter, newParent.String()),
	))
	return nil
}

// GetPendingTransition returns a new referral that the specified account is requested to be moved under. It returns
// (nil, nil) if the account is OK, but a transition is not requested.
func (k Keeper) GetPendingTransition(ctx sdk.Context, acc sdk.AccAddress) (sdk.AccAddress, error) {
	if acc.Empty() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "account address is missing")
	}
	r, err := k.get(ctx, acc)
	if err != nil {
		return nil, err
	}
	return r.Transition, nil
}

// -------------- PRIVATE FUNCTIONS --------------------

// get returns all the data for an account (status, parent, children)
func (k Keeper) get(ctx sdk.Context, acc sdk.AccAddress) (types.R, error) {
	store := ctx.KVStore(k.storeKey)
	var item types.R
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get([]byte(acc)), &item)
	return item, err
}

func (k Keeper) getReferralFeesCore(ctx sdk.Context, acc sdk.AccAddress, companyAccount sdk.AccAddress, toCompany util.Fraction, toAncestors [10]util.Fraction, topReferrer sdk.AccAddress) ([]types.ReferralFee, error) {
	excess := util.Percent(0)
	result := append(make([]types.ReferralFee, 0, 12), types.ReferralFee{Beneficiary: companyAccount, Ratio: toCompany})

	ancestor, err := k.GetParent(ctx, acc)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 10; i++ {
		var (
			data types.R
			err  error
		)
		for {
			if ancestor == nil {
				break
			}
			data, err = k.get(ctx, ancestor)
			if err != nil {
				return nil, err
			}
			if data.Active {
				break
			} else {
				ancestor = data.Referrer
			}
		}
		if ancestor == nil {
			excess = excess.Add(toAncestors[i])
			continue
		}
		if i < data.Status.LinesOpened() {
			result = append(result, types.ReferralFee{Beneficiary: ancestor, Ratio: toAncestors[i]})
		} else {
			excess = excess.Add(toAncestors[i])
		}
		ancestor = data.Referrer
	}
	if !excess.IsZero() {
		result = append(result, types.ReferralFee{Beneficiary: topReferrer, Ratio: excess})
	}
	return result, nil
}

func (k Keeper) set(ctx sdk.Context, acc sdk.AccAddress, value types.R) error {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := k.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)

	return nil
}

func (k Keeper) update(ctx sdk.Context, acc sdk.AccAddress, callback func(value types.R) types.R) error {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(acc)
	var value types.R
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(keyBytes), &value)
	if err != nil {
		return err
	}
	value = callback(value)
	valueBytes, err := k.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (k Keeper) getBalance(ctx sdk.Context, acc sdk.AccAddress) sdk.Int {
	coins := k.accKeeper.GetAccount(ctx, acc).GetCoins()
	return coins.AmountOf(util.ConfigMainDenom).
		Add(coins.AmountOf(util.ConfigDelegatedDenom)).
		Add(coins.AmountOf(util.ConfigRevokingDenom))
}

func (k Keeper) getDelegated(ctx sdk.Context, acc sdk.AccAddress) sdk.Int {
	return k.accKeeper.GetAccount(ctx, acc).GetCoins().AmountOf(util.ConfigDelegatedDenom)
}

func (k Keeper) exists(ctx sdk.Context, acc sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(acc)
	return store.Has(keyBytes)
}

func (k Keeper) setStatus(ctx sdk.Context, target *types.R, value types.Status, acc sdk.AccAddress) {
	if target.Status == value {
		return
	}

	store := ctx.KVStore(k.indexStoreKey)
	key := make([]byte, len([]byte(acc))+1)
	copy(key[1:], acc)

	if target.Status >= minIndexedStatus {
		key[0] = uint8(target.Status)
		store.Delete(key)
	}

	target.Status = value
	if value >= minIndexedStatus {
		key[0] = uint8(value)
		store.Set(key, []byte{0x01})
	}
}

func setOrUpdate(m map[string]bank.Output, key sdk.AccAddress, amt int64) {
	keyStr := key.String()
	if item, ok := m[keyStr]; ok {
		amt += item.Coins.AmountOf(util.ConfigMainDenom).Int64()
	}
	m[keyStr] = bank.NewOutput(key, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(amt))))
}

// ScheduleCompression adds a record to scheduler, but does *NOT* affect referral's own KVStore.
func (k Keeper) ScheduleCompression(ctx sdk.Context, acc sdk.AccAddress, compressionAt int64) error {
	data := acc.Bytes()

	return sdkerrors.Wrap(
		k.scheduleKeeper.ScheduleTask(ctx, uint64(compressionAt), CompressionHookName, &data),
		"cannot schedule compression",
	)
}

// ValidateTransition checks if an account transition valid. This methods fails if subject's R.Transition is not nil.
func (k Keeper) ValidateTransition(ctx sdk.Context, subject, newParent sdk.AccAddress) error {
	return k.validateTransition(ctx, subject, newParent, true)
}

func (k Keeper) validateTransition(ctx sdk.Context, subject, newParent sdk.AccAddress, fresh bool) error {
	var (
		r, p types.R
		err  error
	)

	if subject.Empty() {
		return errors.New("missing subject address")
	}
	if newParent.Empty() {
		return errors.New("missing destination address")
	}
	if subject.Equals(newParent) {
		return errors.New("subject cannot be their own referral")
	}
	if r, err = k.get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}
	if fresh {
		if !r.Transition.Empty() {
			return errors.New("transition is already requested")
		}
	} else {
		if !r.Transition.Equals(newParent) {
			return errors.New("new parent address mismatch")
		}
	}
	if r.Referrer.Equals(newParent) {
		return errors.New("destination address is already subject's referrer")
	}
	if p, err = k.get(ctx, newParent); err != nil {
		return errors.Wrap(err, "destination account data missing")
	}
	for p.Referrer != nil {
		if p.Referrer.Equals(subject) {
			return errors.New("cycles are not allowed")
		}
		if p, err = k.get(ctx, p.Referrer); err != nil {
			panic(errors.Wrap(err, "referral structure is compromised"))
		}
	}
	return nil
}
