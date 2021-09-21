// +build testing

package profile_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/profile/types"
)

func TestProfileGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context
	k       keeper.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)
	s.k = s.app.GetProfileKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestFullData() {
	genesis_time := s.ctx.BlockTime()
	_, _, newAcc := testdata.KeyTestPubAddr()
	s.NoError(
		s.k.CreateAccountWithProfile(s.ctx, newAcc, app.DefaultGenesisUsers["user13"], types.NewProfile(
			genesis_time.Add(42 * 30*time.Second),
			true,
			true,
			true,
			true,
			true,
			"FooBar",
			12345,
		)),
	)
	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.Panics(func() {
		s.k.SetParams(s.ctx, types.Params{
			BaseStorageGb:     13,
			BaseVpnGb:         14,
			StorageGbPrice:    15,
			VpnGbPrice:        16,
			SubscriptionPrice: 17,
			TokenRate:         util.NewFraction(1111111, 9),
			Creators:          []string{app.DefaultGenesisUsers["user5"].String()},
			StorageSigners:    []string{app.DefaultGenesisUsers["user6"].String()},
			VpnSigners:        []string{app.DefaultGenesisUsers["user7"].String()},
			TokenRateSigners:  []string{app.DefaultGenesisUsers["user8"].String()},
			RenamePrice:       1234,
			CardMagic:         999999666666,
		})
	})
	s.k.SetParams(s.ctx, types.Params{
		BaseStorageGb:     13,
		BaseVpnGb:         14,
		StorageGbPrice:    15,
		VpnGbPrice:        16,
		SubscriptionPrice: 17,
		TokenRate:         util.NewFraction(1111111, 9),
		Creators:          []string{app.DefaultGenesisUsers["user5"].String()},
		StorageSigners:    []string{app.DefaultGenesisUsers["user6"].String()},
		VpnSigners:        []string{app.DefaultGenesisUsers["user7"].String()},
		TokenRateSigners:  []string{app.DefaultGenesisUsers["user8"].String()},
		RenamePrice:       1234,
	})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		s.ctx.BlockTime(),
		[]string{
			types.StoreKey,
			types.AliasStoreKey,
			types.CardStoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			types.StoreKey:      app.DummyDecoder,
			types.AliasStoreKey: app.DummyDecoder,
			types.CardStoreKey:  app.Uint64Decoder,
			params.StoreKey:     app.DummyDecoder,
		},
		map[string]app.Decoder{
			types.StoreKey:      app.DummyDecoder,
			types.AliasStoreKey: app.DummyDecoder,
			types.CardStoreKey:  app.DummyDecoder,
			params.StoreKey:     app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}
