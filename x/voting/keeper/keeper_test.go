// +build testing

package keeper_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	votingKeeper "github.com/arterynetwork/artr/x/voting/keeper"
	"github.com/arterynetwork/artr/x/voting/types"
)

func TestVotingKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
	suite.Run(t, new(StatusSuite))
}

type BaseSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context

	k votingKeeper.Keeper

	bbHeader abci.RequestBeginBlock
}

func (s *BaseSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *BaseSuite) setupTest(genesis []byte, consPubKey string) {
	defer func() {
		if err := recover(); err != nil {
			s.FailNow("panic in setup", err)
		}
	}()

	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(genesis)
	s.k = s.app.GetVotingKeeper()

	s.bbHeader = abci.RequestBeginBlock{
		Header: tmproto.Header{
			ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, consPubKey).Address().Bytes(),
		},
	}
}

func (s *BaseSuite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
	bbr := s.app.BeginBlocker(s.ctx, s.bbHeader)
	return ebr, bbr
}


type Suite struct {
	BaseSuite
}

func (s *Suite) SetupTest() { s.setupTest(nil, app.DefaultUser1ConsPubKey) }

func (s *Suite) TestInitialState() {
	_, ok := s.k.GetCurrentPoll(s.ctx)
	s.False(ok)
	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(0, y)
	s.EqualValues(0, n)
}

func (s *Suite) TestStartPoll_Validators() {
	orig := types.NewPollValidators(
		app.DefaultGenesisUsers["user1"],
		"the question",
		"To be or not to be?",
		util.NewFraction(1, 2),
	)
	s.NoError(s.k.StartPoll(s.ctx, orig))

	got, ok := s.k.GetCurrentPoll(s.ctx)
	s.True(ok)
	s.Equal(orig.Author, got.Author)
	s.Equal(orig.Name, got.Name)
	s.Equal(orig.Question, got.Question)
	s.NotNil(got.Quorum)
	s.Equal(*orig.Quorum, *got.Quorum)
	s.NotNil(got.StartTime)
	s.Equal(s.ctx.BlockTime(), *got.StartTime)
	s.NotNil(got.EndTime)
	s.Equal(s.ctx.BlockTime().Add(18 * time.Hour), *got.EndTime)

	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(0, y)
	s.EqualValues(0, n)
}

func (s *Suite) TestVotePoll_Validators() {
	poll := types.NewPollValidators(
		app.DefaultGenesisUsers["user1"],
		"the question",
		"To be or not to be?",
		util.NewFraction(2, 3),
	)
	s.NoError(s.k.StartPoll(s.ctx, poll))

	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user1"].String(), true))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user2"].String(), false))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user3"].String(), true))
	s.Error(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user15"].String(), true))

	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(2, y)
	s.EqualValues(1, n)
}

func (s *Suite) TestEndPoll_Positive() {
	genesisTime := s.ctx.BlockTime()
	poll := types.NewPollValidators(
		app.DefaultGenesisUsers["user1"],
		"the question",
		"To be or not to be?",
		util.NewFraction(2, 3),
	)
	s.NoError(s.k.StartPoll(s.ctx, poll))

	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user1"].String(), true))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user2"].String(), false))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user3"].String(), true))

	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(18*time.Hour)).WithBlockHeight(s.ctx.BlockHeight()+1)
	s.nextBlock()

	_, ok := s.k.GetCurrentPoll(s.ctx)
	s.False(ok)
	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(0, y)
	s.EqualValues(0, n)

	history := s.k.GetPollHistoryAll(s.ctx)
	s.Equal(1, len(history))
	s.Equal(poll.Author, history[0].Poll.Author)
	s.Equal(poll.Name, history[0].Poll.Name)
	s.Equal(poll.Question, history[0].Poll.Question)
	s.NotNil(history[0].Poll.Quorum)
	s.Equal(*poll.Quorum, *history[0].Poll.Quorum)
	s.NotNil(history[0].Poll.StartTime)
	s.Equal(genesisTime, *history[0].Poll.StartTime)
	s.NotNil(history[0].Poll.EndTime)
	s.Equal(genesisTime.Add(18*time.Hour), *history[0].Poll.EndTime)
	s.EqualValues(2, history[0].Yes)
	s.EqualValues(1, history[0].No)
	s.EqualValues(types.DECISION_POSITIVE, history[0].Decision)
}

