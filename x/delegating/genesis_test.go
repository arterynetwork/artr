// +build testing

package delegating_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/schedule"
)

func TestDelegatingGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         delegating.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetDelegatingKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestDelegateAndRevoke() {
	user1 := app.DefaultGenesisUsers["user1"]
	if err := s.k.Delegate(s.ctx, user1, sdk.NewInt(10_000000)); err != nil { panic(err) }
	if err := s.k.Revoke(s.ctx, user1, sdk.NewInt(5_000000)); err != nil { panic(err) }
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, delegating.Params{
		Percentage: delegating.Percentage{
			Minimal:      96,
			ThousandPlus: 97,
			TenKPlus:     98,
			HundredKPlus: 99,
		},
	})
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			delegating.MainStoreKey,
			delegating.ClusterStoreKey,
			schedule.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			delegating.MainStoreKey:    app.AccAddressDecoder,
			delegating.ClusterStoreKey: app.DummyDecoder,
			schedule.StoreKey:          app.Uint64Decoder,
			params.StoreKey:            app.DummyDecoder,
		},
		map[string]app.Decoder{
			delegating.MainStoreKey:    app.DummyDecoder,
			delegating.ClusterStoreKey: app.DummyDecoder,
			schedule.StoreKey:          app.DummyDecoder,
			params.StoreKey:            app.DummyDecoder,
		},)
}
