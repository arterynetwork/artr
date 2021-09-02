//+build testing

package keeper_test

import (
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
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

func (s *SSuite) TestVpnLimit() {
	genesisTime := s.ctx.BlockTime()
	_, _, addr := testdata.KeyTestPubAddr()
	s.NoError(s.k.CreateAccount(s.ctx, addr, app.DefaultGenesisUsers["user1"]))

	p := s.k.GetProfile(s.ctx, addr)
	s.Nil(p.ActiveUntil)
	s.NoError(s.bk.AddCoins(s.ctx, addr, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)))))

	s.NoError(s.k.PayTariff(s.ctx, addr, 5))
	p = s.k.GetProfile(s.ctx, addr)
	s.NotNil(p.ActiveUntil)
	s.Equal(genesisTime.Add(30*24*time.Hour), *p.ActiveUntil)
}
