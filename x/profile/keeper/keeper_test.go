// +build testing

package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/arterynetwork/artr/x/referral"
	scheduleK "github.com/arterynetwork/artr/x/schedule/keeper"
	schedule "github.com/arterynetwork/artr/x/schedule/types"
)

func TestProfileKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	cdc     codec.BinaryMarshaler
	ctx     sdk.Context
	k       keeper.Keeper
	bk      bank.Keeper
	rk      referral.Keeper
	sk      scheduleK.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.k = s.app.GetProfileKeeper()
	s.bk = s.app.GetBankKeeper()
	s.rk = s.app.GetReferralKeeper()
	s.sk = s.app.GetScheduleKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *Suite) TestCreateAccountWithProfile() {
	genesisTime := s.ctx.BlockTime()
	_, _, addr := testdata.KeyTestPubAddr()
	data := types.Profile{
		Nickname:   "v_pupkin",
		CardNumber: 12345,
	}
	s.NoError(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))

	s.Equal(addr, s.k.GetProfileAccountByNickname(s.ctx, "v_pupkin"))

	{
		parent, err := s.rk.GetParent(s.ctx, addr.String())
		s.NoError(err)
		s.Equal(app.DefaultGenesisUsers["user1"].String(), parent)
	}
	{
		refInfo, err := s.rk.Get(s.ctx, addr.String())
		s.NoError(err)
		t := refInfo.CompressionAt
		s.NotNil(t)
		s.Equal(genesisTime.Add(2*30*24*time.Hour), *t)
	}
	{
		t := genesisTime.Add(2 * 30 * 24 * time.Hour)
		tasks := s.sk.GetTasks(s.ctx, t, t.Add(1))
		s.Equal(
			[]schedule.Task{
				{Time: t, HandlerName: referral.CompressionHookName, Data: []byte(addr.String())},
			},
			tasks,
		)
	}
}

func (s *Suite) TestCreateAccountWithProfile_NonUniqueNick() {
	_, _, addr := testdata.KeyTestPubAddr()
	data := types.Profile{
		Nickname:   "v_pupkin",
		CardNumber: 12345,
	}
	s.NoError(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))

	_, _, addr = testdata.KeyTestPubAddr()
	s.Error(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user2"], data))
}

func (s *Suite) TestCreateAccountWithProfile_NickLikeCard() {
	_, _, addr := testdata.KeyTestPubAddr()
	data := types.Profile{
		Nickname:   "ARTR-1122-3344-5566",
		CardNumber: 12345,
	}
	s.Error(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))

	data.Nickname = "artr-1122-3344-5566"
	s.Error(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))
}

func (s *Suite) TestCreateAccount() {
	genesisTime := s.ctx.BlockTime()
	_, _, addr := testdata.KeyTestPubAddr()
	s.NoError(s.k.CreateAccount(s.ctx, addr, app.DefaultGenesisUsers["user1"]))

	s.NotNil(s.k.GetProfile(s.ctx, addr), "profile created")

	{
		parent, err := s.rk.GetParent(s.ctx, addr.String())
		s.NoError(err)
		s.Equal(app.DefaultGenesisUsers["user1"].String(), parent)
	}
	{
		refInfo, err := s.rk.Get(s.ctx, addr.String())
		s.NoError(err)
		t := refInfo.CompressionAt
		s.NotNil(t)
		s.Equal(genesisTime.Add(2*30*24*time.Hour), *t)
	}
	{
		t := genesisTime.Add(2 * 30 * 24 * time.Hour)
		tasks := s.sk.GetTasks(s.ctx, t, t.Add(1))
		s.Equal(
			[]schedule.Task{
				{Time: t, HandlerName: referral.CompressionHookName, Data: []byte(addr.String())},
			},
			tasks,
		)
	}
}

func (s *Suite) TestRename() {
	user := app.DefaultGenesisUsers["user2"]
	p := s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("user2", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // from genesis
		s.bk.GetBalance(s.ctx, user),
	)

	p.Nickname = "user2a"
	s.NoError(s.k.SetProfile(s.ctx, user, *p))
	p = s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("user2a", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(999_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // -1 ARTR for rename
		s.bk.GetBalance(s.ctx, user),
	)
}

func (s *Suite) TestRename_InsufficientFunds() {
	user := app.DefaultGenesisUsers["user2"]
	s.NoError(s.app.GetBankKeeper().SendCoins(s.ctx, user, app.DefaultGenesisUsers["user3"], util.Uartrs(999_000001)))
	p := s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("user2", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(999999)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // from genesis
		s.bk.GetBalance(s.ctx, user),
	)

	p.Nickname = "user2a"
	s.Error(s.k.SetProfile(s.ctx, user, *p))
	p = s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("user2", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(999999)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // nothing changed
		s.bk.GetBalance(s.ctx, user),
	)
}

func (s *Suite) TestRename_ClearAndSet() {
	user := app.DefaultGenesisUsers["user2"]
	p := s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("user2", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // from genesis
		s.bk.GetBalance(s.ctx, user),
	)

	p.Nickname = ""
	s.NoError(s.k.SetProfile(s.ctx, user, *p))
	p = s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // nothing changed, removal is free
		s.bk.GetBalance(s.ctx, user),
	)

	p.Nickname = "user2"
	s.NoError(s.k.SetProfile(s.ctx, user, *p))
	p = s.k.GetProfile(s.ctx, user)

	s.NotNil(p)
	s.Equal("user2", p.Nickname)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(999_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(20_000_000000)),
		), // -1 ARTR
		s.bk.GetBalance(s.ctx, user),
	)
}
