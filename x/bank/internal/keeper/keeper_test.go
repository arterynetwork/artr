// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/bank/types"
	referralK "github.com/arterynetwork/artr/x/referral/keeper"
)

func TestBankKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context

	cdc codec.BinaryMarshaler
	k   bank.Keeper
	rk  referralK.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.k = s.app.GetBankKeeper()
	s.rk = s.app.GetReferralKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *Suite) TestBurn() {
	var (
		user   = app.DefaultGenesisUsers["user4"]
		parent = app.DefaultGenesisUsers["user2"]
	)

	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))).String(),
		s.k.GetBalance(s.ctx, user).String(),
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2015_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(40_000_000000)),
		).String(),
		sdk.Coins(s.k.GetSupply(s.ctx).Total).String(),
	)

	ri, err := s.rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.EqualValues(1_000_000000, ri.Coins[0].Int64())
	s.EqualValues(0, ri.Delegated[0].Int64())

	ri, err = s.rk.Get(s.ctx, parent.String())
	s.NoError(err)
	s.EqualValues(2_000_000000, ri.Coins[1].Int64())
	s.EqualValues(0, ri.Delegated[1].Int64())

	_, err = s.k.Burn(sdk.WrapSDKContext(s.ctx), &types.MsgBurn{
		Account: user.String(),
		Amount:  100_000000,
	})
	s.NoError(err)

	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(900_000000))).String(),
		s.k.GetBalance(s.ctx, user).String(),
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2014_900_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(40_000_000000)),
		).String(),
		sdk.Coins(s.k.GetSupply(s.ctx).Total).String(),
	)

	ri, err = s.rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.EqualValues(900_000000, ri.Coins[0].Int64())
	s.EqualValues(0, ri.Delegated[0].Int64())

	ri, err = s.rk.Get(s.ctx, parent.String())
	s.NoError(err)
	s.EqualValues(1_900_000000, ri.Coins[1].Int64())
	s.EqualValues(0, ri.Delegated[1].Int64())
}
