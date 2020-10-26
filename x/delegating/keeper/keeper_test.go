// +build testing

package keeper_test

import (
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/delegating"
)

func TestDelegatingKeeper(t *testing.T) { suite.Run(t, new(Suite)) }

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()

	cdc       *codec.Codec
	ctx       sdk.Context
	k         delegating.Keeper
	accKeeper auth.AccountKeeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)

	s.cdc       = s.app.Codec()
	s.ctx       = s.app.NewContext(true, abci.Header{})
	s.k         = s.app.GetDelegatingKeeper()
	s.accKeeper = s.app.GetAccountKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s *Suite) TestDelegatingAndRevoking() {
	user := app.DefaultGenesisUsers["user1"]
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1000000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(1000000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(850000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)

	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(850000000)))
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, sdk.NewInt(850000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
	revoking, err := s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Equal(
		[]types.RevokeRequest{{
			HeightToImplementAt: 14 * 2880,
			MicroCoins:          sdk.NewInt(850000000),
		}},
		revoking,
	)

	s.ctx = s.ctx.WithBlockHeight(14 * 2880 - 1)
	s.nextBlock()
	s.Equal(
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(850000000))),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins(),
	)
	revoking, err = s.k.GetRevoking(s.ctx, user)
	s.NoError(err)
	s.Empty(revoking)
}


var bbHeader = abci.RequestBeginBlock{
	Header: abci.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}
func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}