// +build testing

package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/app"
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
			ForSubscription: app.DefaultGenesisUsers["user2"].String(),
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
