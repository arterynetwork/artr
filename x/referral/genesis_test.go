// +build testing

package referral_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
	"github.com/arterynetwork/artr/x/schedule"
	"github.com/arterynetwork/artr/x/subscription"
)

func TestReferralGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         referral.Keeper
	subKeeper subscription.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx            = s.app.NewContext(true, abci.Header{Height: 1})
	s.k              = s.app.GetReferralKeeper()
	s.subKeeper      = s.app.GetSubscriptionKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestStatusDowngrade() {
	var (
		status referral.Status
		check  referral.StatusCheckResult
		err    error
	)
	var (
		user2 = app.DefaultGenesisUsers["user2"]
		user8 = app.DefaultGenesisUsers["user8"]
	)
	if status, err = s.k.GetStatus(s.ctx, user2); err != nil { panic(err) }
	s.subKeeper.SetActivityInfo(s.ctx, user8, subscription.NewActivityInfo(false, 0))
	if err := s.k.SetActive(s.ctx, user8, false); err != nil { panic(err) }
	// so, user2 loses its status
	if check, err = s.k.AreStatusRequirementsFulfilled(s.ctx, user2, status); err != nil { panic(err) }
	s.False(check.Overall)
	s.checkExportImport()
}

func (s Suite) TestCompression() {
	user1 := app.DefaultGenesisUsers["user1"]
	s.subKeeper.SetActivityInfo(s.ctx, user1, subscription.NewActivityInfo(false, 0))
	if err := s.k.SetActive(s.ctx, user1, false); err != nil { panic(err) }
	// so, compression is scheduled
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, referral.Params{
		CompanyAccounts:   referral.CompanyAccounts{
			TopReferrer:     user(10),
			ForSubscription: user(11),
			PromoBonuses:    user(12),
			StatusBonuses:   user(13),
			LeaderBonuses:   user(14),
			ForDelegating:   user(15),
		},
		DelegatingAward:   referral.NetworkAward{
			Network: [10]util.Fraction{
				util.Permille(1),
				util.Permille(3),
				util.Permille(5),
				util.Permille(7),
				util.Permille(9),
				util.Permille(11),
				util.Permille(13),
				util.Permille(15),
				util.Permille(17),
				util.Permille(19),
			},
			Company: util.Percent(21),
		},
		SubscriptionAward: referral.NetworkAward{
			Network: [10]util.Fraction{
				util.Permille(2),
				util.Permille(4),
				util.Permille(6),
				util.Permille(8),
				util.Permille(10),
				util.Permille(12),
				util.Permille(14),
				util.Permille(16),
				util.Permille(18),
				util.Permille(20),
			},
			Company: util.Percent(22),
		},
	})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
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
			referral.StoreKey: func(bz []byte) (string, error) {
				var result types.R
				err := s.app.Codec().UnmarshalBinaryLengthPrefixed(bz, &result)
				if err != nil { return "", err }
				return fmt.Sprintf("%+v", result), nil
			},
			schedule.StoreKey: app.DummyDecoder,
			params.StoreKey:   app.DummyDecoder,
		},)
}

func user(n int) sdk.AccAddress {
	return app.DefaultGenesisUsers[fmt.Sprintf("user%d", n)]
}
