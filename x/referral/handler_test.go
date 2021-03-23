// +build testing

package referral_test

import (
	"github.com/arterynetwork/artr/util"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
)

func TestReferralHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         referral.Keeper
	accKeeper types.AccountKeeper
	handler   sdk.Handler
}

func (s *HandlerSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetReferralKeeper()
	s.accKeeper = s.app.GetAccountKeeper()
	s.handler = referral.NewHandler(s.k)
}

func (s *HandlerSuite) TearDownTest() {
	s.cleanup()
}

func (s *HandlerSuite) TestTransition() {
	var (
		subj      = app.DefaultGenesisUsers["user4"]
		dest      = app.DefaultGenesisUsers["user3"]
		oldParent = app.DefaultGenesisUsers["user2"]
	)
	var (
		THOUSAND = util.Uartrs(1_000_000000)
		STAKE    = sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(10_000_000000)),
		)
	)

	for i, n := range []sdk.Coins{
		THOUSAND,
		STAKE, STAKE,
		THOUSAND, THOUSAND, THOUSAND, THOUSAND,
		THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND,
	} {
		cz := s.accKeeper.GetAccount(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)]).GetCoins()
		s.Equal(n, cz)
	}

	var (
		msg sdk.Msg
		err error
	)

	msg = types.NewMsgRequestTransition(subj, dest)
	_, err = s.handler(s.ctx, msg)
	s.NoError(err)
	s.nextBlock()

	for i, n := range []sdk.Coins{
		util.Uartrs(1_010_000000), // validator's award
		STAKE, STAKE,
		util.Uartrs(990_000000), THOUSAND, THOUSAND, THOUSAND, // fee
		THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND,
	} {
		cz := s.accKeeper.GetAccount(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)]).GetCoins()
		s.Equal(n, cz)
	}

	msg = types.NewMsgResolveTransition(oldParent, subj, true)
	_, err = s.handler(s.ctx, msg)
	s.NoError(err)

	acc, err := s.k.GetParent(s.ctx, subj)
	s.NoError(err, "get parent")
	s.Equal(dest, acc, "new parent")

	accz, err := s.k.GetChildren(s.ctx, oldParent)
	s.NoError(err, "get old parent's children")
	s.Equal(
		[]sdk.AccAddress{app.DefaultGenesisUsers["user5"]},
		accz, "old parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, dest)
	s.NoError(err, "get new parent's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user6"],
			app.DefaultGenesisUsers["user7"],
			subj,
		},
		accz, "new parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, subj)
	s.NoError(err, "get subject's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user8"],
			app.DefaultGenesisUsers["user9"],
		},
		accz, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj)
	s.NoError(err, "get pending transition")
	s.Nil(acc, "pending transition")

	for i, n := range []int64{
		35_000_000000,
		14_000_000000, 19_990_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)], 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

var bbHeader = abci.RequestBeginBlock{
	Header: abci.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}

func (s *HandlerSuite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}
