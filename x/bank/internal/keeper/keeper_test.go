// +build testing

package keeper_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/bank/types"
)

func TestBankKeeper(t *testing.T) { suite.Run(t, new(Suite)) }

type Suite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context

	bk bank.Keeper

	accounts map[string]sdk.AccAddress
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	data, err := ioutil.ReadFile("test-genesis-bugcoins.json")
	if err != nil {
		panic(err)
	}
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(data)

	s.bk = s.app.GetBankKeeper()

	s.accounts = make(map[string]sdk.AccAddress, 2)
	s.accounts["pool"],  _ = sdk.AccAddressFromBech32("artrt1yhy6d3m4utltdml7w7zte7mqx5wyuskqppw34n")
	s.accounts["user1"], _ = sdk.AccAddressFromBech32("artrt1u574gq6xcplupp7jy65fkzcfmayr24wm8k2zgg")
}

func (s *Suite) TestSendBugcoin() {
	_, err := s.bk.Send(sdk.WrapSDKContext(s.ctx), &types.MsgSend{
		FromAddress: s.accounts["pool"].String(),
		ToAddress:   s.accounts["user1"].String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(util.ConfigBughuntingDenom, sdk.NewInt(1500000))),
	})
	s.NoError(err)

	cz := s.bk.GetBalance(s.ctx, s.accounts["user1"])
	s.Equal(int64(12345678), cz.AmountOf(util.ConfigMainDenom).Int64())
	s.Equal(int64(23456789), cz.AmountOf(util.ConfigDelegatedDenom).Int64())
	s.Equal(int64(0), cz.AmountOf(util.ConfigRevokingDenom).Int64())
	s.Equal(int64(1500000), cz.AmountOf(util.ConfigBughuntingDenom).Int64())
}
