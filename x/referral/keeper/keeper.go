package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/referral/types"
)

const (
	minIndexedStatus = types.STATUS_BUSINESSMAN
)

// Keeper of the referral store
type Keeper struct {
	cdc            codec.BinaryMarshaler
	storeKey       sdk.StoreKey
	indexStoreKey  sdk.StoreKey
	paramspace     types.ParamSubspace
	accKeeper      types.AccountKeeper
	scheduleKeeper types.ScheduleKeeper
	bankKeeper     types.BankKeeper
	supplyKeeper   types.SupplyKeeper
	eventHooks     map[string][]func(ctx sdk.Context, acc string) error
}

// NewKeeper creates a referral keeper
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, idxKey sdk.StoreKey, paramspace types.ParamSubspace,
	accKeeper types.AccountKeeper, scheduleKeeper types.ScheduleKeeper, bankKeeper types.BankKeeper,
	supplyKeeper types.SupplyKeeper,
) Keeper {
	keeper := Keeper{
		cdc:            cdc,
		storeKey:       key,
		indexStoreKey:  idxKey,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		accKeeper:      accKeeper,
		scheduleKeeper: scheduleKeeper,
		bankKeeper:     bankKeeper,
		supplyKeeper:   supplyKeeper,
		eventHooks:     make(map[string][]func(ctx sdk.Context, acc string) error),
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetStatus returns a status for an account (i.e. lvl 1 "Lucky", lvl 2 "Leader", lvl 3 "Master" or so on)
func (k Keeper) GetStatus(ctx sdk.Context, acc string) (types.Status, error) {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return 0, err
	}
	return data.Status, nil
}

// GetParent returns a parent for an account
func (k Keeper) GetParent(ctx sdk.Context, acc string) (string, error) {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return "", errors.Wrap(err, "cannot obtain data")
	}
	return data.Referrer, nil
}

// GetChildren returns children (1st line only) for an account
func (k Keeper) GetChildren(ctx sdk.Context, acc string) ([]string, error) {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain data")
	}
	return data.Referrals, nil
}

// GetReferralFeesForSubscription returns a set of account-ratio pairs, describing what part of monthly subscription
// should go to what wallet. 0.85 total. The rest goes for validator and leader bonuses.
func (k Keeper) GetReferralFeesForSubscription(ctx sdk.Context, acc string) ([]types.ReferralFee, error) {
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
func (k Keeper) GetReferralFeesForDelegating(ctx sdk.Context, acc string) ([]types.ReferralFee, error) {
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
func (k Keeper) AreStatusRequirementsFulfilled(ctx sdk.Context, acc string, s types.Status) (types.StatusCheckResult, error) {
	if s < types.MinimumStatus || s > types.MaximumStatus {
		return types.StatusCheckResult{Overall: false}, fmt.Errorf("there is no such status: %d", s)
	}
	data, err := k.Get(ctx, acc)
	if err != nil {
		return types.StatusCheckResult{Overall: false}, err
	}
	return checkStatusRequirements(s, data, newBunchUpdater(k, ctx))
}

// AddTopLevelAccount adds accounts without parent and is supposed to be used during genesis
func (k Keeper) AddTopLevelAccount(ctx sdk.Context, acc string) (err error) {
	k.Logger(ctx).Debug("AddTopLevelAccount", "acc", acc)
	defer func() {
		if e := recover(); e != nil {
			k.Logger(ctx).Error("AddTopLevelAccount paniced", "err", e)
			if er, ok := e.(error); ok {
				err = errors.Wrap(er, "AddTopLevelAccount paniced")
			} else {
				err = errors.Errorf("AddTopLevelAccount paniced: %s", e)
			}
		}
	}()
	if k.exists(ctx, acc) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("account %s already exists", acc),
		)
	}
	var (
		bu        = newBunchUpdater(k, ctx)
		coins     = k.getBalance(ctx, acc)
		delegated = k.getDelegated(ctx, acc)
	)
	newItem := types.NewInfo("", coins, delegated)
	if err = bu.set(acc, newItem); err != nil {
		return err
	}
	if err = bu.commit(); err != nil {
		return err
	}
	return nil
}

// GetTopLevelAccounts returns all accounts without parents and is supposed to be used during genesis export
func (k Keeper) GetTopLevelAndBanishedAccounts(ctx sdk.Context) (topLevel []string, banished []types.Banished, err error) {
	store := ctx.KVStore(k.storeKey)
	itr := store.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		v := itr.Value()
		var record types.Info
		err = k.cdc.UnmarshalBinaryBare(v, &record)
		if err != nil {
			return nil,nil, err
		}
		addr := string(itr.Key())
		if record.Banished {
			banished = append(banished, types.Banished{
				Account:        addr,
				FormerReferrer: record.Referrer,
			})
		} else if record.Referrer == "" {
			topLevel = append(topLevel, addr)
		}
	}
	return topLevel, banished, nil
}

