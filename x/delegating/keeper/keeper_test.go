// +build testing

package keeper_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authK "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

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
	accKeeper authK.AccountKeeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()

	data, err := ioutil.ReadFile("test-genesis.json")
	if err != nil {
		panic(err)
	}
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(data)

	s.cdc = s.app.Codec()
	s.k = s.app.GetDelegatingKeeper()
	s.bk = s.app.GetBankKeeper()
	s.accKeeper = s.app.GetAccountKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

var TENTH = util.NewFraction(1, 10)

func (s *Suite) TestDelegatingAndRevoking() {
	genesis_time := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user4"]
	validator := app.DefaultGenesisUsers["user3"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Nil(
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1000000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(847450000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(847450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
	s.Equal(
		[]types.RevokeRequest{{
			Time:   genesis_time.Add(14*24*time.Hour),
			Amount: sdk.NewInt(847450000),
		}},
		s.k.GetRevoking(s.ctx, user),
	)

	s.ctx = s.ctx.WithBlockHeight(14*2880 - 1).WithBlockTime(genesis_time.Add((14*2880 - 1) *30*time.Second))
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(847450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
	s.Empty(s.k.GetRevoking(s.ctx, user))
}

func (s *Suite) TestAccrueAfterRevoke() {
	user := app.DefaultGenesisUsers["user4"]
	validator := app.DefaultGenesisUsers["user3"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Nil(
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(350_000000)))
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	t := 0
	for ; t < util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(3_471704)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_010446),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	for ; t < 14*util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(398_603856)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_146244),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	for ; t < 15*util.BlocksOneDay; t++ {
		s.nextBlock()
	}

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(402_075560)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_156690),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
}

func (s *Suite) TestAccrueOnRevoke() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user4"]
	validator := app.DefaultGenesisUsers["user3"]
	s.Equal(
		util.Uartrs(1_000_000000),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Nil(
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
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
	s.Equal(genesisTime, acc.Start)
	s.Equal(genesisTime.Add(24*time.Hour), acc.End)
	s.Equal(int64(2_966075), acc.CurrentUartrs)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(350_000000)))
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2_957177)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(350_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_008898),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	// 2 weeks later
	for ; t < util.BlocksOneDay*29/2; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(401_561033)), // 2.957,177 + 14 * ⌈3.482,150 * 99.7%⌉  + 350
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_155142), // 3.008,898 + 14 * ⌊3.482,150 * 0.3%⌋
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
	acc, err = s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.Equal(genesisTime.Add(29 * 12*time.Hour), acc.Start)
	s.Equal(genesisTime.Add(31 * 12*time.Hour), acc.End)
	s.Equal(int64(0), acc.CurrentUartrs)

	// Half a day later
	for ; t < util.BlocksOneDay*15; t++ {
		s.nextBlock()
	}
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(401_561033)), // The same because accrue time has changed
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(497_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_155142), // The same too
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
}

func (s *Suite) TestAccrue_MissedPart() {
	user := app.DefaultGenesisUsers["user4"]
	validator := app.DefaultGenesisUsers["user3"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Nil(
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.setMissedPart(user, TENTH)

	for t := 0; t < util.BlocksOneDay; t++ { s.nextBlock() }

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(5_322919)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_016016),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
	s.Nil(s.k.Get(s.ctx, user).MissedPart)

	for t := 0; t < util.BlocksOneDay; t++ { s.nextBlock() }

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(11_237273)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_033812),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
}

func (s *Suite) TestAccrueOnRevoke_MissedPart() {
	user := app.DefaultGenesisUsers["user4"]
	validator := app.DefaultGenesisUsers["user3"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Nil(
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_000000),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)

	s.setMissedPart(user, TENTH)

	for t := 0; t < util.BlocksOneDay / 4; t++ { s.nextBlock() }
	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(100_000000)))

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(887153)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(747_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(100_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_002669),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
	s.Nil(s.k.Get(s.ctx, user).MissedPart)

	for t := 0; t < util.BlocksOneDay; t++ { s.nextBlock() }

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(6_103607)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(747_450000)),
			sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(100_000000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		util.Uartrs(3_018365),
		s.bk.GetBalance(s.ctx, validator).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
}

