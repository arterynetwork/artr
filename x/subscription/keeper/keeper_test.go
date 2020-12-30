// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/subscription"
	"github.com/arterynetwork/artr/x/vpn"
)

func TestSubscriptionKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc           *codec.Codec
	ctx           sdk.Context
	k             subscription.Keeper
	storeKey      sdk.StoreKey
	accKeeper     auth.AccountKeeper
	supplyKeeper  supply.Keeper
	profileKeeper profile.Keeper
	storageKeeper storage.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetSubscriptionKeeper()
	s.storeKey = s.app.GetKeys()[referral.ModuleName]
	s.accKeeper = s.app.GetAccountKeeper()
	s.supplyKeeper = s.app.GetSupplyKeeper()
	s.profileKeeper = s.app.GetProfileKeeper()
	s.storageKeeper = s.app.GetStorageKeeper()
}

func (s Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestPayment() {
	s.NoError(s.k.PayForSubscription(s.ctx, app.DefaultGenesisUsers["user1"], 5*util.GBSize))
	vpn := s.supplyKeeper.GetModuleAccount(s.ctx, vpn.ModuleName)
	storage := s.supplyKeeper.GetModuleAccount(s.ctx, storage.ModuleName)
	s.Equal(sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(3968060))), vpn.GetCoins())
	s.Equal(sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(7936120))), storage.GetCoins())
}

func (s Suite) TestAutoPayment() {
	info := s.k.GetActivityInfo(s.ctx, app.DefaultGenesisUsers["user1"])
	info.ExpireAt = 0
	info.Active = false
	s.k.SetActivityInfo(s.ctx, app.DefaultGenesisUsers["user1"], info)

	s.NoError(s.k.PayForSubscription(s.ctx, app.DefaultGenesisUsers["user1"], 5*util.GBSize))
	profile := s.profileKeeper.GetProfile(s.ctx, app.DefaultGenesisUsers["user1"])
	profile.AutoPay = true
	s.profileKeeper.SetProfile(s.ctx, app.DefaultGenesisUsers["user1"], *profile)
	//s.NoError(s.k.PayForSubscription(s.ctx, app.DefaultGenesisUsers["user1"], 5*util.GBSize))
	info = s.k.GetActivityInfo(s.ctx, app.DefaultGenesisUsers["user1"])

	s.Equal(int64(util.BlocksOneMonth+1), info.ExpireAt)
	s.Equal(true, info.Active)

	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneMonth)
	s.nextBlock()
	info = s.k.GetActivityInfo(s.ctx, app.DefaultGenesisUsers["user1"])

	s.Equal(int64(util.BlocksOneMonth*2+1), info.ExpireAt)
	s.Equal(true, info.Active)

	profile.AutoPay = false
	s.profileKeeper.SetProfile(s.ctx, app.DefaultGenesisUsers["user1"], *profile)

	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneMonth * 2)
	s.nextBlock()
	info = s.k.GetActivityInfo(s.ctx, app.DefaultGenesisUsers["user1"])
	s.Equal(int64(util.BlocksOneMonth*2+1), info.ExpireAt)
	s.Equal(false, info.Active)
}

func (s Suite) TestAutoPayment_StorageLimitRegression() {
	user := app.DefaultGenesisUsers["user1"]

	info := s.k.GetActivityInfo(s.ctx, user)
	info.ExpireAt = 0
	info.Active = false
	s.k.SetActivityInfo(s.ctx, user, info)

	s.NoError(s.k.PayForSubscription(s.ctx, user, 5*util.GBSize))
	profile := s.profileKeeper.GetProfile(s.ctx, user)
	profile.AutoPay = true
	s.profileKeeper.SetProfile(s.ctx, user, *profile)

	s.storageKeeper.SetLimit(s.ctx, user, 20*util.GBSize)
	s.storageKeeper.SetCurrent(s.ctx, user, 4*util.GBSize)

	s.Equal(
		util.Uartrs(3968060),
		s.supplyKeeper.GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins(),
	)
	s.Equal(
		util.Uartrs(7936120),
		s.supplyKeeper.GetModuleAccount(s.ctx, storage.ModuleName).GetCoins(),
	)

	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneMonth)
	s.nextBlock()

	s.Equal(int64(20*util.GBSize), s.storageKeeper.GetLimit(s.ctx, user))
	s.Equal(int64(4*util.GBSize), s.storageKeeper.GetCurrent(s.ctx, user))
	s.Equal(
		util.Uartrs(2*3968060),
		s.supplyKeeper.GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins(),
	)
	s.Equal(
		util.Uartrs(2*7936120+15*997000),
		s.supplyKeeper.GetModuleAccount(s.ctx, storage.ModuleName).GetCoins(),
	)
}

func (s Suite) TestAutoPayment_InsufficientFunds() {
	user := app.DefaultGenesisUsers["user8"]
	info := s.k.GetActivityInfo(s.ctx, user)
	info.ExpireAt = 0
	info.Active = false
	s.k.SetActivityInfo(s.ctx, user, info)

	s.NoError(s.app.GetBankKeeper().SetCoins(s.ctx, user, util.Uartrs(412_999999))) // = 199 + 214 - o(1)

	s.NoError(s.k.PayForSubscription(s.ctx, user, 5*util.GBSize)) // -199 ARTR
	profile := s.profileKeeper.GetProfile(s.ctx, user)
	profile.AutoPay = true
	s.profileKeeper.SetProfile(s.ctx, user, *profile)

	s.storageKeeper.SetLimit(s.ctx, user, 20*util.GBSize)
	s.storageKeeper.SetCurrent(s.ctx, user, 4*util.GBSize)

	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneMonth)
	s.nextBlock() // -214 ARTR

	s.Equal(util.Uartrs(213_999999), s.accKeeper.GetAccount(s.ctx, user).GetCoins())
}

// ----- private functions ------------

func (s *Suite) setBalance(acc sdk.AccAddress, coins sdk.Coins) error {
	item := s.accKeeper.GetAccount(s.ctx, acc)
	if item == nil {
		item = s.accKeeper.NewAccountWithAddress(s.ctx, acc)
	}
	err := item.SetCoins(coins)
	if err != nil {
		return err
	}
	s.accKeeper.SetAccount(s.ctx, item)
	return nil
}

func (s *Suite) get(acc sdk.AccAddress) (types.R, error) {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.R
	err := s.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &value)
	return value, err
}

func (s *Suite) set(acc sdk.AccAddress, value types.R) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := s.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (s *Suite) update(acc sdk.AccAddress, callback func(*types.R)) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.R
	err := s.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &value)
	if err != nil {
		return err
	}
	callback(&value)
	valueBytes, err = s.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

var bbHeader = abci.RequestBeginBlock{
	Header: abci.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}

func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}
