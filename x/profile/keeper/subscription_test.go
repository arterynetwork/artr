// +build testing

package keeper_test

import (
	scheduleK "github.com/arterynetwork/artr/x/schedule/keeper"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/profile/keeper"
)

func TestSubscription(t *testing.T) {
	suite.Run(t, new(SSuite))
}

type SSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	cdc     codec.BinaryMarshaler
	ctx     sdk.Context
	k       keeper.Keeper
	bk      bank.Keeper
	sk      scheduleK.Keeper

	bbHeader abci.RequestBeginBlock
}

func (s *SSuite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.k = s.app.GetProfileKeeper()
	s.bk = s.app.GetBankKeeper()
	s.sk = s.app.GetScheduleKeeper()

	s.bbHeader = abci.RequestBeginBlock{
		Header: tmproto.Header{
			ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
		},
	}
}

func (s *SSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *SSuite) TestPayTariffInAdvance() {
	wasPaidUpTo, _ := time.Parse(time.RFC3339, "2022-01-04T03:00:00Z")
	addr := app.DefaultGenesisUsers["user1"]

	p := s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.True(p.IsActive(s.ctx))
	s.NoError(s.bk.AddCoins(s.ctx, addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)))))

	s.NoError(s.k.PayTariff(s.ctx, addr, 5, false))
	p = s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo.Add(30*24*time.Hour), *p.ActiveUntil)
	s.True(p.IsActive(s.ctx))
}

func (s *SSuite) TestAutoPay() {
	wasPaidUpTo, _ := time.Parse(time.RFC3339, "2022-01-04T03:00:00Z")
	addr := app.DefaultGenesisUsers["user1"]

	p := *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.True(p.IsActive(s.ctx))
	s.False(p.AutoPay)

	s.NoError(s.bk.AddCoins(s.ctx, addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)))))
	p.AutoPay = true
	s.NoError(s.k.SetProfile(s.ctx, addr, p))

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.True(p.IsActive(s.ctx))
	s.True(p.AutoPay)

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(wasPaidUpTo.Add(-28 * time.Second))
	s.nextBlock()

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.NotEqual(wasPaidUpTo.Add(30*24*time.Hour), *p.ActiveUntil)
	s.Equal(s.ctx.BlockTime().Add(30*24*time.Hour), *p.ActiveUntil)
	s.True(p.ActiveUntil.After(wasPaidUpTo.Add(30 * 24 * time.Hour))) // 2 seconds for free
	s.True(p.IsActive(s.ctx))
	s.True(p.AutoPay)
}

func (s *SSuite) TestPayTariffWhenItIsOver() {
	wasPaidUpTo, _ := time.Parse(time.RFC3339, "2022-01-04T03:00:00Z")
	addr := app.DefaultGenesisUsers["user1"]

	p := *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.True(p.IsActive(s.ctx))
	s.False(p.AutoPay)

	s.ctx = s.ctx.WithBlockHeight(9010).WithBlockTime(wasPaidUpTo.Add(272 * time.Second))
	s.nextBlock()

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.False(p.IsActive(s.ctx))
	s.False(p.AutoPay)

	s.NoError(s.bk.AddCoins(s.ctx, addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)))))
	s.NoError(s.k.PayTariff(s.ctx, addr, 5, false))

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(s.ctx.BlockTime().Add(30*24*time.Hour), *p.ActiveUntil)
}

func (s *SSuite) TestPayTariff_ExtraStorage() {
	addr := app.DefaultGenesisUsers["user1"]
	s.NoError(s.k.BuyStorage(s.ctx, addr, 13))
	s.Equal(uint64((5+13)*util.GBSize), s.k.GetProfile(s.ctx, addr).StorageLimit)
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()

	s.NoError(s.k.PayTariff(s.ctx, addr, 0, false))
	s.Equal(uint64((5+13)*util.GBSize), s.k.GetProfile(s.ctx, addr).StorageLimit)
	s.EqualValues(
		balance-(1990+13*10)*100000,
		s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64(),
	)
}

