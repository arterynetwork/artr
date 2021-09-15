// +build testing

package keeper_test

import (
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/delegating"
)

func TestDelegatingKeeper(t *testing.T) { suite.Run(t, new(Suite)) }

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc       *codec.Codec
	ctx       sdk.Context
	k         delegating.Keeper
	accKeeper auth.AccountKeeper
	bk        bank.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.ctx = s.app.NewContext(true, abci.Header{})
	s.k = s.app.GetDelegatingKeeper()
	s.accKeeper = s.app.GetAccountKeeper()
	s.bk = s.app.GetBankKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s *Suite) TestDelegatingAndRevoking() {
	user := app.DefaultGenesisUsers["user1"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1000000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1000000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(850000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(850000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(850000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
	revoking, err := s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(
		[]types.RevokeRequest{{
			HeightToImplementAt: 14 * 2880,
			MicroCoins:          sdk.NewInt(850000000),
		}},
		revoking,
	)

	s.ctx = s.ctx.WithBlockHeight(14*2880 - 1)
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(850000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
	revoking, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Empty(revoking)
}

func (s *Suite) TestAccrueAfterRevoke() {
	user := app.DefaultGenesisUsers["user1"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(850_000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(350_000000)))
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	t := 0
	for ; t < util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(3_500000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	for ; t < 14*util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(399_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	for ; t < 15*util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(402_500000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
}

func (s *Suite) TestAccrueOnRevoke() {
	user := app.DefaultGenesisUsers["user1"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(850_000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	t := 0

	for ; t < util.BlocksOneDay/2; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(850_000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
	acc, err := s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.Equal(int64(2_975000), acc.CurrentUartrs)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(350_000000)))
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2_975000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	// 2 weeks later
	for ; t < util.BlocksOneDay*29/2; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(401_975000)), // 2.975 + 14 * 3.5 + 350
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
	acc, err = s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.Equal(int64(util.BlocksOneDay*29/2), acc.StartHeight)
	s.Equal(int64(0), acc.CurrentUartrs)

	// Half a day later
	for ; t < util.BlocksOneDay*15; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(401_975000)), // The same because accrue time has changed
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(500_000000)),
		),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
}

func (s *Suite) TestRevokePeriod() {
	user := app.DefaultGenesisUsers["user2"]

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(1_000000)))

	rrz, err := s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(1, len(rrz))
	s.Equal(
		types.RevokeRequest{
			MicroCoins: sdk.NewInt(1_000000),
			HeightToImplementAt: 40_320,
		}, rrz[0],
	)

	s.nextBlock()
	pz := s.k.GetParams(s.ctx)
	pz.RevokePeriod = 20_160
	s.k.SetParams(s.ctx, pz)
	s.nextBlock()

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(2_000000)))

	rrz, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(2, len(rrz))
	s.Equal(
		types.RevokeRequest{
			MicroCoins: sdk.NewInt(1_000000),
			HeightToImplementAt: 40_320,
		}, rrz[0],
	)
	s.Equal(
		types.RevokeRequest{
			MicroCoins: sdk.NewInt(2_000000),
			HeightToImplementAt: 20_162,
		}, rrz[1],
	)

	s.EqualValues(3_000000, s.app.GetBankKeeper().GetCoins(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	s.ctx = s.ctx.WithBlockHeight(20_161)
	s.nextBlock()

	s.EqualValues(1_000000, s.app.GetBankKeeper().GetCoins(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	rrz, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(1, len(rrz))
	s.Equal(
		types.RevokeRequest{
			MicroCoins: sdk.NewInt(1_000000),
			HeightToImplementAt: 40_320,
		}, rrz[0],
	)

	s.ctx = s.ctx.WithBlockHeight(40_319)
	s.nextBlock()
	s.EqualValues(0, s.app.GetBankKeeper().GetCoins(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	rrz, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(0, len(rrz))
}

func (s *Suite) TestDelegateDustAmount() {
	s.bk.SetDustDelegation(s.ctx, 1_000000)
	user := app.DefaultGenesisUsers["user1"]

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000000)))
	s.Equal(int64(850000), s.bk.GetCoins(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64())
	resp, err := s.k.GetAccumulation(s.ctx, user)
	s.Error(err)
	s.Zero(resp)
}

func (s *Suite) TestLeaveDust() {
	s.bk.SetDustDelegation(s.ctx, 1_000000)
	user := app.DefaultGenesisUsers["user1"]

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(10_000000)))
	s.nextBlock()
	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(8_000000)))

	s.Equal(int64(500000), s.bk.GetCoins(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64())
	resp, err := s.k.GetAccumulation(s.ctx, user)
	s.Error(err)
	s.Zero(resp)
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
