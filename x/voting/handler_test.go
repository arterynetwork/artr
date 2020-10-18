// +build testing

package voting_test

import (
	"github.com/arterynetwork/artr/x/voting/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/voting"
)

func TestVotingHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         voting.Keeper
	handler   sdk.Handler
}

func (s *HandlerSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx     = s.app.NewContext(true, abci.Header{Height: 1})
	s.k       = s.app.GetVotingKeeper()
	s.handler = voting.NewHandler(s.k)
}

func (s *HandlerSuite) TearDownTest() {
	s.cleanup()
}

func (s *HandlerSuite) TestAddGovernorProposal() {
	msg := types.NewMsgCreateProposal(
		app.DefaultGenesisUsers["user1"],
		"Heeere's Johnny!",
		types.ProposalTypeGovernmentAdd,
		types.AddressProposalParams{Address: app.DefaultGenesisUsers["user13"]},
	)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.voteFor()

	s.Equal(
		types.Government{
			app.DefaultGenesisUsers["user1"],
			app.DefaultGenesisUsers["user2"],
			app.DefaultGenesisUsers["user3"],
			app.DefaultGenesisUsers["user13"],
		},
		s.k.GetGovernment(s.ctx),
	)
}

func (s *HandlerSuite) TestRemoveGovernorProposal() {
	msg := types.NewMsgCreateProposal(
		app.DefaultGenesisUsers["user1"],
		"We need nobody but us",
		types.ProposalTypeGovernmentRemove,
		types.AddressProposalParams{Address: app.DefaultGenesisUsers["user3"]},
	)
	var err error
	_, err = s.handler(s.ctx, msg)
	s.NoError(err)

	_, err = s.handler(s.ctx, types.NewMsgProposalVote(app.DefaultGenesisUsers["user2"], true))
	s.NoError(err)

	_, err = s.handler(s.ctx, types.NewMsgProposalVote(app.DefaultGenesisUsers["user3"], false))
	s.NoError(err)

	s.Equal(
		types.Government{
			app.DefaultGenesisUsers["user1"],
			app.DefaultGenesisUsers["user2"],
		},
		s.k.GetGovernment(s.ctx),
	)
}

func (s *HandlerSuite) TestSoftwareUpgrade() {
	msg := types.NewMsgCreateProposal(
		app.DefaultGenesisUsers["user1"],
		"Jury rig",
		types.ProposalTypeSoftwareUpgrade,
		types.SoftwareUpgradeProposalParams{
			Name:   "v.2.0.1",
			Height: 5,
			Info:   "https://example.com/binary/v.2.0.1/info.json?checksum=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)
	s.voteFor()

	plan, ok := s.app.GetUpgradeKeeper().GetUpgradePlan(s.ctx)
	s.True(ok)
	s.Equal(
		upgrade.Plan{
			Name:   "v.2.0.1",
			Time:   time.Time{},
			Height: 5,
			Info:   "https://example.com/binary/v.2.0.1/info.json?checksum=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		plan,
	)
}

func (s *HandlerSuite) TestCancelSoftwareUpgrade() {
	s.NoError(s.app.GetUpgradeKeeper().ScheduleUpgrade(s.ctx, upgrade.Plan{
		Name:   "v.2.0.1",
		Time:   time.Time{},
		Height: 5,
		Info:   "https://example.com/binary/v.2.0.1/info.json?checksum=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	}))
	msg := types.NewMsgCreateProposal(
		app.DefaultGenesisUsers["user1"],
		"Oops!",
		types.ProposalTypeCancelSoftwareUpgrade,
		types.EmptyProposalParams{},
	)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)
	s.voteFor()

	_, ok := s.app.GetUpgradeKeeper().GetUpgradePlan(s.ctx)
	s.False(ok)
}

func (s *HandlerSuite) voteFor() {
	msg := types.NewMsgProposalVote(
		app.DefaultGenesisUsers["user2"],
		true,
	)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	msg.Voter = app.DefaultGenesisUsers["user3"]
	_, err = s.handler(s.ctx, msg)
	s.NoError(err)
}