func (s *Suite) TestAccrue_ValidatorBonus() {
	genesisTime := s.ctx.BlockTime()
	validator := app.DefaultGenesisUsers["user3"]

	s.NoError(s.bk.SendCoins(s.ctx, app.DefaultGenesisUsers["user2"], validator, util.Uartrs(1_000_000000)))
	s.nextBlock()

	bonus := util.NewFraction(99, 1000)
	pz := s.k.GetParams(s.ctx)
	pz.ValidatorBonus = bonus
	s.k.SetParams(s.ctx, pz)

	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, validator),
	)

	s.NoError(s.k.Delegate(s.ctx, validator, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, validator),
	)

	s.ctx = s.ctx.WithBlockHeight(1234).WithBlockTime(genesisTime.Add(24*time.Hour))
	s.nextBlock()
	s.nextBlock()

	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(11_728735)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000)),
		),
		s.bk.GetBalance(s.ctx, validator),
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

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(100_000000)))
	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(1_000000)))

	s.Equal(
		[]types.RevokeRequest{
			{
				Amount: sdk.NewInt(1_000000),
				Time: genesisTime.Add(14 * 24 * time.Hour),
			},
		},
		s.k.GetRevoking(s.ctx, user),
	)

	s.nextBlock()
	pz := s.k.GetParams(s.ctx)
	pz.RevokePeriod = 7
	s.k.SetParams(s.ctx, pz)
	s.nextBlock()

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(2_000000)))

	s.Equal(
		[]types.RevokeRequest{
			{
				Amount: sdk.NewInt(1_000000),
				Time: genesisTime.Add(14 * 24 * time.Hour),
			}, {
				Amount: sdk.NewInt(2_000000),
				Time: genesisTime.Add(7 * 24 * time.Hour + time.Minute),
			},
		}, s.k.GetRevoking(s.ctx, user),
	)

	s.EqualValues(3_000000, s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	s.ctx = s.ctx.WithBlockHeight(20_161).WithBlockTime(genesisTime.Add(20_161*30*time.Second))
	s.nextBlock()

	s.EqualValues(1_000000, s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	s.Equal(
		[]types.RevokeRequest{
			{
				Amount: sdk.NewInt(1_000000),
				Time: genesisTime.Add(14 * 24 * time.Hour),
			},
		}, s.k.GetRevoking(s.ctx, user),
	)

	s.ctx = s.ctx.WithBlockHeight(40_319).WithBlockTime(genesisTime.Add(40_319*30*time.Second))
	s.nextBlock()
	s.EqualValues(0, s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigRevokingDenom).Int64())
	s.Empty(s.k.GetRevoking(s.ctx, user))
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
			PercentDaily:  util.NewFraction(21, 30*100).Reduce(),
			TotalUartrs:   5_932150,
			CurrentUartrs: 2_541761,
		},
		*resp,
	)
}

func (s *Suite) TestGetAccumulation_MissedPart() {
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

	s.setMissedPart(user, TENTH.Clone())
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
			PercentDaily:  util.NewFraction(21, 30*100).Reduce(),
			TotalUartrs:   5_338935,
			CurrentUartrs: 1_948546,
			MissedPart:    &TENTH,
		},
		*resp,
	)
}

func (s *Suite) TestGetAccumulation_ValidatorBonus() {
	genesisTime := s.ctx.BlockTime()
	user      := app.DefaultGenesisUsers["user4"]
	validator := app.DefaultGenesisUsers["user3"]

	s.NoError(s.bk.SendCoins(s.ctx, app.DefaultGenesisUsers["user2"], validator, util.Uartrs(1_000_000000)))
	s.nextBlock()

	bonus := util.NewFraction(99, 1000)
	pz := s.k.GetParams(s.ctx)
	pz.ValidatorBonus = bonus
	s.k.SetParams(s.ctx, pz)

	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, validator),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, user),
	)
	s.NoError(s.k.Delegate(s.ctx, validator, sdk.NewInt(1_000_000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000))),
		s.bk.GetBalance(s.ctx, validator),
	)

	s.ctx = s.ctx.WithBlockHeight(1234).WithBlockTime(genesisTime.Add((1234) *30*time.Second))
	s.nextBlock()

	resp, err := s.k.GetAccumulation(s.ctx, user)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(
		types.AccumulationResponse{
			Start:         genesisTime.Add(30*time.Second),
			End:           genesisTime.Add(24*time.Hour +30*time.Second),
			Percent:       21,
			PercentDaily:  util.NewFraction(21, 30*100).Reduce(),
			TotalUartrs:   5_932150,
			CurrentUartrs: 2_541761,
		},
		*resp,
	)

	resp, err = s.k.GetAccumulation(s.ctx, validator)
	s.NoError(err)
	s.NotNil(resp)
	s.Equal(
		types.AccumulationResponse{
			Start:         genesisTime.Add(30*time.Second),
			End:           genesisTime.Add(24*time.Hour +30*time.Second),
			Percent:       30,
			PercentDaily:  util.NewFraction(21, 30*100).Add(bonus.DivInt64(30)).Reduce(),
			TotalUartrs:   8_728735,
			CurrentUartrs: 3_740020,
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

func (s *Suite) TestValidatorBonus() {
	genesisTime := s.ctx.BlockTime()
	validator := app.DefaultGenesisUsers["user3"]
	user := validator

	s.NoError(s.bk.SendCoins(s.ctx, app.DefaultGenesisUsers["user4"], user, util.Uartrs(1_000_000000)))
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000))),
		s.bk.GetBalance(s.ctx, user),
	)

	pz := s.k.GetParams(s.ctx)
	pz.ValidatorBonus = util.Percent(9)
	s.k.SetParams(s.ctx, pz)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1_000_000000)))
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(3_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000)),
		),
		s.bk.GetBalance(s.ctx, user),
	)

	s.ctx = s.ctx.WithBlockHeight(2880).WithBlockTime(genesisTime.Add((2880) *30*time.Second))
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(11_474500)), // 3 + 847.45 * ((21% + 9%) / 30)
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(847_450000)),
		),
		s.bk.GetBalance(s.ctx, user).Add(s.bk.GetBalance(s.ctx, s.accKeeper.GetModuleAddress(auth.FeeCollectorName))...),
	)
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

func (s *Suite) setMissedPart(user sdk.AccAddress, value util.Fraction) {
	store := s.ctx.KVStore(s.app.GetKeys()[delegating.MainStoreKey])
	var data types.Record
	s.cdc.MustUnmarshalBinaryBare(store.Get(user), &data)
	data.MissedPart = &value
	store.Set(user, s.cdc.MustMarshalBinaryBare(&data))
}
