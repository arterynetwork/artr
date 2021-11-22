// +build testing

package voting_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	dt "github.com/arterynetwork/artr/x/delegating/types"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/voting/keeper"
	"github.com/arterynetwork/artr/x/voting/types"
)

func TestVotingGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context
	k       keeper.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)
	s.k = s.app.GetVotingKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestCurrentProposal() {
	s.k.SetCurrentProposal(s.ctx, types.Proposal{
		Name: "halving",
		Type: types.PROPOSAL_TYPE_DELEGATION_AWARD,
		Args: &types.Proposal_DelegationAward{
			DelegationAward: &types.DelegationAwardArgs{
				Award: dt.Percentage{
					Minimal:      11,
					ThousandPlus: 12,
					TenKPlus:     14,
					HundredKPlus: 15,
				},
			},
		},
		Author:   app.DefaultGenesisUsers["user1"].String(),
		EndBlock: 42,
	})
	s.k.SetStartBlock(s.ctx)
	s.k.SetAgreed(s.ctx, types.Government{Members: []string{app.DefaultGenesisUsers["user2"].String()}})
	s.k.SetAgreed(s.ctx, types.Government{Members: []string{app.DefaultGenesisUsers["user3"].String()}})
	s.k.SetDisagreed(s.ctx, types.Government{Members: []string{app.DefaultGenesisUsers["user4"].String()}})
	s.k.SetDisagreed(s.ctx, types.Government{Members: []string{app.DefaultGenesisUsers["user5"].String()}})
	s.checkExportImport()
}

func (s Suite) TestHistory() {
	proposal := types.Proposal{
		Name: "halving",
		Type: types.PROPOSAL_TYPE_DELEGATION_AWARD,
		Args: &types.Proposal_DelegationAward{
			DelegationAward: &types.DelegationAwardArgs{
				Award: dt.Percentage{
					Minimal:      11,
					ThousandPlus: 12,
					TenKPlus:     14,
					HundredKPlus: 15,
				},
			},
		},
		Author:   app.DefaultGenesisUsers["user1"].String(),
		EndTime:  &time.Time{},
	}
	*proposal.EndTime = time.Date(2021, 8, 3, 11, 20, 10, 666128000, time.UTC)

	s.k.SetCurrentProposal(s.ctx, proposal)
	s.k.SetStartBlock(s.ctx)
	s.k.EndProposal(s.ctx, proposal, true)
	s.Equal(1, len(s.k.GetHistory(s.ctx, 100, 1)))
	s.checkExportImport()
}

func (s *Suite) TestGovernment() {
	s.k.AddGovernor(s.ctx, app.DefaultGenesisUsers["user13"])
	s.k.RemoveGovernor(s.ctx, s.k.GetGovernment(s.ctx).GetMember(0))
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, types.NewParams(33, 42))
	s.checkExportImport()
}

func (s *Suite) TestActivePoll_FullData() {
	zero := util.FractionZero()
	s.NoError(s.k.StartPoll(s.ctx, types.Poll{
		Author:   app.DefaultGenesisUsers["user1"].String(),
		Name:     "Hamlet's dilemma",
		Question: "To be or not to be? It's the question.",
		Quorum:   &zero,
		Requirements: &types.Poll_CanValidate{CanValidate: &types.Poll_Unit{}},
	}))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user1"].String(), true))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user2"].String(), false))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user3"].String(), true))
	s.checkExportImport()
}

func (s *Suite) TestActivePoll_Minimal() {
	s.NoError(s.k.StartPoll(s.ctx, types.Poll{
		Author:   app.DefaultGenesisUsers["user1"].String(),
		Name:     "Hamlet's dilemma",
		Requirements: &types.Poll_CanValidate{CanValidate: &types.Poll_Unit{}},
	}))
	s.checkExportImport()
}

func (s *Suite) TestActivePoll_MinStatus() {
	s.NoError(s.k.StartPoll(s.ctx, types.Poll{
		Author:   app.DefaultGenesisUsers["user1"].String(),
		Name:     "Hamlet's dilemma",
		Requirements: &types.Poll_MinStatus{MinStatus: referral.StatusChampion},
	}))
	s.checkExportImport()
}

func (s *Suite) TestPollHistory() {
	zero := util.FractionZero()
	s.NoError(s.k.StartPoll(s.ctx, types.Poll{
		Author:   app.DefaultGenesisUsers["user1"].String(),
		Name:     "Hamlet's dilemma",
		Question: "To be or not to be? It's the question.",
		Quorum:   &zero,
		Requirements: &types.Poll_CanValidate{CanValidate: &types.Poll_Unit{}},
	}))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user1"].String(), true))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user2"].String(), false))
	s.NoError(s.k.Answer(s.ctx, app.DefaultGenesisUsers["user3"].String(), true))

	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(19*time.Hour)).WithBlockHeight(s.ctx.BlockHeight()+1)
	s.nextBlock()

	_, ok := s.k.GetCurrentPoll(s.ctx)
	s.False(ok)

	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		s.ctx.BlockTime(),
		[]string{
			types.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			types.StoreKey: func(bz []byte) (string, error) {
				if (len(bz) == len(types.KeyHistoryPrefix)+8) && bytes.Equal(types.KeyHistoryPrefix, bz[:len(types.KeyHistoryPrefix)]) {
					return fmt.Sprintf("%s %d", string(types.KeyHistoryPrefix), binary.BigEndian.Uint64(bz[len(types.KeyHistoryPrefix):])), nil
				}
				if utf8.Valid(bz) {
					return string(bz), nil
				}
				return "", fmt.Errorf("invalid format")
			},
			params.StoreKey: app.DummyDecoder,
		},
		map[string]app.Decoder{
			types.StoreKey:  app.DummyDecoder,
			params.StoreKey: app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}

func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
	bbr := s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{
		Header: tmproto.Header{
			ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
		},
	})
	return ebr, bbr
}