func (s *SSuite) TestBuyStorage_TiB() {
	addr := app.DefaultGenesisUsers["user1"]
	s.Equal(uint64(0), s.k.GetProfile(s.ctx, addr).StorageLimit)

	s.NoError(s.k.BuyStorage(s.ctx, addr, 1024))

	s.Equal(uint64((5+1024)*util.GBSize), s.k.GetProfile(s.ctx, addr).StorageLimit)
}

func (s *SSuite) TestBuyStorage_PiB() {
	addr := app.DefaultGenesisUsers["user1"]
	s.bk.AddCoins(s.ctx, addr, util.Uartrs(200_000_000000))
	s.Equal(uint64(0), s.k.GetProfile(s.ctx, addr).StorageLimit)

	s.NoError(s.k.BuyStorage(s.ctx, addr, 1048576))

	s.Equal(uint64((5+1048576)*util.GBSize), s.k.GetProfile(s.ctx, addr).StorageLimit)
}

func (s *SSuite) TestGiveUpStorage_NegativeDelta() {
	addr := app.DefaultGenesisUsers["user1"]
	amount := s.k.GetProfile(s.ctx, addr).StorageLimit

	s.Error(s.k.GiveStorageUp(s.ctx, addr, 100500))

	s.Equal(amount, s.k.GetProfile(s.ctx, addr).StorageLimit)
}

func (s *SSuite) TestImExtra_InitialState() {
	addr := app.DefaultGenesisUsers["user1"]
	p := s.k.GetProfile(s.ctx, addr)
	s.False(p.IsExtraImStorageActive(s.ctx))
	s.Zero(p.ImLimitExtra)
	s.EqualValues(5*util.GBSize, p.ImLimitTotal(s.ctx))
	s.Nil(p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_Buy() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))

	p := s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
	s.EqualValues(balance-100000*13*10, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())
}

func (s *SSuite) TestImExtra_Buy_SaleOff() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()
	var price1 int64 = 100000 * 13 * 10
	var price2 int64 = 100000 * 6 * 10 / 3

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(20*24*time.Hour - 30*time.Second))
	s.nextBlock()
	balance += price1 * 3 / 1000 // TX fee
	s.EqualValues(balance-price1, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 6))
	s.EqualValues(balance-price1-price2, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	p := s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(19, p.ImLimitExtra)
	s.EqualValues(24*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_Expiration() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))
	s.True(s.k.GetProfile(s.ctx, addr).IsExtraImStorageActive(s.ctx))

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(30*24*time.Hour - time.Second))
	s.nextBlock()

	p := s.k.GetProfile(s.ctx, addr)
	s.False(p.IsExtraImStorageActive(s.ctx))
	s.Zero(p.ImLimitExtra)
	s.EqualValues(5*util.GBSize, p.ImLimitTotal(s.ctx))
	s.Nil(p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_Prolong() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()
	var price int64 = 100000 * 13 * 10

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))
	s.True(s.k.GetProfile(s.ctx, addr).IsExtraImStorageActive(s.ctx))

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(15 * 24 * time.Hour))
	s.nextBlock()
	balance += price * 3 / 1000 // TX fee
	s.True(s.k.GetProfile(s.ctx, addr).IsExtraImStorageActive(s.ctx))
	s.EqualValues(balance-price, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	s.NoError(s.k.ProlongImExtra(s.ctx, addr))
	p := s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.Equal(genesisTime.Add(2*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
	s.EqualValues(balance-2*price, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(30*24*time.Hour - 59*time.Second))
	s.nextBlock()
	p = s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(2*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)

	s.nextBlock()
	p = s.k.GetProfile(s.ctx, addr)
	p = s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(2*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(2*30*24*time.Hour - 59*time.Second))
	s.nextBlock()
	p = s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(2*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)

	s.nextBlock()
	p = s.k.GetProfile(s.ctx, addr)
	s.False(p.IsExtraImStorageActive(s.ctx))
	s.Zero(p.ImLimitExtra)
	s.EqualValues(5*util.GBSize, p.ImLimitTotal(s.ctx))
	s.Nil(p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_Prolong_Nothing() {
	addr := app.DefaultGenesisUsers["user1"]
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()

	s.False(s.k.GetProfile(s.ctx, addr).IsExtraImStorageActive(s.ctx))
	s.Error(s.k.ProlongImExtra(s.ctx, addr))

	p := s.k.GetProfile(s.ctx, addr)
	s.False(p.IsExtraImStorageActive(s.ctx))
	s.Zero(p.ImLimitExtra)
	s.EqualValues(5*util.GBSize, p.ImLimitTotal(s.ctx))
	s.Nil(p.ExtraImUntil)
	s.EqualValues(balance, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())
}

func (s *SSuite) TestImExtra_GiveUp_Part() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))
	s.nextBlock()
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()

	s.NoError(s.k.GiveImStorageUp(s.ctx, addr, 6))
	s.EqualValues(balance, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())
	p := s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(6, p.ImLimitExtra)
	s.EqualValues(11*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_GiveUp_All() {
	addr := app.DefaultGenesisUsers["user1"]

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))
	s.nextBlock()
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()

	s.NoError(s.k.GiveImStorageUp(s.ctx, addr, 0))
	s.EqualValues(balance, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())
	p := s.k.GetProfile(s.ctx, addr)
	s.False(p.IsExtraImStorageActive(s.ctx))
	s.Zero(p.ImLimitExtra)
	s.EqualValues(5*util.GBSize, p.ImLimitTotal(s.ctx))
	s.Nil(p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_GiveUp_Negative() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))
	s.nextBlock()
	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()

	s.Error(s.k.GiveImStorageUp(s.ctx, addr, 14))
	s.EqualValues(balance, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())
	p := s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
}

