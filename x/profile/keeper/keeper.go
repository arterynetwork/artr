package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile/types"
)

// Keeper of the profile store
type Keeper struct {
	cdc            codec.BinaryMarshaler
	storeKey       sdk.StoreKey
	aliasStoreKey  sdk.StoreKey
	cardsStoreKey  sdk.StoreKey
	paramspace     types.ParamSubspace
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	referralKeeper types.ReferralKeeper
	scheduleKeeper types.ScheduleKeeper
}

// NewKeeper creates a profile keeper
func NewKeeper(
	cdc codec.BinaryMarshaler,
	key sdk.StoreKey,
	aliasKey sdk.StoreKey,
	cardsKey sdk.StoreKey,
	paramspace types.ParamSubspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	referralKeeper types.ReferralKeeper,
	scheduleKeeper types.ScheduleKeeper,
) Keeper {
	keeper := Keeper{
		cdc:            cdc,
		storeKey:       key,
		aliasStoreKey:  aliasKey,
		cardsStoreKey:  cardsKey,
		paramspace:     paramspace.WithKeyTable(types.ParamKeyTable()),
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		referralKeeper: referralKeeper,
		scheduleKeeper: scheduleKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Get returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetProfile(ctx sdk.Context, addr sdk.AccAddress) *types.Profile {
	store := ctx.KVStore(k.storeKey)

	var item types.Profile

	bz := store.Get(addr)
	if bz == nil {
		return nil
	}

	err := k.cdc.UnmarshalBinaryBare(bz, &item)
	if err != nil {
		panic(err)
	}
	return &item
}

func (k Keeper) GetProfileAccountByNickname(ctx sdk.Context, nickname string) sdk.AccAddress {
	store := ctx.KVStore(k.aliasStoreKey)
	return store.Get([]byte(strings.ToLower(nickname)))
}

func (k Keeper) setProfileAccountByNickname(ctx sdk.Context, nickname string, addr sdk.AccAddress) {
	store := ctx.KVStore(k.aliasStoreKey)
	store.Set([]byte(strings.ToLower(nickname)), addr)
}

func (k Keeper) removeProfileAccountByNickname(ctx sdk.Context, nickname string) {
	store := ctx.KVStore(k.aliasStoreKey)
	store.Delete([]byte(strings.ToLower(nickname)))
}

func (k Keeper) GetProfileAccountByCardNumber(ctx sdk.Context, cardNumber uint64) sdk.AccAddress {
	store := ctx.KVStore(k.cardsStoreKey)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, cardNumber)

	return store.Get(buf)
}

func (k Keeper) setProfileAccountByCardNumber(ctx sdk.Context, cardNumber uint64, addr sdk.AccAddress) {
	store := ctx.KVStore(k.cardsStoreKey)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, cardNumber)

	store.Set(buf, addr)
}

func (k Keeper) removeProfileAccountByCardNumber(ctx sdk.Context, cardNumber uint64) {
	store := ctx.KVStore(k.cardsStoreKey)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, cardNumber)
	store.Delete(buf)
}

