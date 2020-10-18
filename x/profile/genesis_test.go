// +build testing

package profile_test

import (
	"github.com/cosmos/cosmos-sdk/x/params"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/profile"
	"github.com/arterynetwork/artr/x/profile/types"
)

func TestProfileGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         profile.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetProfileKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestFullData() {
	_, _, newAcc := authtypes.KeyTestPubAddr()
	s.k.CreateAccountWithProfile(s.ctx, newAcc, app.DefaultGenesisUsers["user13"], types.Profile{
		AutoRedeligate: true,
		AutoPay:        true,
		ActiveUntil:    42,
		Noding:         true,
		Storage:        true,
		Validator:      true,
		VPN:            true,
		Nickname:       "FooBar",
		CardNumber:     12345,
	})
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.Panics(func(){
		s.k.SetParams(s.ctx, profile.Params{
			Creators:  []sdk.AccAddress{app.DefaultGenesisUsers["user5"]},
			Fee:       1234,
			CardMagic: 999999666666,
		})
	})
	s.k.SetParams(s.ctx, profile.Params{
		Creators:  []sdk.AccAddress{app.DefaultGenesisUsers["user5"]},
		Fee:       1234,
	})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			profile.StoreKey,
			profile.AliasStoreKey,
			profile.CardStoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			profile.StoreKey:      app.DummyDecoder,
			profile.AliasStoreKey: app.DummyDecoder,
			profile.CardStoreKey:  app.Uint64Decoder,
			params.StoreKey:       app.DummyDecoder,
		},
		map[string]app.Decoder{
			profile.StoreKey:      app.DummyDecoder,
			profile.AliasStoreKey: app.DummyDecoder,
			profile.CardStoreKey:  app.DummyDecoder,
			params.StoreKey:       app.DummyDecoder,
		},
	)
}