func (s *SSuite) TestImExtra_AutoPay() {
	addr := app.DefaultGenesisUsers["user1"]
	genesisTime := s.ctx.BlockTime()

	s.NoError(s.k.BuyImStorage(s.ctx, addr, 13))
	p := s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	p.AutoPayImExtra = true
	s.NoError(s.k.SetProfile(s.ctx, addr, *p))

	balance := s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64()
	var price int64 = 100000 * 13 * 10
	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(30*24*time.Hour - time.Second))
	s.nextBlock()
	balance += price * 3 / 1000 // TX fee

	p = s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(2*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
	s.EqualValues(balance-price, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(2*30*24*time.Hour - time.Second))
	s.nextBlock()
	balance += price * 3 / 1000 // TX fee

	p = s.k.GetProfile(s.ctx, addr)
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(3*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
	s.EqualValues(balance-2*price, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	p.AutoPayImExtra = false
	s.nextBlock()
	balance += price * 3 / 1000 // TX fee

	s.NoError(s.k.SetProfile(s.ctx, addr, *p))
	s.True(p.IsExtraImStorageActive(s.ctx))
	s.EqualValues(13, p.ImLimitExtra)
	s.EqualValues(18*util.GBSize, p.ImLimitTotal(s.ctx))
	s.NotNil(p.ExtraImUntil)
	s.Equal(genesisTime.Add(3*s.sk.OneMonth(s.ctx)), *p.ExtraImUntil)
	s.EqualValues(balance-2*price, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())

	s.ctx = s.ctx.WithBlockTime(genesisTime.Add(3*30*24*time.Hour - time.Second))
	s.nextBlock() // all TX fees have been paid already
	p = s.k.GetProfile(s.ctx, addr)
	s.False(p.IsExtraImStorageActive(s.ctx))
	s.Zero(p.ImLimitExtra)
	s.EqualValues(5*util.GBSize, p.ImLimitTotal(s.ctx))
	s.Nil(p.ExtraImUntil)
	s.EqualValues(balance-2*price, s.bk.GetBalance(s.ctx, addr).AmountOf(util.ConfigMainDenom).Int64())
}

func (s *SSuite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
	bbr := s.app.BeginBlocker(s.ctx, s.bbHeader)
	return ebr, bbr
}
