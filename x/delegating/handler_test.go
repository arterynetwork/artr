// +build testing

package delegating_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/delegating/types"
)

func TestDelegatingHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite

	app          *app.ArteryApp
	cleanup      func()
	ctx          sdk.Context
	k            delegating.Keeper
	supplyKeeper supply.Keeper
	accKeeper    auth.AccountKeeper
	handler      sdk.Handler
}

func (s *HandlerSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetDelegatingKeeper()
	s.supplyKeeper = s.app.GetSupplyKeeper()
	s.accKeeper = s.app.GetAccountKeeper()
	s.handler = delegating.NewHandler(s.k, s.supplyKeeper)
}

func (s *HandlerSuite) TearDownTest() { s.cleanup() }

func (s *HandlerSuite) TestDelegate_Fee() {
	user := app.DefaultGenesisUsers["root"]
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	// 0.3% fee
	msg := types.NewMsgDelegate(user, sdk.NewInt(1_000_000000))
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(3_000000),
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(847_450000), // = 1000 * 99.7% * 85%
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigDelegatedDenom).Int64(),
	)
	s.Equal(
		int64(999_000_000000), // 1000000(from genesis) - 10000(total)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	// maximal fee
	msg = types.NewMsgDelegate(user, sdk.NewInt(999_000_000000))
	_, err = s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(13_000000), // 3(initial) +10
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(849_988_950000), // = 847.45(initial) + (999000 - 10) * 85%
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigDelegatedDenom).Int64(),
	)
	s.Equal(
		int64(0),
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}

func (s *HandlerSuite) TestDelegate_BelowMinimum() {
	user := app.DefaultGenesisUsers["root"]
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	// < 0.001 ARTR
	msg := types.NewMsgDelegate(user, sdk.NewInt(999))
	_, err := s.handler(s.ctx, msg)
	s.Error(err)
}
