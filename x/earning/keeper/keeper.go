package keeper

import (
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/vpn"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/x/earning/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the earning store
type Keeper struct {
	storeKey       sdk.StoreKey
	cdc            *codec.Codec
	paramspace     types.ParamSubspace
	supplyKeeper   types.SupplyKeeper
	scheduleKeeper types.ScheduleKeeper
}

// NewKeeper creates a earning keeper
func NewKeeper(
	cdc            *codec.Codec,
	key            sdk.StoreKey,
	paramspace     types.ParamSubspace,
	supplyKeeper   types.SupplyKeeper,
	scheduleKeeper types.ScheduleKeeper,
) Keeper {
	keeper := Keeper{
		storeKey:       key,
		cdc:            cdc,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		supplyKeeper:   supplyKeeper,
		scheduleKeeper: scheduleKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ListEarners(ctx sdk.Context, items []types.Earner) error {
	if k.GetState(ctx).Locked { return types.ErrLocked }
	for _, earner := range items {
		if k.has(ctx, earner.Account) {
			k.Logger(ctx).Error("account listed twice", "accAddress", earner.Account.String())
			return types.ErrAlreadyListed
		}
		k.set(ctx, earner.Account, earner.Points)
	}
	return nil
}

func (k Keeper) Run(ctx sdk.Context, fundPart util.Fraction, perBlock uint16, total types.Points, height int64) error {
	if ctx.BlockHeight() >= height { return types.ErrTooLate }

	//TODO: Get VPN/Storage account states for a specific moment of the time
	vpnFund := fundPart.MulInt64(k.supplyKeeper.GetModuleAccount(ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()).Int64()
	storageFund := fundPart.MulInt64(k.supplyKeeper.GetModuleAccount(ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()).Int64()
	if vpnFund == 0 && storageFund == 0 { return types.ErrNoMoney }
	if err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, vpn.ModuleName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(vpnFund)))); err != nil { return err }
	if err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, storage.ModuleName, types.ModuleName, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(storageFund)))); err != nil { return err }
	residualQuotient := util.NewFraction(k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(), vpnFund + storageFund).Reduce()
	vpnPointCost := util.NewFraction(vpnFund, total.Vpn).Mul(residualQuotient)
	storagePointCost := util.NewFraction(storageFund, total.Storage).Mul(residualQuotient)
	k.SetState(ctx, types.NewStateLocked(vpnPointCost, storagePointCost, perBlock))
	if err := k.scheduleKeeper.ScheduleTask(ctx, uint64(height), types.StartHookName, &noPayload); err != nil { return err }

	return nil
}

func (k Keeper) Reset(ctx sdk.Context) {
	k.clear(ctx)
	k.SetState(ctx, types.NewStateUnlocked())
}

func (k Keeper) MustPerformStart(ctx sdk.Context, payload []byte) {
	if err := k.PerformStart(ctx, payload); err != nil {
		panic(sdkerrors.Wrap(err, fmt.Sprintf("cannot process %s hook", types.StartHookName)))
	}
}

func (k Keeper) MustPerformContinue(ctx sdk.Context, payload []byte) {
	if err := k.PerformContinue(ctx, payload); err != nil {
		panic(sdkerrors.Wrap(err, fmt.Sprintf("cannot process %s hook", types.ContinueHookName)))
	}
}

func (k Keeper) PerformStart(ctx sdk.Context, _ []byte) error {
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeStart))
	return k.proceed(ctx)
}

func (k Keeper) PerformContinue(ctx sdk.Context, _ []byte) error { return k.proceed(ctx) }

//-----------------------------------------------------------------------------------------------------------

var noPayload []byte = nil

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
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(byteKey), &item)
	if err != nil {
		return types.Points{}, err
	}
	return item, nil
}

func (k Keeper) set(ctx sdk.Context, key sdk.AccAddress, value types.Points) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(value)
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
	if !p.Locked { return types.ErrNotLocked }
	page := make([]types.Earner, 0, p.ItemsPerBlock)
	store := ctx.KVStore(k.storeKey)
	it := store.Iterator(nil, nil)
	for i := uint16(0); i < p.ItemsPerBlock && it.Valid(); i++ {
		var points types.Points
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(it.Value(), &points); err != nil {
			defer it.Close()
			return err
		}
		page = append(page, types.Earner{
			Points:  points,
			Account: it.Key(),
		})
		it.Next()
	}
	finished := !it.Valid()
	it.Close()

	for _, item := range page {
		vpnAmt := p.VpnPointCost.MulInt64(item.Vpn).Int64()
		storageAmt := p.StoragePointCost.MulInt64(item.Storage).Int64()
		if err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, item.Account, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(vpnAmt + storageAmt)))); err != nil {
			return err
		}
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeEarn,
			sdk.NewAttribute(types.AttributeKeyAddress, item.Account.String()),
			sdk.NewAttribute(types.AttributeKeyVpn, fmt.Sprintf("%d", vpnAmt)),
			sdk.NewAttribute(types.AttributeKeyStorage, fmt.Sprintf("%d", storageAmt)),
		))
		k.delete(ctx, item.Account)
	}

	if finished {
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeFinish))
		k.SetState(ctx, types.NewStateUnlocked())
	} else {
		if err := k.scheduleKeeper.ScheduleTask(ctx, uint64(ctx.BlockHeight()+1), types.ContinueHookName, &noPayload); err != nil { return err }
	}
	return nil
}
