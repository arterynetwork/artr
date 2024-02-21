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
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		).String(),
		s.k.GetBalance(s.ctx, user).String(),
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2015_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(140_000_000000)),
		).String(),
		sdk.Coins(s.k.GetSupply(s.ctx).Total).String(),
	)

	ri, err := s.rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.EqualValues(21_000_000000, ri.Coins[0].Int64())
	s.EqualValues(20_000_000000, ri.Delegated[0].Int64())

	ri, err = s.rk.Get(s.ctx, parent.String())
	s.NoError(err)
	s.EqualValues(42_000_000000, ri.Coins[1].Int64())
	s.EqualValues(40_000_000000, ri.Delegated[1].Int64())

	_, err = s.k.Burn(sdk.WrapSDKContext(s.ctx), &types.MsgBurn{
		Account: user.String(),
		Amount:  100_000000,
	})
	s.NoError(err)

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(900_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		).String(),
		s.k.GetBalance(s.ctx, user).String(),
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2014_900_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(140_000_000000)),
		).String(),
		sdk.Coins(s.k.GetSupply(s.ctx).Total).String(),
	)

	ri, err = s.rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.EqualValues(20_900_000000, ri.Coins[0].Int64())
	s.EqualValues(20_000_000000, ri.Delegated[0].Int64())

	ri, err = s.rk.Get(s.ctx, parent.String())
	s.NoError(err)
	s.EqualValues(41_900_000000, ri.Coins[1].Int64())
	s.EqualValues(40_000_000000, ri.Delegated[1].Int64())
}

func (s *Suite) TestSendWithBlockedAddress() {
	var (
		user1  = app.DefaultGenesisUsers["user1"]
		user15 = app.DefaultGenesisUsers["user15"]
		coins  = sdk.NewCoins(sdk.NewInt64Coin(util.ConfigMainDenom, 1))
		msg    *types.MsgSend
		sdkCtx = sdk.WrapSDKContext(s.ctx)
		err    error
	)

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		),
		s.k.GetBalance(s.ctx, user1),
	)
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.k.GetBalance(s.ctx, user15),
	)

	msg = types.NewMsgSend(user1, user15, coins)
	_, err = s.k.Send(sdkCtx, msg)
	s.NoError(err)

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(999_999999)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		),
		s.k.GetBalance(s.ctx, user1),
	)
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000001))),
		s.k.GetBalance(s.ctx, user15),
	)

	msg = types.NewMsgSend(user15, user1, coins)
	_, err = s.k.Send(sdkCtx, msg)
	s.Error(err)

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(999_999999)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		),
		s.k.GetBalance(s.ctx, user1),
	)
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000001))),
		s.k.GetBalance(s.ctx, user15),
	)
}
