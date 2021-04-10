package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/arterynetwork/artr/x/referral"
)

// Keeper of the profile store
type Keeper struct {
	storeKey        sdk.StoreKey
	aliasStoreKey   sdk.StoreKey
	cardsStoreKey   sdk.StoreKey
	cdc             *codec.Codec
	paramspace      types.ParamSubspace
	AccountKeeper   types.AccountKeeper
	BankKeeper      types.BankKeeper
	ReferralsKeeper types.ReferralsKeeper
	SupplyKeeper    types.SupplyKeeper
}

// NewKeeper creates a profile keeper
func NewKeeper(cdc *codec.Codec,
	key sdk.StoreKey,
	aliasKey sdk.StoreKey,
	cardsKey sdk.StoreKey,
	paramspace types.ParamSubspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	referralsKeeper types.ReferralsKeeper,
	supplyKeeper types.SupplyKeeper,
) Keeper {
	keeper := Keeper{
		storeKey:        key,
		aliasStoreKey:   aliasKey,
		cardsStoreKey:   cardsKey,
		cdc:             cdc,
		paramspace:      paramspace.WithKeyTable(types.ParamKeyTable()),
		AccountKeeper:   accountKeeper,
		BankKeeper:      bankKeeper,
		ReferralsKeeper: referralsKeeper,
		SupplyKeeper:    supplyKeeper,
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

	bz := store.Get(auth.AddressStoreKey(addr))

	if bz == nil {
		return nil
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &item)

	return &item
}

func (k Keeper) GetProfileAccountByNickname(ctx sdk.Context, nickname string) sdk.AccAddress {
	store := ctx.KVStore(k.aliasStoreKey)
	var addr sdk.AccAddress

	bz := store.Get([]byte(strings.ToLower(nickname)))

	if bz == nil {
		return nil
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &addr)

	return addr
}

func (k Keeper) setProfileAccountByNickname(ctx sdk.Context, nickname string, addr sdk.AccAddress) {
	store := ctx.KVStore(k.aliasStoreKey)
	bz := k.cdc.MustMarshalBinaryBare(addr)
	store.Set([]byte(strings.ToLower(nickname)), bz)
}

func (k Keeper) removeProfileAccountByNickname(ctx sdk.Context, nickname string) {
	store := ctx.KVStore(k.aliasStoreKey)
	store.Delete([]byte(strings.ToLower(nickname)))
}

func (k Keeper) GetProfileAccountByCardNumber(ctx sdk.Context, cardNumber uint64) sdk.AccAddress {
	store := ctx.KVStore(k.cardsStoreKey)
	var addr sdk.AccAddress

	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, cardNumber)

	bz := store.Get(buf)

	if bz == nil {
		return nil
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &addr)

	return addr
}

func (k Keeper) setProfileAccountByCardNumber(ctx sdk.Context, cardNumber uint64, addr sdk.AccAddress) {
	store := ctx.KVStore(k.cardsStoreKey)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, cardNumber)

	bz := k.cdc.MustMarshalBinaryBare(addr)
	store.Set(buf, bz)
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
				if err := k.SupplyKeeper.SendCoinsFromAccountToModule(
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
	bz := k.cdc.MustMarshalBinaryBare(profile)
	store.Set(auth.AddressStoreKey(addr), bz)

	acc := k.AccountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		acc = k.AccountKeeper.NewAccountWithAddress(ctx, addr)
		k.AccountKeeper.SetAccount(ctx, acc)
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
	acc := k.AccountKeeper.NewAccountWithAddress(ctx, addr)
	if err := acc.SetCoins(sdk.NewCoins(sdk.NewCoin(types.MainDenom, sdk.NewInt(0)))); err != nil {
		return errors.Wrap(err, "cannot initialize account balance")
	}
	k.AccountKeeper.SetAccount(ctx, acc)
	if err := k.ReferralsKeeper.AppendChild(ctx, refAddr, addr); err != nil {
		return errors.Wrap(err, "cannot add account to referral")
	}
	if err := k.ReferralsKeeper.ScheduleCompression(ctx, addr, ctx.BlockHeight()+referral.CompressionPeriod); err != nil {
		return errors.Wrap(err, "cannot schedule compression")
	}
	profile.CardNumber = k.CardNumberByAccountNumber(ctx, acc.GetAccountNumber())
	if err := k.SetProfile(ctx, addr, profile); err != nil {
		return errors.Wrap(err, "cannot set profile")
	}
	return nil
}

func (k Keeper) CardNumberByAccountNumber(ctx sdk.Context, accNumber uint64) uint64 {
	return accNumber ^ k.GetParams(ctx).CardMagic
}
