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

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/delegating/types"
)

func TestDelegatingKeeper(t *testing.T) { suite.Run(t, new(Suite)) }

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc codec.BinaryMarshaler
	ctx sdk.Context
	k   delegating.Keeper
	bk  bank.Keeper
	//accKeeper authK.AccountKeeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.k = s.app.GetDelegatingKeeper()
	s.bk = s.app.GetBankKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *Suite) TestDelegatingAndRevoking() {
	genesis_time := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user4"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1000000000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1000000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847450000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(847450000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(847450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	revoking, err := s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(
		[]types.RevokeRequest{{
			Time:   genesis_time.Add(14*24*time.Hour),
			Amount: sdk.NewInt(847450000),
		}},
		revoking,
	)

	s.ctx = s.ctx.WithBlockHeight(14*2880 - 1).WithBlockTime(genesis_time.Add((14*2880 - 1) *30*time.Second))
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(847450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	revoking, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Empty(revoking)
}

func (s *Suite) TestAccrueAfterRevoke() {
	user := app.DefaultGenesisUsers["user4"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(350_000000)))
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)

	t := 0
	for ; t < util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(3_482150)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)

	for ; t < 14*util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(398_750100)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)

	for ; t < 15*util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(402_232250)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
}

func (s *Suite) TestAccrueOnRevoke() {
	genesis_time := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user4"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)

	t := 0

	for ; t < util.BlocksOneDay/2; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	acc, err := s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.Equal(genesis_time, acc.Start)
	s.Equal(genesis_time.Add(24*time.Hour), acc.End)
	s.Equal(int64(2_966075), acc.CurrentUartrs)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(350_000000)))
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2_966075)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)

	// 2 weeks later
	for ; t < util.BlocksOneDay*29/2; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(401_716175)), // 2.966075 + 14 * 3.482150 + 350
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	acc, err = s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.Equal(genesis_time.Add(29 * 12*time.Hour), acc.Start)
	s.Equal(genesis_time.Add(31 * 12*time.Hour), acc.End)
	s.Equal(int64(0), acc.CurrentUartrs)

	// Half a day later
	for ; t < util.BlocksOneDay*15; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(401_716175)), // The same because accrue time has changed
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
}

func (s *Suite) TestMinDelegation() {
	user := app.DefaultGenesisUsers["user4"]
	s.ErrorIs(s.k.Delegate(s.ctx, user, sdk.NewInt(999)), types.ErrLessThanMinimum)
	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1000)))

	p := s.k.GetParams(s.ctx)
	p.MinDelegate = 2000
	s.k.SetParams(s.ctx, p)
	s.ErrorIs(s.k.Delegate(s.ctx, user, sdk.NewInt(1999)), types.ErrLessThanMinimum)
	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(2000)))
}

func (s *Suite) TestDelegateDustAmount() {
	p := s.bk.GetParams(s.ctx)
	p.DustDelegation = 1_000000
	s.bk.SetParams(s.ctx, p)
	user := app.DefaultGenesisUsers["user4"]

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000000)))
	s.Equal(int64(847450), s.bk.GetBalance(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64())
	resp, err := s.k.GetAccumulation(s.ctx, user)
	s.Equal(types.ErrNothingDelegated, err)
	s.Nil(resp)
}

func (s *Suite) TestLeaveDust() {
	p := s.bk.GetParams(s.ctx)
	p.DustDelegation = 1_000000
	s.bk.SetParams(s.ctx, p)
	user := app.DefaultGenesisUsers["user4"]

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(10_000000)))
	s.nextBlock()
	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(8_000000)))

	s.Equal(int64(474500), s.bk.GetBalance(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64())
	resp, err := s.k.GetAccumulation(s.ctx, user)
	s.Equal(types.ErrNothingDelegated, err)
	s.Nil(resp)
}

func (s *Suite) TestRevokePeriod() {
	user := app.DefaultGenesisUsers["user2"]
	genesisTime := s.ctx.BlockTime()

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(1_000000)))

	rrz, err := s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(1, len(rrz))
	s.Equal(
		types.RevokeRequest{
			Amount: sdk.NewInt(1_000000),
			Time: genesisTime.Add(14 * 24 * time.Hour),
		}, rrz[0],
	)

	s.nextBlock()
	pz := s.k.GetParams(s.ctx)
	pz.RevokePeriod = 7
	s.k.SetParams(s.ctx, pz)
	s.nextBlock()

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(2_000000)))

	rrz, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(2, len(rrz))
	s.Equal(
		types.RevokeRequest{
			Amount: sdk.NewInt(1_000000),
			Time: genesisTime.Add(14 * 24 * time.Hour),
		}, rrz[0],
	)
	s.Equal(
		types.RevokeRequest{
			Amount: sdk.NewInt(2_000000),
			Time: genesisTime.Add(7 * 24 * time.Hour + time.Minute),
		}, rrz[1],
	)

	s.EqualValues(3_000000, s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	s.ctx = s.ctx.WithBlockHeight(20_161).WithBlockTime(genesisTime.Add(20_161*30*time.Second))
	s.nextBlock()

	s.EqualValues(1_000000, s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	rrz, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(1, len(rrz))
	s.Equal(
		types.RevokeRequest{
			Amount: sdk.NewInt(1_000000),
			Time: genesisTime.Add(14 * 24 * time.Hour),
		}, rrz[0],
	)

	s.ctx = s.ctx.WithBlockHeight(40_319).WithBlockTime(genesisTime.Add(40_319*30*time.Second))
	s.nextBlock()
	s.EqualValues(0, s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	rrz, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(0, len(rrz))
}

func (s *Suite) TestGetAccumulation() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user4"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)

	s.ctx = s.ctx.WithBlockHeight(1234 - 1).WithBlockTime(genesisTime.Add((1234 - 1) *30*time.Second))
	s.nextBlock()

	resp, err := s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(
		types.AccumulationResponse{
			Start:         genesisTime,
			End:           genesisTime.Add(24*time.Hour),
			Percent:       21,
			TotalUartrs:   5_932150,
			CurrentUartrs: 2_541761,
		},
		*resp,
	)
}

func (s *Suite) TestDelegateAfterBanishment() {
	rk := s.app.GetReferralKeeper()
	user := app.DefaultGenesisUsers["user4"]

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 8640).WithBlockTime(s.ctx.BlockTime().Add(4*24*time.Hour))
	s.nextBlock()
	r, err := rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.False(r.Active)
	s.NotNil(r.CompressionAt)

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 172800).WithBlockTime(s.ctx.BlockTime().Add(2*30*24*time.Hour))
	s.nextBlock()
	r, err = rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.NotNil(r.BanishmentAt)

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 86400).WithBlockTime(s.ctx.BlockTime().Add(30*24*time.Hour))
	s.nextBlock()
	r, err = rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.True(r.Banished)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(10_000000)))
	r,err = rk.Get(s.ctx, user.String())
	s.NoError(err)
	s.False(r.Banished)
}

var bbHeader = abci.RequestBeginBlock{
	Header: tmproto.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}

func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30*time.Second))
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}
