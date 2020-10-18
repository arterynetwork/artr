package keeper

import (
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/vpn"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/arterynetwork/artr/x/subscription/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	oneDay   = util.BlocksOneDay
	oneMonth = util.BlocksOneMonth
)

// Keeper of the subscription store
type Keeper struct {
	storeKey       sdk.StoreKey
	cdc            *codec.Codec
	paramspace     types.ParamSubspace
	bankKeeper     types.BankKeeper
	ReferralKeeper types.ReferralKeeper
	scheduleKeeper types.ScheduleKeeper
	vpnKeeper      types.VPNKeeper
	storageKeeper  types.StorageKeeper
	supplyKeeper   types.SupplyKeeper
	profileKeeper  types.ProfileKeeper
}

// NewKeeper creates a subscription keeper
func NewKeeper(cdc *codec.Codec,
	key sdk.StoreKey,
	paramspace types.ParamSubspace,
	bankKeeper types.BankKeeper,
	referralKeeper types.ReferralKeeper,
	scheduleKeeper types.ScheduleKeeper,
	vpnKeeper types.VPNKeeper,
	storageKeeper types.StorageKeeper,
	supplyKeeper types.SupplyKeeper,
	profileKeeper types.ProfileKeeper,
) Keeper {
	keeper := Keeper{
		storeKey:       key,
		cdc:            cdc,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		bankKeeper:     bankKeeper,
		ReferralKeeper: referralKeeper,
		scheduleKeeper: scheduleKeeper,
		vpnKeeper:      vpnKeeper,
		storageKeeper:  storageKeeper,
		supplyKeeper:   supplyKeeper,
		profileKeeper:  profileKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) payUpFees(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Int, event string) (int64, error) {
	fees, err := k.ReferralKeeper.GetReferralFeesForSubscription(ctx, addr)

	if err != nil {
		return 0, err
	}

	totalFee := int64(0)
	outputs := make([]bank.Output, len(fees))
	for i, fee := range fees {
		x := fee.Ratio.MulInt64(amount.Int64()).Int64()
		totalFee += x
		outputs[i] = bank.NewOutput(fee.Beneficiary, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(x))))
		//ctx.EventManager().EmitEvent(sdk.NewEvent(
		//	event,
		//	sdk.NewAttribute(types.AttributeKeyAddress, fee.Beneficiary.String()),
		//	sdk.NewAttribute(types.AttributeKeyAmount, fmt.Sprint(x)),
		//))
	}

	inputs := []bank.Input{bank.NewInput(addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(totalFee))))}

	err = k.bankKeeper.InputOutputCoins(ctx, inputs, outputs)
	if err != nil {
		return totalFee, err
	}

	return totalFee, nil
}

// Monthly payment
func (k Keeper) PayForSubscription(ctx sdk.Context, addr sdk.AccAddress, storageAmount int64) error {
	var (
		price  uint32
		course uint32
	)

	k.paramspace.Get(ctx, types.KeySubscriptionPrice, &price)
	k.paramspace.Get(ctx, types.KeyTokenCourse, &course)

	// Total price
	amount := sdk.NewInt(int64(price) * int64(course))
	txFee := util.CalculateFee(amount)

	// Total price without fee - we calc MLM reward based on this price
	amount = amount.Sub(txFee)

	totalFee, err := k.payUpFees(ctx, addr, amount, types.EventTypeFee)

	if err != nil {
		return err
	}

	// it's remain after all other fees
	moduleFee := amount.SubRaw(totalFee)
	vpnFee := moduleFee.QuoRaw(3)

	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, addr, vpn.ModuleName,
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, vpnFee)))

	if err != nil {
		return err
	}

	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, addr, storage.ModuleName,
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, moduleFee.Sub(vpnFee))))

	if err != nil {
		return err
	}

	// Update activity
	info := k.GetActivityInfo(ctx, addr)
	payInitialStorage := false

	// If not active
	if !info.Active {
		info.ExpireAt = ctx.BlockHeight() + oneMonth
		info.Active = true
		defer k.ReferralKeeper.SetActive(ctx, addr, true)
		payInitialStorage = true
		k.ScheduleRenew(ctx, addr, ctx.BlockHeight()+oneMonth)
		k.resetLimits(ctx, addr)
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeActivityChange,
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
			sdk.NewAttribute(types.AttributeKeyActive, types.AttributeValueKeyActiveActive),
		))
	} else {
		info.ExpireAt += oneMonth

		// Pay for 1 month of storage
	}

	k.SetActivityInfo(ctx, addr, info)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePaySubscription,
		sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		sdk.NewAttribute(types.AttributeKeyNodeFee, txFee.String()),
		sdk.NewAttribute(types.AttributeKeyExpireAt, fmt.Sprintf("%d", info.ExpireAt)),
	))

	if payInitialStorage {
		return k.PayForStorage(ctx, addr, storageAmount)
	} else {
		var storageGb uint32
		k.paramspace.Get(ctx, types.KeyBaseStorageGb, &storageGb)

		baseStorageLimit := int64(storageGb) * util.GBSize
		payAmount := storageAmount - baseStorageLimit
		k.storageKeeper.SetLimit(ctx, addr, storageAmount)

		if payAmount > 0 {
			return k.payForService(ctx, addr, payAmount, storage.ModuleName,
				types.KeyStorageGbPrice, storageAmount, types.EventTypePayStorage)
		}
	}

	return nil
}

