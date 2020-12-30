// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/schedule"
)

func TestReferralGenesis(t *testing.T) {
	suite.Run(t, new(GenSuite))
}

type GenSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context
	k       referral.Keeper
}

func (s *GenSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetReferralKeeper()
}

func (s GenSuite) TearDownTest() {
	s.cleanup()
}

func (s GenSuite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s GenSuite) TestTransition() {
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.checkExportImport()
}

func (s GenSuite) TestTransition_Declined() {
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.NoError(s.k.CancelTransition(s.ctx, subj, false))
	s.checkExportImport()
}

func (s GenSuite) TestParams() {
	s.k.SetParams(s.ctx, referral.Params{
		CompanyAccounts:   referral.CompanyAccounts{
			TopReferrer:     app.DefaultGenesisUsers["user1"],
			ForSubscription: app.DefaultGenesisUsers["user2"],
			PromoBonuses:    app.DefaultGenesisUsers["user3"],
			StatusBonuses:   app.DefaultGenesisUsers["user4"],
			LeaderBonuses:   app.DefaultGenesisUsers["user5"],
			ForDelegating:   app.DefaultGenesisUsers["user6"],
		},
		DelegatingAward:   referral.NetworkAward{
			Network: [10]util.Fraction{
				util.Percent(1), 
				util.Percent(2), 
				util.Percent(3),
				util.Percent(4),
				util.Percent(5),
				util.Percent(6),
				util.Percent(7),
				util.Percent(8),
				util.Percent(9),
				util.Percent(10),
			},
			Company: util.Percent(13),
		},
		SubscriptionAward: referral.NetworkAward{
			Network: [10]util.Fraction{
				util.Permille(1),
				util.Permille(2),
				util.Permille(3),
				util.Permille(4),
				util.Permille(5),
				util.Permille(6),
				util.Permille(7),
				util.Permille(8),
				util.Permille(9),
				util.Permille(10),
			},
			Company: util.Permille(13),
		},
		TransitionCost:    49_000000,
	})
	s.checkExportImport()
}

func (s GenSuite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			referral.StoreKey,
			schedule.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			referral.StoreKey: app.AccAddressDecoder,
			schedule.StoreKey: app.Uint64Decoder,
			params.StoreKey:   app.DummyDecoder,
		},
		map[string]app.Decoder{
			referral.StoreKey: app.DummyDecoder,
			schedule.StoreKey: app.DummyDecoder,
			params.StoreKey:   app.DummyDecoder,
		},
		map[string][][]byte{},
	)
}
