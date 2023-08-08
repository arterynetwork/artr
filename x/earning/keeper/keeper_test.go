// +build testing

package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authK "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/earning"
	profileK "github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/referral"
)

func TestEarningKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc      codec.BinaryMarshaler
	ctx      sdk.Context
	k        earning.Keeper
	ak       authK.AccountKeeper
	bk       bank.Keeper
	pk       profileK.Keeper
	rk       referral.Keeper
	storeKey sdk.StoreKey

	bbHeader abci.RequestBeginBlock
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", "%s", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.k = s.app.GetEarningKeeper()
	s.storeKey = s.app.GetKeys()[earning.ModuleName]
	s.ak = s.app.GetAccountKeeper()
	s.bk = s.app.GetBankKeeper()
	s.pk = s.app.GetProfileKeeper()
	s.rk = s.app.GetReferralKeeper()

	s.bbHeader = abci.RequestBeginBlock{
		Header: tmproto.Header{
			ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
		},
	}
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *Suite) TestFlow() {
	genesisTime := s.ctx.BlockTime()
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	user4 := app.DefaultGenesisUsers["user4"]
	user5 := app.DefaultGenesisUsers["user5"]

	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user13"], 100))
	vpnFund := s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.VpnCollectorName)).AmountOf(util.ConfigMainDenom).Int64()
	storageFund := s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.StorageCollectorName)).AmountOf(util.ConfigMainDenom).Int64()

	user2amt := s.bk.GetBalance(s.ctx, user2).AmountOf(util.ConfigMainDenom).Int64()
	user3amt := s.bk.GetBalance(s.ctx, user3).AmountOf(util.ConfigMainDenom).Int64()
	user4amt := s.bk.GetBalance(s.ctx, user4).AmountOf(util.ConfigMainDenom).Int64()

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

	s.NoError(s.k.Run(s.ctx, util.NewFraction(1, 4), 2, earning.NewPoints(40, 60), genesisTime.Add(5*30*time.Second)))
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

	for h := 1; h < 5; h++ {
		s.nextBlock()
		s.Equal(user2amt, s.bk.GetBalance(s.ctx, user2).AmountOf(util.ConfigMainDenom).Int64(), "user2 at block height %d", h)
		s.Equal(user3amt, s.bk.GetBalance(s.ctx, user3).AmountOf(util.ConfigMainDenom).Int64(), "user3 at block height %d", h)
		s.Equal(user4amt, s.bk.GetBalance(s.ctx, user4).AmountOf(util.ConfigMainDenom).Int64(), "user4 at block height %d", h)
		s.Equal(
			earning.ErrLocked,
			s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
		)
	}

	s.nextBlock()
	s.Equal(
		user2amt+util.NewFraction(1, 16).MulInt64(vpnFund).Int64(),
		s.bk.GetBalance(s.ctx, user2).AmountOf(util.ConfigMainDenom).Int64(),
		"user2 at block height 5",
	)
	s.Equal(
		user3amt,
		s.bk.GetBalance(s.ctx, user3).AmountOf(util.ConfigMainDenom).Int64(),
		"user3 at block height 5",
	)
	s.InDelta(
		user4amt+util.NewFraction(3, 16).MulInt64(vpnFund).Int64()+util.NewFraction(1, 6).MulInt64(storageFund).Int64(),
		s.bk.GetBalance(s.ctx, user4).AmountOf(util.ConfigMainDenom).Int64(),
		1.1,
		"user4 at block height 5",
	)
	s.Equal(
		earning.ErrLocked,
		s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
	)

	s.nextBlock()
	s.Equal(
		user2amt+util.NewFraction(1, 16).MulInt64(vpnFund).Int64(),
		s.bk.GetBalance(s.ctx, user2).AmountOf(util.ConfigMainDenom).Int64(),
		"user2 at block height 6",
	)
	s.Equal(
		user3amt+util.NewFraction(1, 12).MulInt64(storageFund).Int64(),
		s.bk.GetBalance(s.ctx, user3).AmountOf(util.ConfigMainDenom).Int64(),
		"user3 at block height 6",
	)
	s.InDelta(
		user4amt+util.NewFraction(3, 16).MulInt64(vpnFund).Int64()+util.NewFraction(1, 6).MulInt64(storageFund).Int64(),
		s.bk.GetBalance(s.ctx, user4).AmountOf(util.ConfigMainDenom).Int64(),
		1.1,
		"user4 at block height 6",
	)
	s.NoError(
		s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user5, 3, 3)}),
	)
}

func (s *Suite) TestReset() {
	user2 := app.DefaultGenesisUsers["user2"]
	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user13"], 100))

	// Unlocked
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 10, 0)}))
	s.k.Reset(s.ctx)
	s.Empty(s.k.GetEarners(s.ctx))

	// Locked
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 10, 0)}))
	s.NoError(s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(10, 0), s.ctx.BlockTime().Add(100*30*time.Second)))
	s.k.Reset(s.ctx)
	s.Empty(s.k.GetEarners(s.ctx))
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 10, 0)}))
}