func (s *Suite) TestEndPoll_Negative() {
	genesisTime := s.ctx.BlockTime()
	poll := types.NewPollValidators(
		app.DefaultGenesisUsers["user1"],
		"the question",
		"To be or not to be?",
		util.NewFraction(2, 3),
	)
	s.NoError(s.k.StartPoll(s.ctx, poll))

	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user1"].String(), true))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user2"].String(), false))

	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(18*time.Hour)).WithBlockHeight(s.ctx.BlockHeight()+1)
	s.nextBlock()

	_, ok := s.k.GetCurrentPoll(s.ctx)
	s.False(ok)
	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(0, y)
	s.EqualValues(0, n)

	history := s.k.GetPollHistoryAll(s.ctx)
	s.Equal(1, len(history))
	s.Equal(poll.Author, history[0].Poll.Author)
	s.Equal(poll.Name, history[0].Poll.Name)
	s.Equal(poll.Question, history[0].Poll.Question)
	s.NotNil(history[0].Poll.Quorum)
	s.Equal(*poll.Quorum, *history[0].Poll.Quorum)
	s.NotNil(history[0].Poll.StartTime)
	s.Equal(genesisTime, *history[0].Poll.StartTime)
	s.NotNil(history[0].Poll.EndTime)
	s.Equal(genesisTime.Add(18*time.Hour), *history[0].Poll.EndTime)
	s.EqualValues(1, history[0].Yes)
	s.EqualValues(1, history[0].No)
	s.EqualValues(types.DECISION_NEGATIVE, history[0].Decision)
}

func (s *Suite) TestEndPoll_Undecided() {
	genesisTime := s.ctx.BlockTime()
	poll := types.Poll{
		Author:   app.DefaultGenesisUsers["user1"].String(),
		Name:     "the question",
		Question: "To be or not to be?",
	}
	s.NoError(s.k.StartPoll(s.ctx, poll))

	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user1"].String(), true))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user2"].String(), false))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user3"].String(), true))

	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(18*time.Hour)).WithBlockHeight(s.ctx.BlockHeight()+1)
	s.nextBlock()

	_, ok := s.k.GetCurrentPoll(s.ctx)
	s.False(ok)
	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(0, y)
	s.EqualValues(0, n)

	history := s.k.GetPollHistoryAll(s.ctx)
	s.Equal(1, len(history))
	s.Equal(poll.Author, history[0].Poll.Author)
	s.Equal(poll.Name, history[0].Poll.Name)
	s.Equal(poll.Question, history[0].Poll.Question)
	s.Nil(history[0].Poll.Quorum)
	s.NotNil(history[0].Poll.StartTime)
	s.Equal(genesisTime, *history[0].Poll.StartTime)
	s.NotNil(history[0].Poll.EndTime)
	s.Equal(genesisTime.Add(18*time.Hour), *history[0].Poll.EndTime)
	s.EqualValues(2, history[0].Yes)
	s.EqualValues(1, history[0].No)
	s.EqualValues(types.DECISION_UNSPECIFIED, history[0].Decision)
}


type StatusSuite struct {
	BaseSuite

	governor sdk.AccAddress
}

func (s *StatusSuite) SetupTest() {
	data, err := ioutil.ReadFile("../../referral/keeper/test-genesis-status-3x3.json")
	if err != nil {
		panic(err)
	}
	s.setupTest(data, "artrvalconspub1zcjduepqpme87trszw7awc62ra2de9edwr40v7xy7yfhvpvds96fncagm04qxu308e")

	s.governor, err = sdk.AccAddressFromBech32("artr1cd4g3grtpslw799alf78w9gc2vqdnhrldc0tjc")
	if err != nil { panic(err) }
}

func (s *StatusSuite) TestStartPoll_Status() {
	orig := types.NewPollStatus(
		s.governor,
		"the question",
		"To be or not to be?",
		util.NewFraction(1, 2),
		referral.StatusChampion,
	)
	s.NoError(s.k.StartPoll(s.ctx, orig))

	got, ok := s.k.GetCurrentPoll(s.ctx)
	s.True(ok)
	s.Equal(orig.Author, got.Author)
	s.Equal(orig.Name, got.Name)
	s.Equal(orig.Question, got.Question)
	s.NotNil(got.Quorum)
	s.Equal(*orig.Quorum, *got.Quorum)
	s.NotNil(got.StartTime)
	s.Equal(s.ctx.BlockTime(), *got.StartTime)
	s.NotNil(got.EndTime)
	s.Equal(s.ctx.BlockTime().Add(18 * time.Hour), *got.EndTime)

	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(0, y)
	s.EqualValues(0, n)
}

func (s *StatusSuite) TestVotePoll_Status() {
	const (
		root  = "artr1yhy6d3m4utltdml7w7zte7mqx5wyuskq9rr5vg"
		user1 = "artr1d4ezqdj03uachct8hum0z9zlfftzdq2f6yzvhj"
		user2 = "artr1h8s8yf433ypjc5htavsyc9zvg3vk43vms03z3l"
	)
	poll := types.NewPollStatus(
		s.governor,
		"the question",
		"To be or not to be?",
		util.NewFraction(2, 3),
		referral.StatusChampion,
	)
	s.NoError(s.k.StartPoll(s.ctx, poll))

	s.NoError(s.k.Answer(s.ctx, root, true))
	s.Error(s.k.Answer(s.ctx, user1, false))
	s.Error(s.k.Answer(s.ctx, user2, true))

	y, n := s.k.GetPollStatus(s.ctx)
	s.EqualValues(1, y)
	s.EqualValues(0, n)
}
