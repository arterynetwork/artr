// +build testing

package storage_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/storage"
)

func TestStorageGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         storage.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetStorageKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestFullData() {
	user1 := app.DefaultGenesisUsers["user1"]
	s.k.SetLimit(s.ctx, user1, 100500 * util.GBSize)
	s.k.SetCurrent(s.ctx, user1, 42 * util.GBSize + 12345)
	s.k.SetData(s.ctx, user1, []byte{0, 1, 2, 3, 4, 5, 6 ,7})
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			storage.StoreKey,
		},
		map[string]app.Decoder{
			storage.StoreKey: app.DummyDecoder,
		},
		map[string]app.Decoder{
			storage.StoreKey: app.DummyDecoder,
		},)
}