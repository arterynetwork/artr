// +build testing

package subscription_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/schedule"
	"github.com/arterynetwork/artr/x/subscription"
	"github.com/arterynetwork/artr/x/subscription/types"
)

func TestSubscriptionGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         subscription.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetSubscriptionKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestSetInfoAndScheduleRenew() {
	user13 := app.DefaultGenesisUsers["user13"]
	s.k.SetActivityInfo(s.ctx, user13, types.NewActivityInfo(true, 78000))
	s.k.ScheduleRenew(s.ctx, user13, 78000)
	s.k.SetActivityInfo(s.ctx, app.DefaultGenesisUsers["user15"], types.NewActivityInfo(false, -1))
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, subscription.Params{
		TokenCourse:         9994,
		SubscriptionPrice:   9995,
		VPNGBPrice:          9996,
		StorageGBPrice:      9997,
		BaseVPNGb:           9998,
		BaseStorageGb:       9999,
		CourseChangeSigners: []sdk.AccAddress{app.DefaultGenesisUsers["user13"]},
	})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			subscription.StoreKey,
			schedule.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			subscription.StoreKey: app.DummyDecoder,
			schedule.StoreKey:     app.DummyDecoder,
			params.StoreKey:       app.DummyDecoder,
		},
		map[string]app.Decoder{
			subscription.StoreKey: app.DummyDecoder,
			schedule.StoreKey:     app.DummyDecoder,
			params.StoreKey:       app.DummyDecoder,
		},)
}