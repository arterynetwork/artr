// +build testing

package delegating_test

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	delegatingK "github.com/arterynetwork/artr/x/delegating/keeper"
	delegating "github.com/arterynetwork/artr/x/delegating/types"
	schedule "github.com/arterynetwork/artr/x/schedule/types"
)

func TestDelegatingGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context
	k       delegatingK.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)
	s.k = s.app.GetDelegatingKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestDelegateAndRevoke() {
	user1 := app.DefaultGenesisUsers["user1"]
	if err := s.k.Delegate(s.ctx, user1, sdk.NewInt(10_000000)); err != nil {
		panic(err)
	}
	if err := s.k.Revoke(s.ctx, user1, sdk.NewInt(5_000000), false); err != nil {
		panic(err)
	}
	s.checkExportImport()
}

func (s *Suite) TestRevokeAll() {
	user := app.DefaultGenesisUsers["user1"]
	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(10_000000)))
	s.Equal(
		int64(20_009_970000),
		s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64(),
	) // -tx_fee -15%
	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(9_970000), false))

	s.Equal(
		int64(20_000_000000),
		s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64(),
	)
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, delegating.Params{
		MinDelegate:  123456,
		RevokePeriod: 28,
		BurnOnRevoke: util.Percent(50),
		Revoke: delegating.Revoke{
			Period: 28,
			Burn:   util.Percent(50),
		},
		ExpressRevoke: delegating.Revoke{
			Period: 14,
			Burn:   util.Percent(75),
		},
		AccruePercentageTable: []delegating.PercentageListRange{
			{Start: 0, PercentList: []util.Fraction{
				util.Percent(96),
				util.Percent(13),
				util.Percent(1),
				util.Percent(0),
				util.Percent(0),
			}},
			{Start: 1_000_000000, PercentList: []util.Fraction{
				util.Percent(97),
				util.Percent(13),
				util.Percent(1),
				util.Percent(0),
				util.Percent(0),
			}},
			{Start: 10_000_000000, PercentList: []util.Fraction{
				util.Percent(98),
				util.Percent(13),
				util.Percent(1),
				util.Percent(0),
				util.Percent(0),
			}},
			{Start: 100_000_000000, PercentList: []util.Fraction{
				util.Percent(99),
				util.Percent(13),
				util.Percent(1),
				util.Percent(0),
				util.Percent(0),
			}},
		},
	})
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		s.ctx.BlockTime(),
		[]string{
			delegating.MainStoreKey,
			schedule.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			delegating.MainStoreKey: app.AccAddressDecoder,
			schedule.StoreKey:       app.Uint64Decoder,
			params.StoreKey:         app.DummyDecoder,
		},
		map[string]app.Decoder{
			delegating.MainStoreKey: func(bz []byte) (string, error) {
				var data delegating.Record
				if err := proto.Unmarshal(bz, &data); err != nil {
					return "", err
				}
				return fmt.Sprintf("%+v", data), nil
			},
			schedule.StoreKey: app.ScheduleDecoder,
			params.StoreKey:   app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}
