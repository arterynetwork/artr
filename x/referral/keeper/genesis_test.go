// +build testing

package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
	schedule "github.com/arterynetwork/artr/x/schedule/types"
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
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)
	s.k = s.app.GetReferralKeeper()
}

func (s GenSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s GenSuite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s GenSuite) TestTransition() {
	subj := app.DefaultGenesisUsers["user4"].String()
	dest := app.DefaultGenesisUsers["user3"].String()
	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.checkExportImport()
}

func (s GenSuite) TestTransition_Declined() {
	subj := app.DefaultGenesisUsers["user4"].String()
	dest := app.DefaultGenesisUsers["user3"].String()
	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.NoError(s.k.CancelTransition(s.ctx, subj, false))
	s.checkExportImport()
}

func (s GenSuite) TestParams() {
	s.k.SetParams(s.ctx, referral.Params{
		CompanyAccounts: referral.CompanyAccounts{
			TopReferrer:     app.DefaultGenesisUsers["user1"].String(),
			ForSubscription: app.DefaultGenesisUsers["user2"].String(),
			ForDelegating:   app.DefaultGenesisUsers["user6"].String(),
		},
		DelegatingAward: referral.NetworkAward{
			Network: []util.Fraction{
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
			Network: []util.Fraction{
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
		TransitionPrice: 49_000000,
	})
	s.checkExportImport()
}

func (s GenSuite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		s.ctx.BlockTime(),
		[]string{
			referral.StoreKey,
			schedule.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			referral.StoreKey: app.StringDecoder,
			schedule.StoreKey: app.Uint64Decoder,
			params.StoreKey:   app.DummyDecoder,
		},
		map[string]app.Decoder{
			referral.StoreKey: s.RDecoder,
			schedule.StoreKey: app.ScheduleDecoder,
			params.StoreKey:   app.DummyDecoder,
		},
		map[string][][]byte{},
	)
}

func (s *GenSuite) RDecoder(bz []byte) (string, error) {
	var item types.Info
	err := s.app.Codec().UnmarshalBinaryBare(bz, &item)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%+v", item), nil
}