func (k Keeper) payForService(ctx sdk.Context, addr sdk.AccAddress, amount int64,
	moduleName string, priceAttr []byte, limitForEvent int64, eventName string) error {
	var (
		price  uint32
		course uint32
	)

	k.paramspace.Get(ctx, priceAttr, &price)
	k.paramspace.Get(ctx, types.KeyTokenCourse, &course)

	amountPrice := sdk.NewInt(amount *
		int64(price) *
		int64(course) / util.GBSize)

	txFee := util.CalculateFee(amountPrice)
	amountPriceWithFee := amountPrice.Sub(txFee)

	//totalFee, err := k.payUpFees(ctx, addr, amountPriceWithFee, types.EventTypeFee)

	//if err != nil {
	//	return err
	//}

	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, addr, moduleName,
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, amountPriceWithFee)))

	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		eventName,
		sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amountPrice.String()),
		sdk.NewAttribute(types.AttributeKeyLimit, fmt.Sprintf("%d", limitForEvent)),
	))

	return nil
}

// Payment for VPN
func (k Keeper) PayForVPN(ctx sdk.Context, addr sdk.AccAddress, amount int64) error {

	newLimit, err := k.vpnKeeper.AddLimit(ctx, addr, amount)

	if err != nil {
		return err
	}

	return k.payForService(ctx, addr, amount, vpn.ModuleName,
		types.KeyVPNGbPrice, newLimit, types.EventTypePayVPN)
}

// Payment for Storage place
func (k Keeper) PayForStorage(ctx sdk.Context, addr sdk.AccAddress, amount int64) error {

	current := k.storageKeeper.GetCurrent(ctx, addr)

	if current > amount {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "limit is smaller then current size")
	}

	var storageGb uint32
	k.paramspace.Get(ctx, types.KeyBaseStorageGb, &storageGb)

	baseStorageLimit := int64(storageGb) * util.GBSize

	if amount < baseStorageLimit {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "limit is smaller then minimum")
	}

	currentLimit := k.storageKeeper.GetLimit(ctx, addr)

	if currentLimit < baseStorageLimit {
		currentLimit = baseStorageLimit
	}

	info := k.GetActivityInfo(ctx, addr)

	if !info.Active || (info.ExpireAt-ctx.BlockHeight() <= 0) {
		return types.ErrInactiveSubscription
	}

	newAmount := amount - currentLimit
	k.storageKeeper.SetLimit(ctx, addr, amount)

	if newAmount > 0 {
		remainAmount := sdk.NewInt(newAmount).
			MulRaw(info.ExpireAt - ctx.BlockHeight()).
			QuoRaw(util.BlocksOneMonth).
			Int64()

		return k.payForService(ctx, addr, remainAmount, storage.ModuleName,
			types.KeyStorageGbPrice, amount, types.EventTypePayStorage)
	}

	return nil
}

func (k Keeper) IsActive(ctx sdk.Context, addr sdk.AccAddress) bool {
	info := k.GetActivityInfo(ctx, addr)
	return info.Active
}

