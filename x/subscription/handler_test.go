// +build testing

package subscription_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/subscription"
	"github.com/arterynetwork/artr/x/subscription/types"
	"github.com/arterynetwork/artr/x/vpn"
)

func TestSubscriptionHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite

	app          *app.ArteryApp
	cleanup      func()
	ctx          sdk.Context
	k            subscription.Keeper
	supplyKeeper supply.Keeper
	accKeeper    auth.AccountKeeper
	handler      sdk.Handler
}

func (s *HandlerSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx            = s.app.NewContext(true, abci.Header{Height: 1})
	s.k              = s.app.GetSubscriptionKeeper()
	s.supplyKeeper   = s.app.GetSupplyKeeper()
	s.accKeeper      = s.app.GetAccountKeeper()
	s.handler        = subscription.NewHandler(s.k)
}

func (s *HandlerSuite) TearDownTest() { s.cleanup() }

func (s *HandlerSuite) TestPayForSubscription_TxFee_NoExtraSpace() {
	user := app.DefaultGenesisUsers["root"]
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := types.NewMsgPaySubscription(user, 5 * util.GBSize)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(597000), // = 199 * 0.3%
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(3_968060), // = 199 * 99.7% * 2%
		s.supplyKeeper.GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(7_936120), // = 199 * 99.7% * 4%
		s.supplyKeeper.GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(999_801_000000), // 1000000(from genesis) - 199(total)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}

func (s *HandlerSuite) TestPayForSubscription_TxFee_ExtraSpace() {
	user := app.DefaultGenesisUsers["root"]
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := types.NewMsgPaySubscription(user, 15 * util.GBSize)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(627000), // = (199 + 10) * 0.3%
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(3_968060), // = 199 * 99.7% * 2%
		s.supplyKeeper.GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(17_906120), // = 199 * 99.7% * 4% + 10 * 99.7%
		s.supplyKeeper.GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(999_791_000000), // 1000000(from genesis) - (199 + 10)(total)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}

func (s *HandlerSuite) TestPayForVPN_TxFee() {
	user := app.DefaultGenesisUsers["root"]
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := types.NewMsgPayVPN(user, 10 * util.GBSize)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(30000), // = 10 * 0.3%
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(9_970000), // = 10 * 99.7%
		s.supplyKeeper.GetModuleAccount(s.ctx, vpn.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(999_990_000000), // 1000000(from genesis) - 10(total)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}

func (s *HandlerSuite) TestPayForStorage_TxFee() {
	user := app.DefaultGenesisUsers["root"]
	s.k.SetActivityInfo(s.ctx, user, types.NewActivityInfo(true, 1 + util.BlocksOneMonth))
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := types.NewMsgPayStorage(user, (5 + 10) * util.GBSize)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(30000), // = 10 * 0.3%
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(9_970000), // = 10 * 99.7%
		s.supplyKeeper.GetModuleAccount(s.ctx, storage.ModuleName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(999_990_000000), // 1000000(from genesis) - 10(total)
		s.accKeeper.GetAccount(s.ctx, user).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}
