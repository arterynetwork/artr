// +build testing

package vpn_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/vpn"
	"github.com/arterynetwork/artr/x/vpn/types"
)

func TestVpnGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         vpn.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetVpnKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestFullData() {
	s.k.SetInfo(s.ctx, app.DefaultGenesisUsers["user1"], types.VpnInfo{
		Current: 9000,
		Limit:   100500,
	})
	s.k.SetInfo(s.ctx, app.DefaultGenesisUsers["user2"], types.VpnInfo{
		Current: 7,
		Limit:   40,
	})
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, vpn.Params{Signers: []sdk.AccAddress{app.DefaultGenesisUsers["user7"]}})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			vpn.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			vpn.StoreKey:    app.DummyDecoder,
			params.StoreKey: app.DummyDecoder,
		},
		map[string]app.Decoder{
			vpn.StoreKey:    app.DummyDecoder,
			params.StoreKey: app.DummyDecoder,
		},
	)
}
