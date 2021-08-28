package keeper

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/earning/types"
)

// Keeper of the earning store
type Keeper struct {
	cdc            codec.BinaryMarshaler
	storeKey       sdk.StoreKey
	paramspace     types.ParamSubspace
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	scheduleKeeper types.ScheduleKeeper
}

// NewKeeper creates a earning keeper
func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	paramspace types.ParamSubspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	scheduleKeeper types.ScheduleKeeper,
) Keeper {
	keeper := Keeper{
		cdc:            cdc,
		storeKey:       key,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		scheduleKeeper: scheduleKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ListEarners(ctx sdk.Context, items []types.Earner) error {
	if k.GetState(ctx).Locked {
		return types.ErrLocked
	}
	for _, earner := range items {
		if k.has(ctx, earner.GetAccount()) {
			k.Logger(ctx).Error("account listed twice", "accAddress", earner.Account)
			return types.ErrAlreadyListed
		}
		k.set(ctx, earner.GetAccount(), earner.GetPoints())
	}
	return nil
}

func (k Keeper) Run(ctx sdk.Context, fundPart util.Fraction, perBlock uint32, total types.Points, time time.Time) error {
	if ctx.BlockTime().After(time) {
		return types.ErrTooLate
	}

	//TODO: Get VPN/Storage account states for a specific moment of the time
	vpnFund := fundPart.MulInt64(k.getModuleBalance(ctx, types.VpnCollectorName)).Int64()
	storageFund := fundPart.MulInt64(k.getModuleBalance(ctx, types.StorageCollectorName)).Int64()
	if vpnFund == 0 && storageFund == 0 {
		return types.ErrNoMoney
	}
	if vpnFund > 0 {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.VpnCollectorName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(vpnFund)))); err != nil {
			panic(errors.Wrap(err, "cannot send coins from VPN"))
		}
	}
	if storageFund > 0 {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.StorageCollectorName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(storageFund)))); err != nil {
			panic(errors.Wrap(err, "cannot send coins from Storage"))
		}
	}
	residualQuotient := util.NewFraction(k.getModuleBalance(ctx, types.ModuleName), vpnFund+storageFund).Reduce()
	vpnPointCost := util.NewFraction(vpnFund, total.Vpn).Mul(residualQuotient)
	storagePointCost := util.NewFraction(storageFund, total.Storage).Mul(residualQuotient)
	k.SetState(ctx, types.NewStateLocked(vpnPointCost, storagePointCost, perBlock))
	k.scheduleKeeper.ScheduleTask(ctx, time, types.StartHookName, nil)

	return nil
}

func (k Keeper) getModuleBalance(ctx sdk.Context, module string) int64 {
	return k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(module)).AmountOf(util.ConfigMainDenom).Int64()
}

func (k Keeper) Reset(ctx sdk.Context) {
	k.clear(ctx)
	k.SetState(ctx, types.NewStateUnlocked())
}

func (k Keeper) MustPerformStart(ctx sdk.Context, _ []byte, _ time.Time) {
	if err := k.PerformStart(ctx); err != nil {
		panic(sdkerrors.Wrap(err, fmt.Sprintf("cannot process %s hook", types.StartHookName)))
	}
}

func (k Keeper) MustPerformContinue(ctx sdk.Context, _ []byte, _ time.Time) {
	if err := k.PerformContinue(ctx); err != nil {
		panic(sdkerrors.Wrap(err, fmt.Sprintf("cannot process %s hook", types.ContinueHookName)))
	}
}

func (k Keeper) PerformStart(ctx sdk.Context) error {
	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventStartPaying{},
	); err != nil { panic(err) }
	return k.proceed(ctx)
}

func (k Keeper) PerformContinue(ctx sdk.Context) error { return k.proceed(ctx) }

//-----------------------------------------------------------------------------------------------------------

func (k Keeper) has(ctx sdk.Context, key sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	byteKey := []byte(key)
	return store.Has(byteKey)
}

// Get returns the pubkey from the adddress-pubkey relation
func (k Keeper) get(ctx sdk.Context, key sdk.AccAddress) (types.Points, error) {
	store := ctx.KVStore(k.storeKey)
	var item types.Points
	byteKey := []byte(key)
	err := k.cdc.UnmarshalBinaryBare(store.Get(byteKey), &item)
	if err != nil {
		return types.Points{}, err
	}
	return item, nil
}

func (k Keeper) set(ctx sdk.Context, key sdk.AccAddress, value types.Points) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryBare(&value)
	if err != nil {
		panic(err)
	}
	store.Set([]byte(key), bz)
}

func (k Keeper) delete(ctx sdk.Context, key sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(key))
}

func (k Keeper) clear(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	var keys [][]byte
	it := store.Iterator(nil, nil)
	for ; it.Valid(); it.Next() {
		keys = append(keys, it.Key())
	}
	it.Close()
	for _, key := range keys {
		store.Delete(key)
	}
}

func (k Keeper) proceed(ctx sdk.Context) error {
	p := k.GetState(ctx)
	if !p.Locked {
		return types.ErrNotLocked
	}
	page := make([]types.Earner, 0, p.ItemsPerBlock)
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	for i := uint32(0); i < p.ItemsPerBlock && it.Valid(); i++ {
		var points types.Points
		if err := k.cdc.UnmarshalBinaryBare(it.Value(), &points); err != nil {
			defer it.Close()
			return err
		}
		page = append(page, types.Earner{
			Account: sdk.AccAddress(it.Key()).String(),
			Vpn:     points.Vpn,
			Storage: points.Storage,
		})
		it.Next()
	}
	finished := !it.Valid()
	it.Close()

	for _, item := range page {
		vpnAmt := p.VpnPointCost.MulInt64(item.Vpn).Int64()
		storageAmt := p.StoragePointCost.MulInt64(item.Storage).Int64()
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, item.GetAccount(), sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(vpnAmt+storageAmt)))); err != nil {
			return err
		}
		if err := ctx.EventManager().EmitTypedEvent(
			&types.EventEarn{
				Address: item.Account,
				Vpn:     uint64(vpnAmt),
				Storage: uint64(storageAmt),
			},
		); err != nil { panic(err) }
		k.delete(ctx, item.GetAccount())
	}

	if finished {
		if err := ctx.EventManager().EmitTypedEvent(&types.EventFinishPaying{}); err != nil { panic(err) }
		k.SetState(ctx, types.NewStateUnlocked())
	} else {
		k.scheduleKeeper.ScheduleTask(ctx, ctx.BlockTime().Add(time.Nanosecond), types.ContinueHookName, nil)
	}
	return nil
}
