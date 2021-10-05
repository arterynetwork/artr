//+build testing

package keeper_test

import (
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

	s.NoError(s.k.PayTariff(s.ctx, addr, 5))
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
	s.False(p.AutoPay)

	s.NoError(s.bk.AddCoins(s.ctx, addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)))))
	p.AutoPay = true
	s.NoError(s.k.SetProfile(s.ctx, addr, p))

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.True(p.IsActive(s.ctx))
	s.True(p.AutoPay)

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(wasPaidUpTo.Add(-28*time.Second))
	s.nextBlock()

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(s.ctx.BlockTime().Add(30*24*time.Hour), *p.ActiveUntil)
	s.True(p.ActiveUntil.After(wasPaidUpTo.Add(30*24*time.Hour)))  // 2 seconds for free
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

	s.ctx = s.ctx.WithBlockHeight(9010).WithBlockTime(wasPaidUpTo.Add(272*time.Second))
	s.nextBlock()

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(wasPaidUpTo, *p.ActiveUntil)
	s.False(p.IsActive(s.ctx))
	s.False(p.AutoPay)

	s.NoError(s.bk.AddCoins(s.ctx, addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)))))
	s.NoError(s.k.PayTariff(s.ctx, addr, 5))

	p = *s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(s.ctx.BlockTime().Add(30*24*time.Hour), *p.ActiveUntil)
}

var bbHeader = abci.RequestBeginBlock{
	Header: tmproto.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}

func (s *SSuite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}
