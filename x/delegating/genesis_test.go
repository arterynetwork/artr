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
	if err := s.k.Revoke(s.ctx, user1, sdk.NewInt(5_000000)); err != nil {
		panic(err)
	}
	s.checkExportImport()
}

func (s *Suite) TestRevokeAll() {
	user := app.DefaultGenesisUsers["user1"]
	s.NoError(s.k.Delegate(s.ctx, user, sdk.NewInt(10_000000)))
	s.Equal(
		int64(8_474500),
		s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).Int64(),
	)  // -tx_fee -15%
 	s.NoError(s.k.Revoke(s.ctx, user, sdk.NewInt(8_474500)))

	s.True(s.app.GetBankKeeper().GetBalance(s.ctx, user).AmountOf(util.ConfigDelegatedDenom).IsZero())
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
		MinDelegate:   123456,
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
			delegating.MainStoreKey: func(bz []byte) (string, error) {
				var data delegating.Record
				if err := proto.Unmarshal(bz, &data); err != nil {
					return "", err
				}
				return fmt.Sprintf("%+v", data), nil
			},
			delegating.ClusterStoreKey: app.DummyDecoder,
			schedule.StoreKey:          app.ScheduleDecoder,
			params.StoreKey:            app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}