func (k Keeper) GetActivityInfo(ctx sdk.Context, addr sdk.AccAddress) types.ActivityInfo {
	store := ctx.KVStore(k.storeKey)
	var info types.ActivityInfo
	bz := store.Get(auth.AddressStoreKey(addr))

	if bz == nil {
		return types.ActivityInfo{}
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)

	return info
}

func (k Keeper) SetActivityInfo(ctx sdk.Context, addr sdk.AccAddress, info types.ActivityInfo) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryLengthPrefixed(info)

	if err != nil {
		panic(err)
	}

	store.Set(auth.AddressStoreKey(addr), bz)
}

func (k Keeper) ScheduleRenew(ctx sdk.Context, addr sdk.AccAddress, height int64) {
	bytes := addr.Bytes()
	k.scheduleKeeper.ScheduleTask(ctx, uint64(height), types.HookName, &bytes)
	//fmt.Println(height, bytes)
}

func (k Keeper) resetLimits(ctx sdk.Context, addr sdk.AccAddress) {
	var VPNGb uint32
	var storageGb uint32
	k.paramspace.Get(ctx, types.KeyBaseVPNGb, &VPNGb)
	k.paramspace.Get(ctx, types.KeyBaseStorageGb, &storageGb)

	k.vpnKeeper.SetLimit(ctx, addr, int64(VPNGb)*util.GBSize)
	k.vpnKeeper.SetCurrent(ctx, addr, 0)
	storageLimit := k.storageKeeper.GetLimit(ctx, addr)
	if storageLimit == 0 {
		k.storageKeeper.SetLimit(ctx, addr, int64(storageGb)*util.GBSize)
	}
}

func (k Keeper) autoPay(ctx sdk.Context, addr sdk.AccAddress) error {
	current := k.storageKeeper.GetCurrent(ctx, addr)
	return k.PayForSubscription(ctx, addr, current)
}

func (k Keeper) deactivateAccount(ctx sdk.Context, addr sdk.AccAddress, info types.ActivityInfo) {
	info.Active = false
	k.SetActivityInfo(ctx, addr, info)
	err := k.ReferralKeeper.SetActive(ctx, addr, false)
	if err != nil {
		k.Logger(ctx).Error(err.Error(), addr)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeActivityChange,
		sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
		sdk.NewAttribute(types.AttributeKeyActive, types.AttributeVAlueKeyActiveInactive),
	))
}

func (k Keeper) ProcessSchedule(ctx sdk.Context, data []byte) {
	var addr sdk.AccAddress
	err := addr.Unmarshal(data)

	if err != nil {
		ctx.Logger().Error("ProcessSchedule with empty address", err)
		return
	}

	info := k.GetActivityInfo(ctx, addr)

	if info.ExpireAt > ctx.BlockHeight() {
		k.ScheduleRenew(ctx, addr, ctx.BlockHeight()+oneMonth)
		k.resetLimits(ctx, addr)
	} else {
		profile := k.profileKeeper.GetProfile(ctx, addr)

		if profile != nil && profile.AutoPay {
			err := k.autoPay(ctx, addr)
			if err != nil {
				ctx.EventManager().EmitEvent(sdk.NewEvent(
					types.EventTypeAutoPayFailed,
					sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
				))
				k.deactivateAccount(ctx, addr, info)
			} else {
				k.ScheduleRenew(ctx, addr, ctx.BlockHeight()+oneMonth)
				k.resetLimits(ctx, addr)
			}
		} else {
			k.deactivateAccount(ctx, addr, info)
		}
	}
}

func (k Keeper) GetPrices(ctx sdk.Context) (course, subscription, vpn, storage, baseStorage, baseVpn uint32) {
	k.paramspace.Get(ctx, types.KeyTokenCourse, &course)
	k.paramspace.Get(ctx, types.KeySubscriptionPrice, &subscription)
	k.paramspace.Get(ctx, types.KeyVPNGbPrice, &vpn)
	k.paramspace.Get(ctx, types.KeyStorageGbPrice, &storage)
	k.paramspace.Get(ctx, types.KeyBaseStorageGb, &baseStorage)
	k.paramspace.Get(ctx, types.KeyBaseVPNGb, &baseVpn)

	return course,
		subscription,
		vpn,
		storage,
		baseStorage,
		baseVpn
}
