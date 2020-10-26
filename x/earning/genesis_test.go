// +build testing

package earning_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/earning"
	"github.com/arterynetwork/artr/x/schedule"
)

func TestEarningGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         earning.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetEarningKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestUnlocked() {
	user1 := app.DefaultGenesisUsers["user1"]
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	if err := s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user1, 10, 0),
		earning.NewEarner(user2, 0, 15),
		earning.NewEarner(user3, 20, 30),
	}); err != nil { panic(err) }
	s.checkExportImport()
}

func (s Suite) TestLocked() {
	user1 := app.DefaultGenesisUsers["user1"]
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	if err := s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, user1, 5 * util.GBSize); err != nil { panic(err) }
	if err := s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user1, 10, 0),
		earning.NewEarner(user2, 0, 15),
		earning.NewEarner(user3, 20, 30),
	}); err != nil { panic(err) }
	if err := s.k.Run(
		s.ctx,
		util.NewFraction(7, 30),
		2,
		earning.NewPoints(30, 45),
		10,
	); err != nil { panic(err) }
	s.checkExportImport()
}

func (s Suite) TestSecondPage() {
	user1 := app.DefaultGenesisUsers["user1"]
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	if err := s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, user1, 5 * util.GBSize); err != nil { panic(err) }
	if err := s.k.ListEarners(s.ctx, []earning.Earner{
		earning.NewEarner(user1, 10, 0),
		earning.NewEarner(user2, 0, 15),
		earning.NewEarner(user3, 20, 30),
	}); err != nil { panic(err) }
	if err := s.k.Run(
		s.ctx,
		util.NewFraction(7, 30),
		2,
		earning.NewPoints(30, 45),
		2,
	); err != nil { panic(err) }

	s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(2)
	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{
		Header: abci.Header{
			ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
		},
	})

	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, earning.Params{
		Signers: []sdk.AccAddress{
			app.DefaultGenesisUsers["user9"],
		},
	})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			earning.StoreKey,
			schedule.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			earning.StoreKey:  app.AccAddressDecoder,
			schedule.StoreKey: app.Uint64Decoder,
			params.StoreKey:   app.DummyDecoder,
		},
		map[string]app.Decoder{
			earning.StoreKey:  app.DummyDecoder,
			schedule.StoreKey: app.DummyDecoder,
			params.StoreKey:   app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}