// AppendChild adds a new account to the referral structure. The parent account should already exist and the child one
// should not.
func (k Keeper) AppendChild(ctx sdk.Context, parentAcc string, childAcc string) error {
	return k.appendChild(ctx, parentAcc, childAcc, false, true)
}
func (k Keeper) appendChild(ctx sdk.Context, parentAcc string, childAcc string, skipActivityCheck, setCompressionTime bool) error {
	if parentAcc == "" {
		return types.ErrParentNil
	}
	if k.exists(ctx, childAcc) {
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("account %s already exists", childAcc),
		)
	}
	var (
		bu            = newBunchUpdater(k, ctx)
		anc           = parentAcc
		coins         = k.getBalance(ctx, childAcc)
		delegated     = k.getDelegated(ctx, childAcc)
	)
	newItem := types.NewInfo(parentAcc, coins, delegated)
	if setCompressionTime {
		compressionAt := ctx.BlockTime().Add(k.CompressionPeriod(ctx))
		newItem.CompressionAt = &compressionAt
	}
	err := bu.set(childAcc, newItem)
	if err != nil {
		return sdkerrors.Wrap(err, "cannot set "+childAcc)
	}

	var registrationClosed bool
	err = bu.update(parentAcc, true, func(value *types.Info) error {
		value.Coins[1] = value.Coins[1].Add(coins)
		value.Delegated[1] = value.Delegated[1].Add(delegated)
		bu.addCallback(StakeChangedCallback, anc)
		value.Referrals = append(value.Referrals, childAcc)
		anc = value.Referrer
		if !skipActivityCheck {
			registrationClosed = value.RegistrationClosed(ctx, k.scheduleKeeper)
		}
		return nil
	})
	if err != nil {
		return sdkerrors.Wrap(err, "cannot update "+anc)
	}
	if registrationClosed {
		return types.ErrRegistrationClosed
	}

	for i := 1; i < 10; i++ {
		if anc == "" {
			break
		}
		err = bu.update(anc, true, func(value *types.Info) error {
			value.Coins[i+1] = value.Coins[i+1].Add(coins)
			value.Delegated[i+1] = value.Delegated[i+1].Add(delegated)
			bu.addCallback(StakeChangedCallback, anc)
			if i == 0 {
				value.Referrals = append(value.Referrals, childAcc)
			}
			anc = value.Referrer
			return nil
		})
		if err != nil {
			return sdkerrors.Wrap(err, "cannot update "+anc)
		}
	}

	if err := bu.commit(); err != nil {
		return sdkerrors.Wrap(err, "cannot commit")
	}
	return nil
}