func (k Keeper) SetProfile(ctx sdk.Context, addr sdk.AccAddress, profile types.Profile) error {
	// 1 - load current profile
	oldProfile := k.GetProfile(ctx, addr)

	nickname := strings.TrimSpace(profile.Nickname)
	if err := k.ValidateProfileNickname(ctx, addr, nickname); err != nil {
		return errors.Wrap(err, "invalid nickname")
	}

	// If old profile filled
	if oldProfile != nil {
		// Check if nickname changed
		if oldProfile.Nickname != profile.Nickname {
			// Profile.Nickname changed - we need to remove old nickname from store
			if oldProfile.Nickname != "" {
				k.removeProfileAccountByNickname(ctx, oldProfile.Nickname)
			}

			// If new nickname not empty - add it to KVStore
			if nickname != "" {
				if err := k.bankKeeper.SendCoinsFromAccountToModule(
					ctx, addr, auth.FeeCollectorName,
					util.Uartrs(1_000000),
				); err != nil {
					return errors.Wrap(err, "cannot charge a rename fee")
				}
				k.setProfileAccountByNickname(ctx, nickname, addr)
			}
		}

		if profile.CardNumber != 0 {
			if oldProfile.CardNumber != profile.CardNumber {
				if oldProfile.CardNumber != 0 {
					k.removeProfileAccountByCardNumber(ctx, oldProfile.CardNumber)
				}

				k.setProfileAccountByCardNumber(ctx, profile.CardNumber, addr)
			}
		} else {
			profile.CardNumber = oldProfile.CardNumber
		}
	} else {
		// we need to add new nickname to store if not empty
		if nickname != "" {
			k.setProfileAccountByNickname(ctx, nickname, addr)
		}

		if profile.CardNumber != 0 {
			k.setProfileAccountByCardNumber(ctx, profile.CardNumber, addr)
		}
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryBare(&profile)
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)

	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		acc = k.accountKeeper.NewAccountWithAddress(ctx, addr)
		k.accountKeeper.SetAccount(ctx, acc)
	}

	if active := profile.IsActive(ctx); active != (oldProfile != nil && oldProfile.IsActive(ctx)) {
		k.referralKeeper.MustSetActive(ctx, addr.String(), active)
	}

	return nil
}

func (k Keeper) ValidateProfileNickname(ctx sdk.Context, addr sdk.AccAddress, nickname string) error {
	if len(nickname) == 0 {
		return nil
	}

	if strings.HasPrefix(strings.ToTitle(nickname), "ARTR-") {
		return types.ErrNicknamePrefix
	}

	namesake := k.GetProfileAccountByNickname(ctx, nickname)
	if namesake != nil && !namesake.Equals(addr) {
		return types.ErrNicknameAlreadyInUse
	}

	return nil
}

func (k Keeper) CreateAccount(ctx sdk.Context, addr sdk.AccAddress, refAddr sdk.AccAddress) error {
	return k.CreateAccountWithProfile(ctx, addr, refAddr, types.Profile{})
}

func (k Keeper) CreateAccountWithProfile(ctx sdk.Context, addr sdk.AccAddress, refAddr sdk.AccAddress, profile types.Profile) error {
	acc := k.accountKeeper.NewAccountWithAddress(ctx, addr)
	k.accountKeeper.SetAccount(ctx, acc)
	if err := k.referralKeeper.AppendChild(ctx, refAddr.String(), addr.String()); err != nil {
		return errors.Wrap(err, "cannot add account to referral")
	}
	k.referralKeeper.ScheduleCompression(ctx, addr.String(), ctx.BlockTime().Add(k.referralKeeper.CompressionPeriod(ctx)))
	profile.CardNumber = k.CardNumberByAccountNumber(ctx, acc.GetAccountNumber())
	if err := k.SetProfile(ctx, addr, profile); err != nil {
		return errors.Wrap(err, "cannot set profile")
	}
	return nil
}

func (k Keeper) CardNumberByAccountNumber(ctx sdk.Context, accNumber uint64) uint64 {
	return accNumber ^ k.GetParams(ctx).CardMagic
}

func (k Keeper) SetStorageCurrent(ctx sdk.Context, addr sdk.AccAddress, value uint64) error {
	profile := k.GetProfile(ctx, addr)
	if profile == nil {
		return types.ErrNotFound
	}
	profile.StorageCurrent = value

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryBare(profile)
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)

	return nil
}

func (k Keeper) SetVpnCurrent(ctx sdk.Context, addr sdk.AccAddress, value uint64) error {
	profile := k.GetProfile(ctx, addr)
	if profile == nil {
		return types.ErrNotFound
	}
	profile.VpnCurrent = value

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryBare(profile)
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)

	return nil
}
