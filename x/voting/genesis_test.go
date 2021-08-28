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

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	dt "github.com/arterynetwork/artr/x/delegating/types"
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

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			types.StoreKey,
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
		},
		map[string]app.Decoder{
			types.StoreKey: app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}
