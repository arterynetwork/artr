// +build testing

package noding_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/noding"
	"github.com/arterynetwork/artr/x/noding/types"
)

func TestNodingHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context
	k       noding.Keeper
	handler sdk.Handler
}

func (s *HandlerSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetNodingKeeper()
	s.handler = noding.NewHandler(s.k)
}

func (s *HandlerSuite) TearDownTest() {
	s.cleanup()
}

func (s *HandlerSuite) TestUnjail() {
	user2 := app.DefaultGenesisUsers["user2"]
	msg := types.NewMsgUnjail(user2)

	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	_, pubkey, _ := app.NewTestConsPubAddress()
	if err := s.k.SwitchOn(s.ctx, user2, pubkey); err != nil {
		panic(err)
	}

	validator := abci.Validator{
		Address: pubkey.Address().Bytes(),
		Power:   10,
	}
	votes := []abci.VoteInfo{{Validator: validator, SignedLastBlock: false}}

	// First missed block
	s.nextBlock(proposerKey, votes, nil)

	_, err := s.handler(s.ctx, msg)
	s.Equal(noding.ErrNotJailed, err)

	// Second missed block, jail
	s.nextBlock(proposerKey, votes, nil)

	data, err := s.k.Get(s.ctx, user2)
	s.NoError(err)
	s.True(data.Jailed)

	s.ctx = s.ctx.WithBlockHeight(200)

	_, err = s.handler(s.ctx, msg)
	s.NoError(err)
}

func (s *HandlerSuite) nextBlock(proposer crypto.PubKey, votes []abci.VoteInfo, byzantine []abci.Evidence) (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{
		Header: abci.Header{
			ProposerAddress: proposer.Address().Bytes(),
		},
		LastCommitInfo: abci.LastCommitInfo{
			Votes: votes,
		},
		ByzantineValidators: byzantine,
	})
	return ebr, bbr
}