// Compress relocates all account's children under its parent, so the account looses its entire network.
func (k Keeper) Compress(ctx sdk.Context, acc string) error {
	var (
		bu         = newBunchUpdater(k, ctx)

		coins      []sdk.Int
		delegated  []sdk.Int
		children   []string
		activeRefs []string
		refsCount  []uint64
		parent     string
	)
	// Compressed account itself:
	//   * no referrals
	//   * no coins (neither delegated nor free)
	//   * status dump
	//   * shorten legs
	//   * no own children

	err := bu.update(acc, false, func(value *types.Info) error {
		children = value.Referrals
		activeRefs = value.ActiveReferrals
		coins = value.Coins
		delegated = value.Delegated
		parent = value.Referrer
		refsCount = value.ActiveRefCounts

		value.Referrals = nil
		value.ActiveReferrals = nil
		value.ActiveRefCounts = []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		value.Coins = []sdk.Int{
			coins[0],
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
		}
		value.Delegated = []sdk.Int{
			delegated[0],
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
			sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
		}
		value.CompressionAt = nil
		bu.addCallback(StakeChangedCallback, acc)
		k.setStatus(ctx, value, types.STATUS_LUCKY, acc)
		bu.addCallback(StatusUpdatedCallback, acc)

		if delegated[0].Int64() <= k.bankKeeper.GetParams(ctx).DustDelegation {
			k.scheduleBanishment(ctx, acc, value)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Children: just new referrer
	for _, child := range children {
		if err = bu.update(child, false, func(value *types.Info) error {
			value.Referrer = parent
			return nil
		}); err != nil {
			return err
		}
	}

	// Ancestors (level k, 1 <= k <= 10):
	//   * coins[i] pop from level k+i to level k+i-1 (, for 0 < i < 11-k)
	//   * coins[11-k] appears at level 10
	//   * extend leg (as a distance shrinks, new nodes might appear in 10-lvl-radius)
	// Parent (k = 1) only:
	//   * new referrals
	for k, anc := 1, parent; k <= 10 && anc != ""; k++ {
		err = bu.update(anc, true, func(value *types.Info) error {
			bu.addCallback(StakeChangedCallback, anc)
			anc = value.Referrer
			value.Coins[k] = value.Coins[k].Add(coins[1])
			value.Delegated[k] = value.Delegated[k].Add(delegated[1])
			value.ActiveRefCounts[k] += refsCount[1]
			for i := 1; i < 10-k; i++ {
				value.Coins[k+i] = value.Coins[k+i].Add(coins[i+1]).Sub(coins[i])
				value.Delegated[k+i] = value.Delegated[k+i].Add(delegated[i+1]).Sub(delegated[i])
				value.ActiveRefCounts[k+i] += refsCount[i+1] - refsCount[i]
			}
			if k == 1 {
				value.Referrals = append(value.Referrals, children...)
				value.ActiveReferrals = util.MergeStringsSorted(value.ActiveReferrals, activeRefs)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	util.EmitEvent(ctx,
		&types.EventCompression{
			Address: acc,
			Referrer: parent,
			Referrals: children,
		},
	)

	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}

// GetCoinsInNetwork returns total amount of coins (delegated and not) in a person's network
// (at levels that are open according the person's current status, but no deeper than `maxDepth` levels down).
// Own coins inclusive. maxDepth = 0 means no limits.
func (k Keeper) GetCoinsInNetwork(ctx sdk.Context, acc string, maxDepth int) (sdk.Int, error) {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	data, err := k.Get(ctx, acc)
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
func (k Keeper) GetDelegatedInNetwork(ctx sdk.Context, acc string, maxDepth int) (sdk.Int, error) {
	data, err := k.Get(ctx, acc)
	if err != nil {
		return sdk.Int{}, err
	}
	d := data.Status.LinesOpened()
	if d > maxDepth {
		d = maxDepth
	}
	return data.DelegatedAtLevelsUpTo(d), nil
}

func (k Keeper) OnBalanceChanged(ctx sdk.Context, acc string) error {
	k.Logger(ctx).Debug("OnBalanceChanged", "acc", acc)
	var (
		bu = newBunchUpdater(k, ctx)

		dc, dd   sdk.Int
		node     string
		banished bool
	)
	if err := bu.update(acc, true, func(value *types.Info) error {
		if value.IsEmpty() {
			return types.ErrNotFound
		}
		newBalance := k.getBalance(ctx, acc)
		newDelegated := k.getDelegated(ctx, acc)

		dc = newBalance.Sub(value.Coins[0])
		dd = newDelegated.Sub(value.Delegated[0])
		if !dd.IsZero() {
			bu.addCallback(StakeChangedCallback, acc)

			if !value.Active {
				if newDelegated.Int64() <= k.bankKeeper.GetParams(ctx).DustDelegation {
					if !value.Banished && value.BanishmentAt == nil && (value.CompressionAt == nil || ctx.BlockTime().After(*value.CompressionAt)) {
						k.scheduleBanishment(ctx, acc, value)
					}
				} else {
					if value.BanishmentAt != nil {
						k.scheduleKeeper.Delete(ctx, *value.BanishmentAt, BanishHookName, []byte(acc))
						value.BanishmentAt = nil
					}
					if value.Banished {
						// TODO: Refactor
						// We cannot use ComeBack method here because of bunch updater cache. It's probably the perfect
						// time to get rid of it.

						var parent string
						for parent = value.Referrer; parent != ""; {
							pi, err := bu.get(parent)
							if err != nil {
								panic(errors.Wrapf(err, "cannot obtain parent's (%s) data", parent))
							}
							if !pi.RegistrationClosed(ctx, k.scheduleKeeper) {
								break
							}
							parent = pi.Referrer
						}
						value.Referrer = parent
						c := value.Coins[0]
						d := value.Delegated[0]

						value.Banished = false
						value.Status = types.STATUS_LUCKY

						if parent != "" {
							var p2 string
							if err := bu.update(parent, true, func(value *types.Info) error {
								p2 = value.Referrer

								value.Referrals = append(value.Referrals, acc)
								value.Coins[1] = value.Coins[1].Add(c)
								value.Delegated[1] = value.Delegated[1].Add(d)

								return nil
							}); err != nil {
								return errors.Wrapf(err, "cannot update parent's (%s) data", parent)
							} else {
								parent = p2
							}
						}
						if !(c.IsZero() && d.IsZero() || parent == "") {
							for lvl := 2; lvl <= 10; lvl++ {
								if parent == "" {
									break
								}

								var p2 string
								if err := bu.update(parent, true, func(value *types.Info) error {
									p2 = value.Referrer

									value.Coins[lvl] = value.Coins[lvl].Add(c)
									value.Delegated[lvl] = value.Delegated[lvl].Add(d)

									return nil
								}); err != nil {
									return errors.Wrapf(err, "cannot update level %d ancestor's (%s) data", lvl, parent)
								} else {
									parent = p2
								}
							}
						}
					}
				}
			}
		}
		node = value.Referrer
		banished = value.Banished

		value.Coins[0] = newBalance
		value.Delegated[0] = newDelegated
		return nil
	}); err != nil {
		if errors.Is(err, types.ErrNotFound) {
			k.Logger(ctx).Debug("account is out of the referral", "acc", acc)
			return nil
		} else {
			k.Logger(ctx).Error("OnBalanceChanged hook failed", "acc", acc, "step", 0, "error", err)
			return err
		}
	}

	if !banished {
		for i := 1; i <= 10; i++ {
			if node == "" {
				break
			}

			if err := bu.update(node, true, func(value *types.Info) error {
				value.Coins[i] = value.Coins[i].Add(dc)
				value.Delegated[i] = value.Delegated[i].Add(dd)
				if !dd.IsZero() {
					bu.addCallback(StakeChangedCallback, node)
				}

				node = value.Referrer
				return nil
			}); err != nil {
				k.Logger(ctx).Error("OnBalanceChanged hook failed", "acc", acc, "step", i, "error", err)
				return err
			}
		}
	}

	if err := bu.commit(); err != nil {
		k.Logger(ctx).Error("OnBalanceChanged hook failed", "acc", acc, "step", "commit", "error", err)
		return err
	}
	return nil
}

func (k Keeper) SetActive(ctx sdk.Context, acc string, value, checkAncestorsForStatusUpdate bool) error {
	var (
		bu                = newBunchUpdater(k, ctx)
		valueIsAlreadySet = false

		parent        string
		delta         func(*uint64)
		refDelta      func(*[]string)
		compressionAt time.Time
	)
	if value {
		delta = func(x *uint64) { *x += 1 }
		refDelta = func(xs *[]string) { util.AddStringSorted(xs, acc) }
	} else {
		delta = func(x *uint64) { *x -= 1 }
		refDelta = func(xs *[]string) { util.RemoveStringPreserveOrder(xs, acc) }
		compressionAt = ctx.BlockTime().Add(k.CompressionPeriod(ctx))
	}

	err := bu.update(acc, false, func(x *types.Info) error {
		if x.Active == value {
			valueIsAlreadySet = true
		} else {
			x.Active = value
			delta(&x.ActiveRefCounts[0])
			if compressionAt.IsZero() {
				x.CompressionAt = nil
			} else {
				x.CompressionAt = &compressionAt
			}
			parent = x.Referrer
			if value && x.BanishmentAt != nil {
				k.scheduleKeeper.Delete(ctx, *x.BanishmentAt, BanishHookName, []byte(acc))
				x.BanishmentAt = nil
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "cannot update acc info")
	} else if valueIsAlreadySet {
		return nil
	}

	for i := 0; i < 10; i++ {
		if parent == "" {
			break
		}

		err = bu.update(parent, checkAncestorsForStatusUpdate, func(x *types.Info) error {
			if i == 0 {
				refDelta(&x.ActiveReferrals)
			}
			delta(&x.ActiveRefCounts[i+1])
			parent = x.Referrer
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "cannot update ancestor's referral count (#%d)", i)
		}
	}

	if !value && !valueIsAlreadySet {
		k.ScheduleCompression(ctx, acc, ctx.BlockTime().Add(k.CompressionPeriod(ctx)))
	}

	if err := bu.commit(); err != nil {
		return errors.Wrap(err, "cannot persist data")
	}
	return nil
}

func (k Keeper) MustSetActive(ctx sdk.Context, acc string, value bool) {
	if err := k.SetActive(ctx, acc, value, true); err != nil {
		panic(err)
	}
}

// MustSetActiveWithoutStatusUpdate updates active referrals but skips status update check after it. So this check MUST
// be performed from the outer code later. This is useful for massive updates like genesis init, because it allows to
// avoid excessive checks repeating again and again for the same account (every time any of referrals up to 10 lines
// down changes its activity).
func (k Keeper) MustSetActiveWithoutStatusUpdate(ctx sdk.Context, acc string, value bool) {
	if err := k.SetActive(ctx, acc, value, false); err != nil {
		panic(err)
	}
}

func (k Keeper) PayStatusBonus(ctx sdk.Context) error {
	var (
		ca     = k.GetParams(ctx).CompanyAccounts
		sender = ca.GetStatusBonuses()
		topRef = ca.GetTopReferrer()
		amt    = k.bankKeeper.GetBalance(ctx, sender).AmountOf(util.ConfigMainDenom).Int64() / 5
	)
	if amt == 0 {
		k.Logger(ctx).Debug("Nothing to pay")
		return nil
	}
	var (
		store           = ctx.KVStore(k.indexStoreKey)
		receivers       = make([]sdk.AccAddress, 0)
		outMap          = make(map[string]bank.Output)
		total     int64 = 0
	)

	for status := types.STATUS_ABSOLUTE_CHAMPION; status >= types.STATUS_BUSINESSMAN; status-- {
		it := sdk.KVStorePrefixIterator(store, []byte{uint8(status)})
		for ; it.Valid(); it.Next() {
			acc, err := sdk.AccAddressFromBech32(string(it.Key()[1:]))
			if err != nil {
				panic(err)
			}
			receivers = append(receivers, acc)
		}
		it.Close()
		if len(receivers) == 0 {
			setOrUpdate(outMap, topRef, amt)
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
		util.EmitEvent(ctx,
			&types.EventStatusBonus{
				Address: out.Address.String(),
				Amount:  out.Coins.AmountOf(util.ConfigMainDenom).Uint64(),
			},
		)
	}

	inputs := []bank.Input{bank.NewInput(sender, util.Uartrs(total))}
	k.Logger(ctx).Debug("PayStatusBonus", "in", inputs, "out", outputs)
	return k.bankKeeper.InputOutputCoins(ctx, inputs, outputs)
}

func (k Keeper) Iterate(ctx sdk.Context, callback func(acc string, r *types.Info) (changed, checkForStatusUpdate bool)) {
	bu := newBunchUpdater(k, ctx)
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	defer func() {
		if it != nil {
			it.Close()
		}
	}()
	for ; it.Valid(); it.Next() {
		var acc = string(it.Key())
		var item types.Info
		if err := k.cdc.UnmarshalBinaryBare(it.Value(), &item); err != nil {
			panic(errors.Wrapf(err, `cannot unmarshal info for "%s"`, acc))
		}
		if changed, checkForStatusUpdate := callback(acc, &item); changed || checkForStatusUpdate {
			var f func(r *types.Info) error
			if changed {
				f = func(r *types.Info) error {
					*r = item
					return nil
				}
			} else {
				f = func(_ *types.Info) error {
					return nil
				}
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

// RequestTransaction is supposed to be called when a user wants to be moved under another referrer. If the current
// referrer do not approve this operation in a day, it will be cancelled.
func (k Keeper) RequestTransition(ctx sdk.Context, subject, newParent string) error {
	var (
		r   types.Info
		err error
	)

	if r, err = k.Get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}
	if err = k.validateTransition(ctx, subject, newParent, true); err != nil {
		return errors.Wrap(err, "transition is invalid")
	}

	params := k.GetParams(ctx)
	if params.TransitionPrice > 0 {
		if subject, err := sdk.AccAddressFromBech32(subject); err != nil {
			return errors.Wrap(err, "invalid subject address")
		} else {
			err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, subject, auth.FeeCollectorName, util.UartrsUint64(params.TransitionPrice))
			if err != nil {
				return errors.Wrap(err, "cannot pay commission")
			}
		}

		if r, err = k.Get(ctx, subject); err != nil {
			// This cannot be, because the same data was read just fine a moment ago.
			panic(err)
		}
	}

	r.Transition = newParent
	if err = k.set(ctx, subject, r); err != nil {
		panic(errors.Wrap(err, "cannot write to KVStore"))
	}

	k.scheduleKeeper.ScheduleTask(ctx, ctx.BlockTime().Add(k.scheduleKeeper.OneDay(ctx)), TransitionTimeoutHookName, []byte(subject))

	util.EmitEvent(ctx,
		&types.EventTransitionRequested{
			Address: subject,
			Before:  r.Referrer,
			After:   newParent,
		},
	)
	return nil
}

// CancelTransition is supposed to be called when either a current referrer declines a referral transition or this
// transition timeout occurs. See also RequestTransition method.
func (k Keeper) CancelTransition(ctx sdk.Context, subject string, timeout bool) error {
	var (
		r   types.Info
		err error
	)
	if r, err = k.Get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}
	value := r.Transition
	r.Transition = ""
	if err = k.set(ctx, subject, r); err != nil {
		panic(errors.Wrap(err, "cannot write to KVStore"))
	}

	var reason types.EventTransitionDeclined_Reason
	if timeout {
		reason = types.REASON_TIMEOUT
	} else {
		reason = types.REASON_DECLINED
	}
	util.EmitEvent(ctx,
		&types.EventTransitionDeclined{
			Address: subject,
			Before:  r.Referrer,
			After:   value,
			Reason:  reason,
		},
	)
	return nil
}

// AffirmTransition is supposed to be called when a current referrer approves a referral transition. Actual subtree
// relocation and all the according recalculations and updates are done here.
func (k Keeper) AffirmTransition(ctx sdk.Context, subject string) error {
	var (
		r   types.Info
		err error
	)

	if r, err = k.Get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}

	// we should double-check, just in case something has changed
	if err = k.validateTransition(ctx, subject, r.Transition, false); err != nil {
		return errors.Wrap(err, "transition is invalid")
	}

	oldParent, newParent := r.Referrer, r.Transition
	r.Referrer, r.Transition = newParent, ""
	if err = k.set(ctx, subject, r); err != nil {
		panic(errors.Wrap(err, "cannot write to KVStore"))
	}

	var (
		bu                       = newBunchUpdater(k, ctx)
		oldAncestor, newAncestor string
	)

	if err = bu.update(oldParent, true, func(value *types.Info) error {
		idx := 0
		for ; idx < len(value.Referrals) && value.Referrals[idx] != subject; idx++ {
		}
		value.Referrals[idx] = value.Referrals[len(value.Referrals)-1]
		value.Referrals = value.Referrals[:len(value.Referrals)-1]
		util.RemoveStringFast(&value.Referrals, subject)
		if r.Active {
			util.RemoveStringFast(&value.ActiveReferrals, subject)
		}

		for i := 1; i <= 10; i++ {
			value.Coins[i] = value.Coins[i].Sub(r.Coins[i-1])
			value.Delegated[i] = value.Delegated[i].Sub(r.Delegated[i-1])
			value.ActiveRefCounts[i] -= r.ActiveRefCounts[i-1]
		}

		oldAncestor = value.Referrer
		return nil
	}); err != nil {
		panic(errors.Wrap(err, "cannot update old referrer data"))
	}

	if err = bu.update(newParent, true, func(value *types.Info) error {
		value.Referrals = append(value.Referrals, subject)
		if r.Active {
			value.ActiveReferrals = append(value.ActiveReferrals, subject)
		}

		for i := 1; i <= 10; i++ {
			value.Coins[i] = value.Coins[i].Add(r.Coins[i-1])
			value.Delegated[i] = value.Delegated[i].Add(r.Delegated[i-1])
			value.ActiveRefCounts[i] += r.ActiveRefCounts[i-1]
		}

		newAncestor = value.Referrer
		return nil
	}); err != nil {
		panic(errors.Wrap(err, "cannot update new referrer data"))
	}

	for level := 2; level <= 10; level++ {
		if oldAncestor == newAncestor {
			break
		}
		if oldAncestor != "" {
			if err = bu.update(oldAncestor, true, func(value *types.Info) error {
				for i := level; i <= 10; i++ {
					value.Coins[i] = value.Coins[i].Sub(r.Coins[i-level])
					value.Delegated[i] = value.Delegated[i].Sub(r.Delegated[i-level])
					value.ActiveRefCounts[i] -= r.ActiveRefCounts[i-level]
				}
				oldAncestor = value.Referrer
				return nil
			}); err != nil {
				panic(errors.Wrapf(err, "cannot update old level-%d ancestor data", level))
			}
		}
		if newAncestor != "" {
			if err = bu.update(newAncestor, true, func(value *types.Info) error {
				for i := level; i <= 10; i++ {
					value.Coins[i] = value.Coins[i].Add(r.Coins[i-level])
					value.Delegated[i] = value.Delegated[i].Add(r.Delegated[i-level])
					value.ActiveRefCounts[i] += r.ActiveRefCounts[i-level]
				}
				newAncestor = value.Referrer
				return nil
			}); err != nil {
				panic(errors.Wrapf(err, "cannot update new level-%d ancestor data", level))
			}
		}
	}

	if err = bu.commit(); err != nil {
		panic(errors.Wrap(err, "cannot commit changes"))
	}

	util.EmitEvent(ctx,
		&types.EventTransitionPerformed{
			Address: subject,
			Before:  oldParent,
			After:   newParent,
		},
	)
	return nil
}

// GetPendingTransition returns a new referral that the specified account is requested to be moved under. It returns
// (nil, nil) if the account is OK, but a transition is not requested.
func (k Keeper) GetPendingTransition(ctx sdk.Context, acc string) (string, error) {
	if acc == "" {
		return "", sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "account address is missing")
	}
	r, err := k.Get(ctx, acc)
	if err != nil {
		return "", err
	}
	return r.Transition, nil
}

// Banish excludes an account from the referral due to a long inactivity
func (k Keeper) Banish(ctx sdk.Context, acc string) error {
	bu := newBunchUpdater(k, ctx)
	var (
		parent   string
		c, d     sdk.Int
		banished bool
	)
	if err := bu.update(acc, false, func(value *types.Info) error {
		parent = value.Referrer
		c = value.Coins[0]
		d = value.Delegated[0]
		banished = value.Banished

		// Double-check, just in case
		if value.Active {
			return errors.New("must not banish: account is active")
		}
		if d.Int64() > k.bankKeeper.GetParams(ctx).DustDelegation {
			return errors.New("must not banish: delegation")
		}

		value.Banished = true
		// Purge account data
		value.Status = types.STATUS_UNSPECIFIED
		value.CompressionAt = nil
		value.StatusDowngradeAt = nil
		value.BanishmentAt = nil

		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot update account info")
	}

	if banished {
		return errors.New("already banished")
	}

	if parent != "" {
		var p2 string
		if err := bu.update(parent, true, func(value *types.Info) error {
			p2 = value.Referrer

			util.RemoveStringFast(&value.Referrals, acc)
			value.Coins[1] = value.Coins[1].Sub(c)
			value.Delegated[1] = value.Delegated[1].Sub(d)

			return nil
		}); err != nil {
			return errors.Wrapf(err, "cannot update parent's (%s) info", parent)
		} else {
			parent = p2
		}
	}

	if !(parent == "" || c.IsZero() && d.IsZero()) {
		for lvl := 2; lvl <= 10; lvl++ {
			if parent == "" {
				break
			}
			var p2 string
			if err := bu.update(parent, true, func(value *types.Info) error {
				p2 = value.Referrer

				value.Coins[lvl] = value.Coins[lvl].Sub(c)
				value.Delegated[lvl] = value.Delegated[lvl].Sub(d)

				return nil
			}); err != nil {
				return errors.Wrapf(err, "cannot update level %d parent's (%s) info", lvl, parent)
			} else {
				parent = p2
			}
		}
	}

	if err := bu.commit(); err != nil {
		return errors.Wrap(err, "cannot commit changes")
	}

	if err := k.callback(BanishedCallback, ctx, acc); err != nil {
		return errors.Wrap(err, "callback failed")
	}

	util.EmitEvent(ctx,
		&types.EventAccBanished{
			Address: acc,
		},
	)
	return nil
}

// ComeBack returns a banished account back to the referral
func (k Keeper) ComeBack(ctx sdk.Context, acc string) error {
	bu := newBunchUpdater(k, ctx)

	var parent string
	var c, d sdk.Int
	if err := bu.update(acc, false, func(value *types.Info) error {
		for parent = value.Referrer; parent != ""; {
			pi, err := bu.get(parent)
			if err != nil {
				return errors.Wrapf(err, "cannot obtain parent's (%s) data", parent)
			}
			if !pi.RegistrationClosed(ctx, k.scheduleKeeper) {
				break
			}
			parent = pi.Referrer
		}
		value.Referrer = parent
		c = value.Coins[0]
		d = value.Delegated[0]

		value.Banished = false
		value.BanishmentAt = nil
		value.CompressionAt = nil
		value.Status = types.STATUS_LUCKY

		return nil
	}); err != nil {
		return errors.Wrap(err, "cannot update account data")
	}
	if parent != "" {
		var p2 string
		if err := bu.update(parent, true, func(value *types.Info) error {
			p2 = value.Referrer

			value.Referrals = append(value.Referrals, acc)
			value.Coins[1] = value.Coins[1].Add(c)
			value.Delegated[1] = value.Delegated[1].Add(d)

			return nil
		}); err != nil {
			return errors.Wrapf(err, "cannot update parent's (%s) data", parent)
		} else {
			parent = p2
		}
	}
	if !(parent == "" || c.IsZero() && d.IsZero()) {
		for lvl := 2; lvl <= 10; lvl++ {
			if parent == "" {
				break
			}

			var p2 string
			if err := bu.update(parent, true, func(value *types.Info) error {
				p2 = value.Referrer

				value.Coins[lvl] = value.Coins[lvl].Add(c)
				value.Delegated[lvl] = value.Delegated[lvl].Add(d)

				return nil
			}); err != nil {
				return errors.Wrapf(err, "cannot update level %d ancestor's (%s) data", lvl, parent)
			} else {
				parent = p2
			}
		}
	}

	if err := bu.commit(); err != nil {
		return errors.Wrap(err, "cannot apply changes")
	}

	//TODO: Emit an event if needed
	return nil
}

// Get returns all the data for an account (status, parent, children)
func (k Keeper) Get(ctx sdk.Context, acc string) (types.Info, error) {
	store := ctx.KVStore(k.storeKey)
	var item types.Info
	err := errors.Wrapf(
		k.cdc.UnmarshalBinaryBare(store.Get([]byte(acc)), &item),
		"no data for %s", acc,
	)
	return item, err
}

func (k Keeper) getReferralFeesCore(ctx sdk.Context, acc string, companyAccount string, toCompany util.Fraction, toAncestors []util.Fraction, topReferrer string) ([]types.ReferralFee, error) {
	if len(toAncestors) != 10 {
		return nil, errors.Errorf("toAncestors param must have exactly 10 items (%d found)", len(toAncestors))
	}
	excess := util.Percent(0)
	result := append(make([]types.ReferralFee, 0, 12), types.ReferralFee{Beneficiary: companyAccount, Ratio: toCompany})

	ancestor, err := k.GetParent(ctx, acc)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 10; i++ {
		var (
			data types.Info
			err  error
		)
		for {
			if ancestor == "" {
				break
			}
			data, err = k.Get(ctx, ancestor)
			if err != nil {
				return nil, err
			}
			if data.Active {
				break
			} else {
				ancestor = data.Referrer
			}
		}
		if ancestor == "" {
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

func (k Keeper) set(ctx sdk.Context, acc string, value types.Info) error {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := k.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)

	return nil
}

func (k Keeper) update(ctx sdk.Context, acc string, callback func(value types.Info) types.Info) error {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(acc)
	var value types.Info
	err := k.cdc.UnmarshalBinaryBare(store.Get(keyBytes), &value)
	if err != nil {
		return err
	}
	value = callback(value)
	valueBytes, err := k.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (k Keeper) getBalance(ctx sdk.Context, acc string) sdk.Int {
	if acc, err := sdk.AccAddressFromBech32(acc); err != nil {
		panic(err)
	} else {
		coins := k.bankKeeper.GetBalance(ctx, acc)
		return coins.AmountOf(util.ConfigMainDenom).
			Add(coins.AmountOf(util.ConfigDelegatedDenom)).
			Add(coins.AmountOf(util.ConfigRevokingDenom))
	}
}

func (k Keeper) getDelegated(ctx sdk.Context, acc string) sdk.Int {
	if acc, err := sdk.AccAddressFromBech32(acc); err != nil {
		panic(err)
	} else {
		return k.bankKeeper.GetBalance(ctx, acc).AmountOf(util.ConfigDelegatedDenom)
	}
}

func (k Keeper) exists(ctx sdk.Context, acc string) bool {
	store := ctx.KVStore(k.storeKey)
	keyBytes := []byte(acc)
	return store.Has(keyBytes)
}

func (k Keeper) setStatus(ctx sdk.Context, target *types.Info, value types.Status, acc string) {
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
func (k Keeper) ScheduleCompression(ctx sdk.Context, acc string, compressionAt time.Time) {
	k.scheduleKeeper.ScheduleTask(ctx, compressionAt, CompressionHookName, []byte(acc))
}

func (k Keeper) validateTransition(ctx sdk.Context, subject, newParent string, fresh bool) error {
	var (
		r, p types.Info
		err  error
	)

	if _, err = sdk.AccAddressFromBech32(subject); err != nil {
		return errors.Wrap(err, "invalid subject address")
	}
	if _, err = sdk.AccAddressFromBech32(newParent); err != nil {
		return errors.New("invalid destination address")
	}
	if subject == newParent {
		return errors.New("subject cannot be their own referral")
	}
	if r, err = k.Get(ctx, subject); err != nil {
		return errors.Wrap(err, "subject account data missing")
	}
	if fresh {
		if r.Transition != "" {
			return errors.New("transition is already requested")
		}
	} else {
		if r.Transition != newParent {
			return errors.New("new parent address mismatch")
		}
	}
	if r.Referrer == newParent {
		return errors.New("destination address is already subject's referrer")
	}
	if p, err = k.Get(ctx, newParent); err != nil {
		return errors.Wrap(err, "destination account data missing")
	}
	if p.RegistrationClosed(ctx, k.scheduleKeeper) {
		return types.ErrRegistrationClosed
	}
	for p.Referrer != "" {
		ref := p.Referrer
		if ref == subject {
			return errors.New("cycles are not allowed")
		}
		if p, err = k.Get(ctx, ref); err != nil {
			panic(errors.Wrap(err, "referral structure is compromised"))
		}
	}
	return nil
}

func (k Keeper) CompressionPeriod(ctx sdk.Context) time.Duration {
	return 2 * k.scheduleKeeper.OneMonth(ctx)
}

func (k Keeper) scheduleBanishment(ctx sdk.Context, acc string, value *types.Info) {
	t := ctx.BlockTime().Add(k.scheduleKeeper.OneMonth(ctx))
	k.scheduleKeeper.ScheduleTask(ctx, t, BanishHookName, []byte(acc))
	value.BanishmentAt = &t
}