func (s *Suite) TestNoMoney() {
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(app.DefaultGenesisUsers["user2"], 10, 0)}))
	s.Equal(
		earning.ErrNoMoney,
		s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(10, 0), s.ctx.BlockTime().Add(100*30*time.Second)),
	)

	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user13"], 10))
	s.NoError(
		s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(10, 0), s.ctx.BlockTime().Add(100*30*time.Second)),
	)
}

func (s *Suite) TestLowMoney() {
	genesisTime := s.ctx.BlockTime()
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]

	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user13"], 5))
	// vpnFund     := s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.VpnCollectorName)).AmountOf(util.ConfigMainDenom).Int64()
	// storageFund := s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.StorageCollectorName)).AmountOf(util.ConfigMainDenom).Int64()
	// => vpnFund = 3_968060, storageFund = 7_936120
	prettyMuch := util.NewFraction(99_999999, 100_000000)

	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 1, 0),
		earning.NewEarner(user3, 0, 1),
	}))
	s.NoError(s.k.Run(s.ctx, prettyMuch, 100, earning.NewPoints(1, 1), genesisTime.Add(30*time.Second)))
	s.nextBlock()

	s.Equal(int64(1), s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.VpnCollectorName)).AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(1), s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.StorageCollectorName)).AmountOf(util.ConfigMainDenom).Int64())

	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 1, 1),
		earning.NewEarner(user3, 1, 1),
	}))
	s.Equal(
		earning.ErrNoMoney,
		s.k.Run(s.ctx, prettyMuch, 100, earning.NewPoints(2, 2), genesisTime.Add(2*30*time.Second)),
	)
	s.NoError(s.k.Run(s.ctx, util.FractionInt(1), 100, earning.NewPoints(2, 2), genesisTime.Add(2*30*time.Second)))
	s.nextBlock()

	// Assert nothing's payed, but list's unlocked
	s.Equal(int64(0), s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.VpnCollectorName)).AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(0), s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.StorageCollectorName)).AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(2), s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.ModuleName)).AmountOf(util.ConfigMainDenom).Int64())
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 1, 1)}))
}

func (s *Suite) TestHeightCheck() {
	genesisTime := s.ctx.BlockTime()
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]

	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user13"], 5))
	s.nextBlock()
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user2, 1, 1)}))
	s.Equal(earning.ErrTooLate, s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(1, 1), genesisTime.Add(29*time.Second)))
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{earning.NewEarner(user3, 1, 1)}))
	s.NoError(s.k.Run(s.ctx, util.NewFraction(7, 30), 100, earning.NewPoints(2, 2), genesisTime.Add(3*30*time.Second)))
}

func (s *Suite) TestEmptyList() {
	genesisTime := s.ctx.BlockTime()
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]

	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user13"], 5))
	vpnFund := s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.VpnCollectorName)).AmountOf(util.ConfigMainDenom).Int64()
	storageFund := s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.StorageCollectorName)).AmountOf(util.ConfigMainDenom).Int64()
	quarter := util.NewFraction(1, 4)
	s.NoError(s.k.Run(s.ctx, quarter, 100, earning.NewPoints(0, 0), genesisTime.Add(30*time.Second)))
	s.nextBlock()

	s.Equal(
		quarter.MulInt64(vpnFund).Int64()+quarter.MulInt64(storageFund).Int64(),
		s.bk.GetBalance(s.ctx, s.ak.GetModuleAddress(earning.ModuleName)).AmountOf(util.ConfigMainDenom).Int64(),
	)

	user2amt := s.bk.GetBalance(s.ctx, user2).AmountOf(util.ConfigMainDenom).Int64()
	user3amt := s.bk.GetBalance(s.ctx, user3).AmountOf(util.ConfigMainDenom).Int64()
	s.NoError(s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user2, 1, 1),
		earning.NewEarner(user3, 1, 1),
	}))
	s.NoError(s.k.Run(s.ctx, quarter, 100, earning.NewPoints(2, 2), genesisTime.Add(2*30*time.Second)))
	s.nextBlock()

	frac := util.NewFraction(7, 32) // (1/4 + 1/4 * (1 - 1/4)) / 2
	s.InDelta(
		user2amt+frac.MulInt64(vpnFund).Int64()+frac.MulInt64(storageFund).Int64(),
		s.bk.GetBalance(s.ctx, user2).AmountOf(util.ConfigMainDenom).Int64(),
		1., // 1 uARTR - possible rounding error
	)
	s.InDelta(
		user3amt+frac.MulInt64(vpnFund).Int64()+frac.MulInt64(storageFund).Int64(),
		s.bk.GetBalance(s.ctx, user3).AmountOf(util.ConfigMainDenom).Int64(),
		1., // 1 uARTR - possible rounding error
	)
}

func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
	bbr := s.app.BeginBlocker(s.ctx, s.bbHeader)
	return ebr, bbr
}
