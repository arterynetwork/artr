// +build testing

package bank_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
)

func TestBankHandler(t *testing.T) {
	suite.Run(t, new(OriginalBankSuite))
	suite.Run(t, new(HandlerSuite))
}

type OriginalBankSuite struct{ suite.Suite }

func (s *OriginalBankSuite) TestInvalidMsg() {
	h := bank.NewHandler(nil, nil, nil)

	res, err := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg())
	require.Error(s.T(), err)
	require.Nil(s.T(), res)

	_, _, log := sdkerrors.ABCIInfo(err, false)
	require.True(s.T(), strings.Contains(log, "unrecognized bank message type"))
}

type HandlerSuite struct {
	suite.Suite

	app          *app.ArteryApp
	cleanup      func()
	ctx          sdk.Context
	k            bank.Keeper
	supplyKeeper supply.Keeper
	accKeeper    auth.AccountKeeper
	handler      sdk.Handler
}

func (s *HandlerSuite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetBankKeeper()
	s.supplyKeeper = s.app.GetSupplyKeeper()
	s.accKeeper = s.app.GetAccountKeeper()
	s.handler = bank.NewHandler(s.k, s.supplyKeeper, s.accKeeper)
}

func (s *HandlerSuite) TearDownTest() { s.cleanup() }

func (s *HandlerSuite) TestSend_TxFee() {
	userA := app.DefaultGenesisUsers["root"]
	s.Equal(
		int64(1_000_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, userA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	userB := app.DefaultGenesisUsers["user2"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, userB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := bank.NewMsgSend(userA, userB, util.Uartrs(1_000_000000))
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(3_000000), // = 1000 * 0.3%
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(2_000_000000), // = 1000(from genesis) + 1000
		s.accKeeper.GetAccount(s.ctx, userB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(998_997_000000), // 1000000(from genesis) - 1000 * 100.3%
		s.accKeeper.GetAccount(s.ctx, userA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}

func (s *HandlerSuite) TestSend_SendAll() {
	userA := app.DefaultGenesisUsers["user1"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, userA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	userB := app.DefaultGenesisUsers["user2"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, userB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := bank.NewMsgSend(userA, userB, util.Uartrs(1_000_000000)) // all the money
	_, err := s.handler(s.ctx, msg)
	s.Error(err)
	_, _, log := sdkerrors.ABCIInfo(err, false)
	require.True(s.T(), strings.Contains(log, "insufficient funds"))
}

func (s*HandlerSuite) TestSend_ToNowhere() {
	userA := app.DefaultGenesisUsers["user1"]
	s.Equal(
		int64(1_000_000000),
		s.accKeeper.GetAccount(s.ctx, userA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	userB := app.NonExistingUser
	s.Nil(s.accKeeper.GetAccount(s.ctx, userB))

	msg := bank.NewMsgSend(userA, userB, util.Uartrs(1_000))
	_, err := s.handler(s.ctx, msg)
	s.Error(err)
	_, _, log := sdkerrors.ABCIInfo(err, false)
	s.Contains(log, "account doesn't exist")

	s.Equal(
		int64(1_000_000000), // as it was
		s.accKeeper.GetAccount(s.ctx, userA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Nil(s.accKeeper.GetAccount(s.ctx, userB))
}

func (s *HandlerSuite) TestMultiSend_TxFee() {
	senderA := app.DefaultGenesisUsers["user1"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, senderA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	senderB := app.DefaultGenesisUsers["user2"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, senderB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	receiverA := app.DefaultGenesisUsers["user3"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, receiverA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	receiverB := app.DefaultGenesisUsers["user4"]
	s.Equal(
		int64(1_000_000000), // (from genesis)
		s.accKeeper.GetAccount(s.ctx, receiverB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)

	msg := bank.NewMsgMultiSend(
		[]bank.Input{
			bank.NewInput(senderA, util.Uartrs(700_000000)),
			bank.NewInput(senderB, util.Uartrs(800_000000)),
		},
		[]bank.Output{
			bank.NewOutput(receiverA, util.Uartrs(600_000000)),
			bank.NewOutput(receiverB, util.Uartrs(900_000000)),
		},
	)
	_, err := s.handler(s.ctx, msg)
	s.NoError(err)

	s.Equal(
		int64(4_500000), // = (700 + 800) * 0.3%
		s.supplyKeeper.GetModuleAccount(s.ctx, auth.FeeCollectorName).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(297_900000), // = 1000(from genesis) - 700 * 100.3%
		s.accKeeper.GetAccount(s.ctx, senderA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(197_600000), // = 1000(from genesis) - 800 * 100.3%
		s.accKeeper.GetAccount(s.ctx, senderB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(1_600_000000), // 1000(from genesis) + 600
		s.accKeeper.GetAccount(s.ctx, receiverA).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
	s.Equal(
		int64(1_900_000000), // 1000(from genesis) + 600
		s.accKeeper.GetAccount(s.ctx, receiverB).GetCoins().AmountOf(util.ConfigMainDenom).Int64(),
	)
}
