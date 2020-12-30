// +build testing

package voting_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/voting"
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
	k       voting.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetVotingKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestCurrentProposal() {
	s.k.SetCurrentProposal(s.ctx, types.Proposal{
		Name:     "halving",
		TypeCode: types.ProposalTypeDelegationAward,
		Params: types.DelegationAwardProposalParams{
			Minimal:      11,
			ThousandPlus: 12,
			TenKPlus:     14,
			HundredKPlus: 15,
		},
		Author:   app.DefaultGenesisUsers["user1"],
		EndBlock: 42,
	})
	s.k.SetStartBlock(s.ctx)
	s.k.SetAgreed(s.ctx, []sdk.AccAddress{app.DefaultGenesisUsers["user2"]})
	s.k.SetAgreed(s.ctx, []sdk.AccAddress{app.DefaultGenesisUsers["user3"]})
	s.k.SetDisagreed(s.ctx, []sdk.AccAddress{app.DefaultGenesisUsers["user4"]})
	s.k.SetDisagreed(s.ctx, []sdk.AccAddress{app.DefaultGenesisUsers["user5"]})
	s.checkExportImport()
}

func (s Suite) TestHistory() {
	proposal := types.Proposal{
		Name:     "halving",
		TypeCode: types.ProposalTypeDelegationAward,
		Params: types.DelegationAwardProposalParams{
			Minimal:      11,
			ThousandPlus: 12,
			TenKPlus:     14,
			HundredKPlus: 15,
		},
		Author:   app.DefaultGenesisUsers["user1"],
		EndBlock: 42,
	}
	s.k.SetCurrentProposal(s.ctx, proposal)
	s.k.SetStartBlock(s.ctx)
	s.k.EndProposal(s.ctx, proposal, true)
	s.Equal(1, len(s.k.GetHistory(s.ctx, 100, 1)))
	s.checkExportImport()
}

func (s *Suite) TestGovernment() {
	s.k.AddGovernor(s.ctx, app.DefaultGenesisUsers["user13"])
	s.k.RemoveGovernor(s.ctx, s.k.GetGovernment(s.ctx)[0])
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			voting.StoreKey,
		},
		map[string]app.Decoder{
			voting.StoreKey: func(bz []byte) (string, error) {
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
			voting.StoreKey: app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}
