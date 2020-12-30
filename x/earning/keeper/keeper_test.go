// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/earning"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/vpn"
)

func TestEarningKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc       *codec.Codec
	ctx       sdk.Context
	k         earning.Keeper
	storeKey  sdk.StoreKey
	accKeeper auth.AccountKeeper
	refKeeper referral.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetEarningKeeper()
	s.storeKey = s.app.GetKeys()[earning.ModuleName]
	s.accKeeper = s.app.GetAccountKeeper()
	s.refKeeper = s.app.GetReferralKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s *Suite) TestFlow() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	user4 := app.DefaultGenesisUsers["user4"]
	user5 := app.DefaultGenesisUsers["user5"]

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user13"], 100*util.GBSize))
	vpnFund := s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	storageFund := s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()

	user2amt := s.accKeeper.GetAccount(s.ctx, user2).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	user3amt := s.accKeeper.GetAccount(s.ctx, user3).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	user4amt := s.accKeeper.GetAccount(s.ctx, user4).GetCoins().AmountOf(util.ConfigMainDenom).Int64()

	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 10, 0),
	}))
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user3, 0, 20),
		earning.NewEarner(user4, 30, 40),
	}))
	s.Equal(
		earning.ErrAlreadyListed,
		s.k.ListEarners(s.ctx, []earning.Earner{
			earning.NewEarner(user2, 10, 0),
			earning.NewEarner(user5, 3, 3),
		}),
	)
	s.Equal([]earning.Earner{
		// Actual items are sorted naturally by account address bytes
		earning.NewEarner(user2, 10, 0),
		earning.NewEarner(user4, 30, 40),
		earning.NewEarner(user3, 0, 20),
	}, s.k.GetEarners(s.ctx))

	s.NoError(s.k.Run(s.ctx, util.NewFraction(1, 4), 2, earning.NewPoints(40, 60), 5))
	s.Equal(
		earning.ErrLocked,
		s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
	)
	s.Equal([]earning.Earner{
		// Actual items are sorted naturally by account address bytes
		earning.NewEarner(user2, 10, 0),
		earning.NewEarner(user4, 30, 40),
		earning.NewEarner(user3, 0, 20),
	}, s.k.GetEarners(s.ctx))

	for h := 2; h < 5; h++ {
		s.nextBlock()
		s.Equal(user2amt, s.accKeeper.GetAccount(s.ctx, user2).GetCoins().AmountOf(util.ConfigMainDenom).Int64(), "user2 at block height %d", h)
		s.Equal(user3amt, s.accKeeper.GetAccount(s.ctx, user3).GetCoins().AmountOf(util.ConfigMainDenom).Int64(), "user3 at block height %d", h)
		s.Equal(user4amt, s.accKeeper.GetAccount(s.ctx, user4).GetCoins().AmountOf(util.ConfigMainDenom).Int64(), "user4 at block height %d", h)
		s.Equal(
			earning.ErrLocked,
			s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
		)
	}

	s.nextBlock()
	s.Equal(
		user2amt+util.NewFraction(1, 16).MulInt64(vpnFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user2).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
		"user2 at block height 5",
	)
	s.Equal(
		user3amt,
		s.accKeeper.GetAccount(s.ctx, user3).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
		"user3 at block height 5",
	)
	s.Equal(
		user4amt+util.NewFraction(3, 16).MulInt64(vpnFund).Int64()+util.NewFraction(1, 6).MulInt64(storageFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user4).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
		"user4 at block height 5",
	)
	s.Equal(
		earning.ErrLocked,
		s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
	)

	s.nextBlock()
	s.Equal(
		user2amt+util.NewFraction(1, 16).MulInt64(vpnFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user2).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
		"user2 at block height 6",
	)
	s.Equal(
		user3amt+util.NewFraction(1, 12).MulInt64(storageFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user3).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
		"user3 at block height 6",
	)
	s.Equal(
		user4amt+util.NewFraction(3, 16).MulInt64(vpnFund).Int64()+util.NewFraction(1, 6).MulInt64(storageFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user4).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
		"user4 at block height 6",
	)
	s.NoError(
		s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
	)
}

func (s *Suite) TestReset() {
	user2 := app.DefaultGenesisUsers["user2"]
	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user13"], 100*util.GBSize))

	// Unlocked
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 10, 0)}))
	s.k.Reset(s.ctx)
	s.Empty(s.k.GetEarners(s.ctx))

	// Locked
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 10, 0)}))
	s.NoError(s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(10, 0), 100))
	s.k.Reset(s.ctx)
	s.Empty(s.k.GetEarners(s.ctx))
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 10, 0)}))
}

func (s *Suite) TestNoMoney() {
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(app.DefaultGenesisUsers["user2"], 10, 0)}))
	s.Equal(
		earning.ErrNoMoney,
		s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(10, 0), 100),
	)

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user13"], 10*util.GBSize))
	s.NoError(
		s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(10, 0), 100),
	)
}

func (s *Suite) TestLowMoney() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user13"], util.GBSize))
	// vpnFund     := s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	// storageFund := s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	// => vpnFund = 3_968060, storageFund = 7_936120
	prettyMuch := util.NewFraction(99_999999, 100_000000)

	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 1, 0),
		earning.NewEarner(user3, 0, 1),
	}))
	s.NoError(s.k.Run(s.ctx, prettyMuch, 100, earning.NewPoints(1, 1), 2))
	s.nextBlock()

	s.Equal(int64(1), s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(1), s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64())

	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 1, 1),
		earning.NewEarner(user3, 1, 1),
	}))
	s.Equal(
		earning.ErrNoMoney,
		s.k.Run(s.ctx, prettyMuch, 100, earning.NewPoints(2, 2), 3),
	)
	s.NoError(s.k.Run(s.ctx, util.FractionInt(1), 100, earning.NewPoints(2, 2), 3))
	s.nextBlock()

	// Assert nothing's payed, but list's unlocked
	s.Equal(int64(0), s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(0), s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(2), s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, earning.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64())
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 1, 1)}))
}

func (s *Suite) TestHeightCheck() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user13"], util.GBSize))
	s.nextBlock()
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 1, 1)}))
	s.Equal(earning.ErrTooLate, s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(1, 1), 2))
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user3, 1, 1)}))
	s.NoError(s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(2, 2), 3))
}

func (s *Suite) TestEmptyList() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user13"], util.GBSize))
	vpnFund := s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	storageFund := s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	quarter := util.NewFraction(1, 4)
	s.NoError(s.k.Run(s.ctx, quarter, 100, earning.NewPoints(0, 0), 2))
	s.nextBlock()

	s.Equal(
		quarter.MulInt64(vpnFund).Int64()+quarter.MulInt64(storageFund).Int64(),
		s.app.GetSupplyKeeper().GetModuleAccount(s.ctx, earning.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	user2amt := s.accKeeper.GetAccount(s.ctx, user2).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	user3amt := s.accKeeper.GetAccount(s.ctx, user3).GetCoins().AmountOf(util.ConfigMainDenom).Int64()
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 1, 1),
		earning.NewEarner(user3, 1, 1),
	}))
	s.NoError(s.k.Run(s.ctx, quarter, 100, earning.NewPoints(2, 2), 3))
	s.nextBlock()

	frac := util.NewFraction(7, 32) // (1/4 + 1/4 * (1 - 1/4)) / 2
	s.Equal(
		user2amt+frac.MulInt64(vpnFund).Int64()+frac.MulInt64(storageFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user2).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		user3amt+frac.MulInt64(vpnFund).Int64()+frac.MulInt64(storageFund).Int64(),
		s.accKeeper.GetAccount(s.ctx, user3).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
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
