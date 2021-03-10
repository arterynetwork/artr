// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile"
	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/schedule"
)

func TestProfileKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app            *app.ArteryApp
	cleanup        func()
	cdc            *codec.Codec
	ctx            sdk.Context
	k              profile.Keeper
	refKeeper      referral.Keeper
	scheduleKeeper schedule.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetProfileKeeper()
	s.refKeeper = s.app.GetReferralKeeper()
	s.scheduleKeeper = s.app.GetScheduleKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s *Suite) TestCreateAccountWithProfile() {
	_, _, addr := authtypes.KeyTestPubAddr()
	data := types.Profile{
		Nickname:   "v_pupkin",
		CardNumber: 12345,
	}
	s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data)

	s.Equal(addr, s.k.GetProfileAccountByNickname(s.ctx, "v_pupkin"))

	{
		parent, err := s.refKeeper.GetParent(s.ctx, addr)
		s.NoError(err)
		s.Equal(app.DefaultGenesisUsers["user1"], parent)
	}
	{
		h, err := s.refKeeper.GetCompressionBlockHeight(s.ctx, addr)
		s.NoError(err)
		s.Equal(int64(1+2*util.BlocksOneMonth), h)
	}
	{
		tasks := s.scheduleKeeper.GetTasks(s.ctx, 1+2*util.BlocksOneMonth)
		s.Equal(schedule.Schedule{schedule.Task{referral.CompressionHookName, addr}}, tasks)
	}
}

func (s *Suite) TestCreateAccountWithProfile_NonUniqueNick() {
	_, _, addr := authtypes.KeyTestPubAddr()
	data := types.Profile{
		Nickname:   "v_pupkin",
		CardNumber: 12345,
	}
	s.NoError(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))

	_, _, addr = authtypes.KeyTestPubAddr()
	s.Error(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user2"], data))
}

func (s *Suite) TestCreateAccountWithProfile_NickLikeCard() {
	_, _, addr := authtypes.KeyTestPubAddr()
	data := types.Profile{
		Nickname:   "ARTR-1122-3344-5566",
		CardNumber: 12345,
	}
	s.Error(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))

	data.Nickname = "artr-1122-3344-5566"
	s.Error(s.k.CreateAccountWithProfile(s.ctx, addr, app.DefaultGenesisUsers["user1"], data))
}

func (s *Suite) TestCreateAccount() {
	_, _, addr := authtypes.KeyTestPubAddr()
	s.k.CreateAccount(s.ctx, addr, app.DefaultGenesisUsers["user1"])

	s.NotNil(s.k.GetProfile(s.ctx, addr))

	{
		parent, err := s.refKeeper.GetParent(s.ctx, addr)
		s.NoError(err)
		s.Equal(app.DefaultGenesisUsers["user1"], parent)
	}
	{
		h, err := s.refKeeper.GetCompressionBlockHeight(s.ctx, addr)
		s.NoError(err)
		s.Equal(int64(1+2*util.BlocksOneMonth), h)
	}
	{
		tasks := s.scheduleKeeper.GetTasks(s.ctx, 1+2*util.BlocksOneMonth)
		s.Equal(schedule.Schedule{schedule.Task{referral.CompressionHookName, addr}}, tasks)
	}
}
