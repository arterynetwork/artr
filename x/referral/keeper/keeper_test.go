// +build testing

package keeper_test

import (
	"github.com/arterynetwork/artr/x/subscription"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	delegatingK "github.com/arterynetwork/artr/x/delegating/keeper"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
)

func TestReferralKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
	suite.Run(t, new(TransitionBorderlineSuite))
	suite.Run(t, new(StatusUpgradeSuite))
	suite.Run(t, new(Status3x3Suite))
	suite.Run(t, new(StatusBonusSuite))
}

type BaseSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc       *codec.Codec
	ctx       sdk.Context
	k         referral.Keeper
	storeKey  sdk.StoreKey
	accKeeper auth.AccountKeeper
}

func (s *BaseSuite) setupTest(genesis json.RawMessage) {
	defer func() {
		if err := recover(); err != nil {
			s.FailNow("panic in setup", err)
		}
	}()

	s.app, s.cleanup = app.NewAppFromGenesis(genesis)

	s.cdc = s.app.Codec()
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k = s.app.GetReferralKeeper()
	s.storeKey = s.app.GetKeys()[referral.ModuleName]
	s.accKeeper = s.app.GetAccountKeeper()
}

func (s *BaseSuite) TearDownTest() {
	if s.cleanup == nil {
		s.FailNow("cleanup callback is not set")
	}
	s.cleanup()
}

type Suite struct{
	BaseSuite

	pk subscription.Keeper
	dk delegatingK.Keeper
}

func (s *Suite) SetupTest() {
	s.setupTest(nil)
	s.pk = s.app.GetSubscriptionKeeper()
	s.dk = s.app.GetDelegatingKeeper()
}

var (
	THOUSAND = util.Uartrs(1_000_000000)
	STAKE    = sdk.NewCoins(
		sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
		sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(10_000_000000)),
	)
)

func (s *Suite) TestAppendChild() {
	accounts := [12]sdk.AccAddress{}
	for i := 0; i < 12; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.Nil(
			s.setBalance(addr, sdk.Coins{sdk.Coin{
				Denom:  util.ConfigMainDenom,
				Amount: sdk.NewInt(1 << i),
			}}),
		)
	}

	s.Nil(s.set(accounts[0], types.NewR(nil, sdk.NewInt(1), sdk.ZeroInt())))
	s.Nil(s.k.SetActive(s.ctx, accounts[0], true))

	for i := 0; i <= 10; i++ {
		s.Nil(s.k.AppendChild(s.ctx, accounts[i], accounts[i+1]))
		s.Nil(s.k.SetActive(s.ctx, accounts[i+1], true))
	}

	for i, expected := range [12][11]int64{
		{0x0001, 0x0002, 0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400},
		{0x0002, 0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0080, 0x0100, 0x0200, 0x0400, 0x0800},
		{0x0100, 0x0200, 0x0400, 0x0800},
		{0x0200, 0x0400, 0x0800},
		{0x0400, 0x0800},
		{0x0800},
	} {
		value, err := s.get(accounts[i])
		s.Nilf(err, "Get account #%d", i)
		for j := 0; j <= 10; j++ {
			s.Equalf(
				expected[j], value.Coins[j].Int64(),
				"Coins at lvl #%d for item #%d", j, i)
		}

		if i == 0 {
			s.Nil(value.Referrer, "GetParent #0")
		} else {
			s.Equalf(
				accounts[i-1],
				value.Referrer,
				"GetParent #%d", i,
			)
		}

		if i == 11 {
			s.Empty(value.Referrals, "GetChildren #11")
		} else {
			s.Equalf(
				[]sdk.AccAddress{accounts[i+1]},
				value.Referrals,
				"GetChildren #%d", i,
			)
		}

		expectedRefCount := [11]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		for j := 10; j > 11-i; j-- {
			expectedRefCount[j] = 0
		}
		s.Equalf(
			expectedRefCount,
			value.ActiveReferralsCount,
			"ActiveReferralsCount #%d", i,
		)
	}
}

func (s *Suite) TestGetters() {
	_, _, acc := authtypes.KeyTestPubAddr()
	_, _, parent := authtypes.KeyTestPubAddr()
	_, _, child1 := authtypes.KeyTestPubAddr()
	_, _, child2 := authtypes.KeyTestPubAddr()
	s.Nil(
		s.set(acc, types.R{
			Status:    types.Hero,
			Referrer:  parent,
			Referrals: []sdk.AccAddress{child1, child2},
			//			Coins:                [11]sdk.Int{},
			//			Delegated:            [11]sdk.Int{},
			//			Active:               false,
			//			ActiveReferralsCount: [11]int{},
		}),
	)

	resultStatus, err := s.k.GetStatus(s.ctx, acc)
	s.Nil(err, "GetStatus without error")
	s.Equal(types.Hero, resultStatus, "GetStatus")

	resultParent, err := s.k.GetParent(s.ctx, acc)
	s.Nil(err, "GetParent without error")
	s.Equal(parent, resultParent, "GetParent")

	resultChildren, err := s.k.GetChildren(s.ctx, acc)
	s.Nil(err, "GetChildren without error")
	s.Equal([]sdk.AccAddress{child1, child2}, resultChildren, "GetChildren")
}

func (s *Suite) TestGetCoinsInNetwork() {
	accounts := [12]sdk.AccAddress{}
	for i := 0; i < 12; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.Nil(
			s.setBalance(addr, sdk.Coins{
				sdk.Coin{
					Denom:  util.ConfigMainDenom,
					Amount: sdk.NewInt(1 << (2 * i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i + 1)),
				},
			}),
		)
	}
	s.Nil(s.set(accounts[0], types.R{
		Status:               types.Leader,
		StatusDowngradeAt:    -1,
		Active:               true,
		ActiveReferralsCount: [11]int{1},
		Coins:                [11]sdk.Int{sdk.NewInt(3)},
		Delegated:            [11]sdk.Int{sdk.NewInt(2)},
	}))

	//                  0
	//              ┌──┘ └──┐
	//              1       B
	//          ┌──┘ └──┐
	//          2       7
	//          │       │
	//          3       8
	//      ┌──┘ └──┐   │
	//      4       6   9
	//  ═══ │ ═════════ │ ═════ end of open lines
	//      5           A
	s.Nil(s.k.AppendChild(s.ctx, accounts[0], accounts[1]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[1], accounts[2]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[2], accounts[3]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[3], accounts[4]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[4], accounts[5]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[3], accounts[6]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[1], accounts[7]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[7], accounts[8]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[8], accounts[9]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[9], accounts[10]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[0], accounts[11]))
	for i := 1; i <= 11; i++ {
		s.Nil(s.k.SetActive(s.ctx, accounts[i], true))
	}

	res, err := s.k.GetCoinsInNetwork(s.ctx, accounts[0], 10)
	s.Nil(err, "GetCoinsInNetwork")
	s.Equal(uint64(0x00CFF3FF), res.Uint64(), "GetCoinsInNetwork")

	res, err = s.k.GetDelegatedInNetwork(s.ctx, accounts[0], 10)
	s.Nil(err, "GetDelegatedInNetwork")
	s.Equal(uint64(0x008AA2AA), res.Uint64(), "GetDelegatedInNetwork")
}

func (s *Suite) TestReferralFees() {
	accounts := [12]sdk.AccAddress{}
	for i := 0; i < 12; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.Nil(
			s.setBalance(addr, sdk.Coins{sdk.Coin{
				Denom:  util.ConfigMainDenom,
				Amount: sdk.NewInt(1),
			}}),
		)
	}
	s.NoError(
		s.set(accounts[0], types.R{
			Status:    types.Lucky,
			Coins:     [11]sdk.Int{sdk.NewInt(1)},
			Delegated: [11]sdk.Int{},
		}),
	)
	s.NoError(s.k.SetActive(s.ctx, accounts[0], true))
	for i := 0; i < 12-1; i++ {
		s.NoError(s.k.AppendChild(s.ctx, accounts[i], accounts[i+1]))
	}
	for i := 0; i < 12; i++ {
		s.NoError(s.k.SetActive(s.ctx, accounts[i], true))
	}

	var companyAccs types.CompanyAccounts
	s.app.GetSubspaces()[referral.DefaultParamspace].Get(s.ctx, types.KeyCompanyAccounts, &companyAccs)

	res, err := s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating all newbies: no error")
	s.Equal(4, len(res), "GetReferralFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(5),
	}, "GetReferralFesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all newbies: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio:       util.Permille(5),
	}, "GetReferralFesForDelegating all newbies: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Permille(85),
	}, "GetReferralFesForDelegating all newbies: \"top referrer\"")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription all newbies: no error")
	s.Equal(7, len(res), "GetReferralFeesForSubscription all newbies: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(15),
	}, "GetReferralFeesForSubscription all newbies: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio:       util.Percent(10),
	}, "GetReferralFeesForSubscription all newbies: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio:       util.Percent(10),
	}, "GetReferralFeesForSubscription all newbies: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Percent(44),
	}, "GetReferralFeesForSubscription all newbies: \"top referrer\"")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all newbies: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all newbies: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all newbies: status bonus")

	for i := 0; i < 12; i++ {
		s.Nil(s.update(accounts[i], func(value *types.R) {
			value.Status = types.AbsoluteChampion
		}))
	}

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating all pros: no error")
	s.Equal(11, len(res), "GetReferralFeesForDelegating all pros: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(5),
	}, "GetReferralFesForDelegating all pros: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[8],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 3")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[7],
		Ratio:       util.Percent(2),
	}, "GetReferralFesForDelegating all pros: lvl 4")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[6],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 5")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[5],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 6")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[4],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 7")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[3],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 8")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[2],
		Ratio:       util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 9")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[1],
		Ratio:       util.Permille(5),
	}, "GetReferralFesForDelegating all pros: lvl 10")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio:       util.Permille(5),
	}, "GetReferralFesForDelegating all pros: company")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription all pros: no error")
	s.Equal(14, len(res), "GetReferralFeesForSubscription all pros: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(15),
	}, "GetReferralFeesForSubscription all pros: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio:       util.Percent(10),
	}, "GetReferralFeesForSubscription all pros: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[8],
		Ratio:       util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 3")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[7],
		Ratio:       util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 4")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[6],
		Ratio:       util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 5")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[5],
		Ratio:       util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 6")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[4],
		Ratio:       util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 7")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[3],
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: lvl 8")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[2],
		Ratio:       util.Percent(2),
	}, "GetReferralFeesForSubscription all pros: lvl 9")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[1],
		Ratio:       util.Percent(2),
	}, "GetReferralFeesForSubscription all pros: lvl 10")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio:       util.Percent(10),
	}, "GetReferralFeesForSubscription all pros: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: status bonus")

	s.Nil(s.update(accounts[10], func(value *types.R) {
		value.Referrer = nil
	}))

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating short chain: no error")
	s.Equal(3, len(res), "GetReferralFeesForDelegating short chain: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(5),
	}, "GetReferralFesForDelegating short chain: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio:       util.Permille(5),
	}, "GetReferralFesForDelegating short chain: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Permille(95),
	}, "GetReferralFesForDelegating short chain: \"top referrer\"")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription short chain: no error")
	s.Equal(6, len(res), "GetReferralFeesForSubscription short chain: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(15),
	}, "GetReferralFeesForSubscription short chain: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio:       util.Percent(10),
	}, "GetReferralFeesForSubscription short chain: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Percent(54),
	}, "GetReferralFeesForSubscription short chain: \"top referrer\"")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription short chain: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription short chain: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription short chain: status bonus")

	s.Nil(s.update(accounts[11], func(value *types.R) {
		value.Referrer = nil
	}))

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating top account: no error")
	s.Equal(2, len(res), "GetReferralFeesForDelegating top account: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio:       util.Permille(5),
	}, "GetReferralFesForDelegating top account: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Permille(145),
	}, "GetReferralFesForDelegating top account: \"top referrer\"")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription top account: no error")
	s.Equal(5, len(res), "GetReferralFeesForSubscription top account: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio:       util.Percent(10),
	}, "GetReferralFeesForSubscription top account: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Percent(69),
	}, "GetReferralFeesForSubscription top account: \"top referrer\"")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription top account: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription top account: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio:       util.Percent(5),
	}, "GetReferralFeesForSubscription top account: status bonus")
}

func (s *Suite) TestCompression() {
	accounts := [10]sdk.AccAddress{}
	for i := 0; i < 10; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.Nil(
			s.setBalance(addr, sdk.Coins{
				sdk.Coin{
					Denom:  util.ConfigMainDenom,
					Amount: sdk.NewInt(1 << (2 * i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i + 1)),
				},
			}),
		)
	}
	s.Nil(s.set(accounts[0], types.R{
		Status:               types.Lucky,
		StatusDowngradeAt:    -1,
		Active:               true,
		ActiveReferralsCount: [11]int{1},
		Coins:                [11]sdk.Int{sdk.NewInt(3)},
		Delegated:            [11]sdk.Int{sdk.NewInt(2)},
		CompressionAt:        -1,
	}))

	//           0                        0
	//        ┌──┴──┐                ┌────┴──┐
	//        1     9                1       9
	//    ┌───┴──┐             ┌───┬─┴─┬───┐
	//    2     (4)            2  (4)  5   8
	//    │   ┌──┴──┐          │    ┌──┴──┐
	//    3   5     8          3    6     7
	//     ┌──┴──┐
	//     6      7
	s.Nil(s.k.AppendChild(s.ctx, accounts[0], accounts[1]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[1], accounts[2]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[2], accounts[3]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[1], accounts[4]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[4], accounts[5]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[5], accounts[6]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[5], accounts[7]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[4], accounts[8]))
	s.Nil(s.k.AppendChild(s.ctx, accounts[0], accounts[9]))
	for i := 1; i <= 9; i++ {
		s.Nil(s.k.SetActive(s.ctx, accounts[i], true))
	}

	s.Nil(s.k.SetActive(s.ctx, accounts[4], false))
	s.Nil(s.k.Compress(s.ctx, accounts[4]))

	zero := sdk.ZeroInt()
	for i, expected := range [10]types.R{
		{ // item #0
			Status:            types.Lucky,
			StatusDowngradeAt: -1,
			Active:            true,
			Referrer:          nil,
			Referrals: []sdk.AccAddress{
				accounts[1],
				accounts[9],
			},
			ActiveReferralsCount: [11]int{1, 2, 3, 3},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x000003),
				sdk.NewInt(0x0C000C),
				sdk.NewInt(0x030F30),
				sdk.NewInt(0x00F0C0),
				zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x000002),
				sdk.NewInt(0x080008),
				sdk.NewInt(0x020A20),
				sdk.NewInt(0x00A080),
				zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #1
			Status:            types.Lucky,
			StatusDowngradeAt: -1,
			Active:            true,
			Referrer:          accounts[0],
			Referrals: []sdk.AccAddress{
				accounts[2],
				accounts[4],
				accounts[5],
				accounts[8],
			},
			ActiveReferralsCount: [11]int{1, 3, 3},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x00000C),
				sdk.NewInt(0x030F30),
				sdk.NewInt(0x00F0C0),
				zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x000008),
				sdk.NewInt(0x020A20),
				sdk.NewInt(0x00A080),
				zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #2
			Status:            types.Lucky,
			StatusDowngradeAt: -1,
			Active:            true,
			Referrer:          accounts[1],
			Referrals: []sdk.AccAddress{
				accounts[3],
			},
			ActiveReferralsCount: [11]int{1, 1},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x000030),
				sdk.NewInt(0x0000C0),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x000020),
				sdk.NewInt(0x000080),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #3
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[2],
			ActiveReferralsCount: [11]int{1},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x0000C0),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x000080),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #4
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               false,
			Referrer:             accounts[1],
			ActiveReferralsCount: [11]int{},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x000300),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x000200),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #5
			Status:            types.Lucky,
			StatusDowngradeAt: -1,
			Active:            true,
			Referrer:          accounts[1],
			Referrals: []sdk.AccAddress{
				accounts[6],
				accounts[7],
			},
			ActiveReferralsCount: [11]int{1, 2},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x000C00),
				sdk.NewInt(0x00F000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x000800),
				sdk.NewInt(0x00A000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #6
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[5],
			ActiveReferralsCount: [11]int{1},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x003000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x002000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #7
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[5],
			ActiveReferralsCount: [11]int{1},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x00C000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x008000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #8
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[1],
			ActiveReferralsCount: [11]int{1},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x030000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x020000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
		{ // item #9
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[0],
			ActiveReferralsCount: [11]int{1},
			Coins: [11]sdk.Int{
				sdk.NewInt(0x0C0000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: [11]sdk.Int{
				sdk.NewInt(0x080000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt: -1,
		},
	} {
		value, err := s.get(accounts[i])
		s.Nilf(err, "get item #%d without error", i)
		s.Equalf(expected, value, "value of item #%d", i)
	}
}

func (s *Suite) TestAddChildJustBeforeCompression() {
	user1 := app.DefaultGenesisUsers["user1"]

	accounts := [3]sdk.AccAddress{}
	for i := 0; i < 3; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.NoError(
			s.setBalance(addr, sdk.Coins{
				sdk.Coin{
					Denom:  util.ConfigMainDenom,
					Amount: sdk.NewInt(1 << (2 * i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i + 1)),
				},
			}),
		)
	}
	// Right now
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[0]))

	// When tariff is over
	s.ctx = s.ctx.WithBlockHeight(8999)
	s.nextBlock()
	info, err := s.get(user1)
	s.NoError(err)
	s.False(info.Active)
	s.False(info.RegistrationClosed(s.ctx))
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[1]))

	// One month later
	s.ctx = s.ctx.WithBlockHeight(9001+86400)
	s.nextBlock()
	info, err = s.get(user1)
	s.NoError(err)
	s.False(info.Active)
	s.True(info.RegistrationClosed(s.ctx))
	s.Error(s.k.AppendChild(s.ctx, user1, accounts[2]))
}

func (s *Suite) TestAddChildAfterCompression() {
	user1 := app.DefaultGenesisUsers["user1"]

	accounts := [2]sdk.AccAddress{}
	for i := 0; i < 2; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.NoError(
			s.setBalance(addr, sdk.Coins{
				sdk.Coin{
					Denom:  util.ConfigMainDenom,
					Amount: sdk.NewInt(1 << (2 * i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i + 1)),
				},
			}),
		)
	}
	// Right now
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[0]))

	// When tariff is over
	s.ctx = s.ctx.WithBlockHeight(8999)
	s.nextBlock()
	info, err := s.get(user1)
	s.NoError(err)
	s.False(info.Active)
	s.False(info.RegistrationClosed(s.ctx))
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[1]))

	// After compression
	s.ctx = s.ctx.WithBlockHeight(8999+2*86400)
	s.nextBlock()
	info, err = s.get(user1)
	s.NoError(err)
	s.Zero(len(info.Referrals))
	s.True(info.RegistrationClosed(s.ctx))
	s.Error(s.k.AppendChild(s.ctx, user1, accounts[1]))
}

func (s *Suite) TestAddChildAfterReactivation() {
	user1 := app.DefaultGenesisUsers["user1"]

	accounts := [1]sdk.AccAddress{}
	for i := 0; i < 1; i++ {
		_, _, addr := authtypes.KeyTestPubAddr()
		accounts[i] = addr
		s.NoError(
			s.setBalance(addr, sdk.Coins{
				sdk.Coin{
					Denom:  util.ConfigMainDenom,
					Amount: sdk.NewInt(1 << (2 * i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i + 1)),
				},
			}),
		)
	}

	// Tariff is over
	s.ctx = s.ctx.WithBlockHeight(8999)
	s.nextBlock()
	info, err := s.get(user1)
	s.NoError(err)
	s.False(info.Active)

	// Compression
	s.ctx = s.ctx.WithBlockHeight(8999+2*86400)
	s.nextBlock()
	info, err = s.get(user1)
	s.NoError(err)
	s.Zero(len(info.Referrals))

	// Pay tariff
	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user1"], 5 * util.GBSize))
	info, err = s.get(user1)
	s.NoError(err)
	s.True(info.Active)
	s.False(info.RegistrationClosed(s.ctx))
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[0]))
}

func (s *Suite) TestStatusDowngrade() {
	if err := s.k.Compress(s.ctx, app.DefaultGenesisUsers["user4"]); err != nil {
		panic(err)
	}
	// After that, user2 does not fulfill level2 requirements anymore

	addr := app.DefaultGenesisUsers["user2"]
	if r, err := s.get(addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLeader, r.Status)
		s.Equal(int64(86401), r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLeader, status)
	}

	// Next block (nothing should happen) ...
	s.nextBlock()
	if r, err := s.get(addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLeader, r.Status)
		s.Equal(int64(86401), r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLeader, status)
	}

	// One month later
	s.ctx = s.ctx.WithBlockHeight(86400)
	s.nextBlock()
	if r, err := s.get(addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLucky, r.Status)
		s.Equal(int64(-1), r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLucky, status)
	}
}

func (s Suite) TestTransition() {
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)

	s.NoError(s.k.AffirmTransition(s.ctx, subj), "affirm transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)

	acc, err := s.k.GetParent(s.ctx, subj)
	s.NoError(err, "get parent")
	s.Equal(dest, acc, "new parent")

	accz, err := s.k.GetChildren(s.ctx, oldParent)
	s.NoError(err, "get old parent's children")
	s.Equal(
		[]sdk.AccAddress{app.DefaultGenesisUsers["user5"]},
		accz, "old parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, dest)
	s.NoError(err, "get new parent's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user6"],
			app.DefaultGenesisUsers["user7"],
			subj,
		},
		accz, "new parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, subj)
	s.NoError(err, "get subject's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user8"],
			app.DefaultGenesisUsers["user9"],
		},
		accz, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj)
	s.NoError(err, "get pending transition")
	s.Nil(acc, "pending transition")

	for i, n := range []int64{
		34_990_000000,
		14_000_000000, 19_990_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)], 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

func (s Suite) TestTransition_Decline() {
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)

	s.NoError(s.k.CancelTransition(s.ctx, subj, false), "decline transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)

	acc, err := s.k.GetParent(s.ctx, subj)
	s.NoError(err, "get parent")
	s.Equal(oldParent, acc, "new parent")

	accz, err := s.k.GetChildren(s.ctx, oldParent)
	s.NoError(err, "get old parent's children")
	s.Equal(
		[]sdk.AccAddress{
			subj,
			app.DefaultGenesisUsers["user5"],
		},
		accz, "old parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, dest)
	s.NoError(err, "get new parent's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user6"],
			app.DefaultGenesisUsers["user7"],
		},
		accz, "new parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, subj)
	s.NoError(err, "get subject's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user8"],
			app.DefaultGenesisUsers["user9"],
		},
		accz, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj)
	s.NoError(err, "get pending transition")
	s.Nil(acc, "pending transition")

	for i, n := range []int64{
		34_990_000000,
		16_990_000000, 17_000_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)], 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

func (s Suite) TestTransition_Timeout() {
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.NoError(s.k.RequestTransition(s.ctx, subj, dest), "request transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)

	for i, n := range []sdk.Coins{
		THOUSAND,
		STAKE, STAKE,
		util.Uartrs(990_000000), THOUSAND, THOUSAND, THOUSAND, // transition fee
		THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND,
	} {
		cz := s.accKeeper.GetAccount(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)]).GetCoins()
		s.Equal(n, cz)
	}
	for i, n := range []int64{
		34_990_000000,
		16_990_000000, 17_000_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)], 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}

	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneDay)
	s.nextBlock()
	for i, n := range []sdk.Coins{
		util.Uartrs(1_010_000000), // validator's award
		STAKE, STAKE,
		util.Uartrs(990_000000), THOUSAND, THOUSAND, THOUSAND,
		THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND,
	} {
		cz := s.accKeeper.GetAccount(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)]).GetCoins()
		s.Equal(n, cz)
	}

	acc, err := s.k.GetParent(s.ctx, subj)
	s.NoError(err, "get parent")
	s.Equal(oldParent, acc, "new parent")

	accz, err := s.k.GetChildren(s.ctx, oldParent)
	s.NoError(err, "get old parent's children")
	s.Equal(
		[]sdk.AccAddress{
			subj,
			app.DefaultGenesisUsers["user5"],
		},
		accz, "old parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, dest)
	s.NoError(err, "get new parent's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user6"],
			app.DefaultGenesisUsers["user7"],
		},
		accz, "new parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, subj)
	s.NoError(err, "get subject's children")
	s.Equal(
		[]sdk.AccAddress{
			app.DefaultGenesisUsers["user8"],
			app.DefaultGenesisUsers["user9"],
		},
		accz, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj)
	s.NoError(err, "get pending transition")
	s.Nil(acc, "pending transition")

	for i, n := range []int64{
		35_000_000000,
		16_990_000000, 17_000_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)], 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

func (s Suite) TestTransition_Validate_Circle() {
	subj := app.DefaultGenesisUsers["user2"]
	dest := app.DefaultGenesisUsers["user5"]

	s.EqualError(
		s.k.RequestTransition(s.ctx, subj, dest),
		"transition is invalid: cycles are not allowed",
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(10_000_000000)),
		),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)
}

func (s Suite) TestTransition_Validate_Self() {
	subj := app.DefaultGenesisUsers["user2"]

	s.EqualError(
		s.k.RequestTransition(s.ctx, subj, subj),
		"transition is invalid: subject cannot be their own referral",
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(10_000_000000)),
		),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)
}

func (s Suite) TestTransition_Validate_OldParent() {
	subj := app.DefaultGenesisUsers["user4"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.EqualError(
		s.k.RequestTransition(s.ctx, subj, oldParent),
		"transition is invalid: destination address is already subject's referrer",
	)
	s.Equal(
		util.Uartrs(1_000_000000),
		s.app.GetAccountKeeper().GetAccount(s.ctx, subj).GetCoins(),
	)
}

func (s Suite) TestBanishment() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"]
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, user, sdk.NewInt( 10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(8999).WithBlockTime(genesisTime.Add(8999*30*time.Second))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5 * util.GBSize))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.NotZero(len(info.Referrals))
	s.Equal(types.Leader, info.Status)

	s.ctx = s.ctx.WithBlockHeight(8999+2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second+2*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5 * util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.Equal(types.Lucky, info.Status)

	s.ctx = s.ctx.WithBlockHeight(8999+3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second+3*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5 * util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.True(info.Banished)
	s.Equal(types.Status(0), info.Status)
	s.Equal(parent, info.Referrer)

	info, err = s.get(parent)
	s.NotContains(info.Referrals, user)
	s.NoError(err)
}

type TransitionBorderlineSuite struct {
	BaseSuite

	accounts map[string]sdk.AccAddress
}

func (s *TransitionBorderlineSuite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()

	data, err := ioutil.ReadFile("test-genesis-transitions.json")
	if err != nil {
		panic(err)
	}
	s.setupTest(data)

	s.accounts = map[string]sdk.AccAddress{
		"1":     accAddr("artr1qq9gvskgjkwfkqexeapwps0cnqj6pxkz4nevre"),
		"1.1":   accAddr("artr1qqxwvzmhjsrwa9fuyafu2jcxcrv2fclwrpy33g"),
		"1.1.1": accAddr("artr1qqvnckqa5yqaps2v9wfeqpzkum4cmexcmr38kj"),
		"1.1.2": accAddr("artr1pg635yjdpg62pjvsxfz5xyhxcxk2ss4lkepp7x"),
		"2":     accAddr("artr1sxwwflxyj2wl0l3ltl83kn7sxvrkfalymmhvf0"),
		"2.1":   accAddr("artr1sxnhvuyuac9x52lmpduyf9uaz763nw0wwdu5qm"),
		"2.1.1": accAddr("artr1sx48ywhy3yqyhf4h4yxc4n2ucz62xkzva3e7d8"),
		"2.1.2": accAddr("artr13366fwedzhlu7l66kmrq3utq9x5y0f7f46yzj9"),
	}
}

func (s TransitionBorderlineSuite) TestAlmostUp() {
	data, err := s.k.Get(s.ctx, s.accounts["2"])
	s.NoError(err)
	s.Equal(referral.StatusTopLeader, data.Status)

	s.NoError(s.k.RequestTransition(s.ctx, s.accounts["2.1.1"], s.accounts["2.1.2"]))
	s.NoError(s.k.AffirmTransition(s.ctx, s.accounts["2.1.1"]))

	data, err = s.k.Get(s.ctx, s.accounts["2"])
	s.NoError(err)
	s.Equal(referral.StatusTopLeader, data.Status)
	s.Equal(int64(-1), data.StatusDowngradeAt)
}

func (s TransitionBorderlineSuite) TestAlmostDown() {
	data, err := s.k.Get(s.ctx, s.accounts["1"])
	s.NoError(err)
	s.Equal(referral.StatusHero, data.Status)

	s.NoError(s.k.RequestTransition(s.ctx, s.accounts["1.1.1"], s.accounts["1.1.2"]))
	s.NoError(s.k.AffirmTransition(s.ctx, s.accounts["1.1.1"]))

	scr ,err := s.k.AreStatusRequirementsFulfilled(s.ctx, s.accounts["1"], referral.StatusHero)
	s.NoError(err)
	s.True(scr.Overall)

	data, err = s.k.Get(s.ctx, s.accounts["1"])
	s.NoError(err)
	s.Equal(referral.StatusHero, data.Status)
	s.Equal(int64(-1), data.StatusDowngradeAt)
}

func (s Suite) TestBanishment_Undelegation() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"]
	parent := app.DefaultGenesisUsers["user1"]

	s.ctx = s.ctx.WithBlockHeight(8999).WithBlockTime(genesisTime.Add(8999 * 30 * time.Second))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.NotZero(len(info.Referrals))
	s.Equal(types.Leader, info.Status)

	s.ctx = s.ctx.WithBlockHeight(8999 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.Equal(types.Lucky, info.Status)

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.False(info.Banished)
	s.Equal(types.Lucky, info.Status)
	s.Equal(parent, info.Referrer)
	s.Zero(info.BanishmentAt)

	s.NoError(s.dk.Revoke(s.ctx, user, sdk.NewInt(10_000_000000)))

	info, err = s.get(user)
	s.NoError(err)
	s.NotNil(info.BanishmentAt)

	s.ctx = s.ctx.WithBlockHeight(8999 + 4*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second + 4*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.True(info.Banished)
	s.Equal(types.Status(0), info.Status)
	s.Equal(parent, info.Referrer)
}

func (s Suite) TestBanishment_DelegationAfterCompression() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"]
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(8999).WithBlockTime(genesisTime.Add(8999 * 30 * time.Second))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.NotZero(len(info.Referrals))
	s.Equal(types.Leader, info.Status)

	s.ctx = s.ctx.WithBlockHeight(8999 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.Equal(types.Lucky, info.Status)
	s.NotNil(info.BanishmentAt)

	s.ctx = s.ctx.WithBlockHeight(8999 + 2*util.BlocksOneMonth + util.BlocksOneDay).WithBlockTime(genesisTime.Add(8999*30*time.Second + 2*30*24*time.Hour + 24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.NoError(s.dk.Delegate(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(1_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(8999*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.False(info.Banished)
	s.Equal(types.Lucky, info.Status)
	s.Equal(parent, info.Referrer)
	s.Zero(info.BanishmentAt)
}

func (s Suite) TestComeBack() {
	user := app.DefaultGenesisUsers["user2"]
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, user, sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(8999)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(8999 + 2*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.True(info.Banished)

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth + util.BlocksOneDay)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.NoError(s.pk.PayForSubscription(s.ctx, app.DefaultGenesisUsers["user2"], 5*util.GBSize))

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Banished)
	s.Zero(info.BanishmentAt)
	s.Equal(parent, info.Referrer)
	s.True(info.Active)

	info, err = s.get(parent)
	s.NoError(err)
	s.Contains(info.Referrals, user)
}

func (s Suite) TestComeBack_BubbleUp() {
	var (
		user1 = app.DefaultGenesisUsers["user1"]
		user2 = app.DefaultGenesisUsers["user2"]
		user4 = app.DefaultGenesisUsers["user4"]
		user8 = app.DefaultGenesisUsers["user8"]
	)

	s.ctx = s.ctx.WithBlockHeight(8999)
	s.NoError(s.pk.PayForSubscription(s.ctx, user1, 5*util.GBSize))
	s.NoError(s.pk.PayForSubscription(s.ctx, user2, 5*util.GBSize))
	s.NoError(s.pk.PayForSubscription(s.ctx, user4, 5*util.GBSize))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(8999 + util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, user1, 5*util.GBSize))
	s.NoError(s.pk.PayForSubscription(s.ctx, user4, 5*util.GBSize))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(8999 + 2*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, user1, 5*util.GBSize))
	s.nextBlock()

	info, err := s.get(user8)
	s.NoError(err)
	s.False(info.Banished)
	s.Equal(user4.String(), info.Referrer.String())

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, user1, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user8)
	s.NoError(err)
	s.True(info.Banished)
	s.Equal(user4.String(), info.Referrer.String())

	info, err = s.get(user4)
	s.NoError(err)
	s.False(info.Active)
	s.True(info.RegistrationClosed(s.ctx))

	s.ctx = s.ctx.WithBlockHeight(8999 + 4*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, user1, 5*util.GBSize))
	s.NoError(s.pk.PayForSubscription(s.ctx, user8, 5*util.GBSize))
	s.nextBlock()

	info, err = s.get(user4)
	s.NoError(err)
	s.False(info.Active)
	s.True(info.RegistrationClosed(s.ctx))

	info, err = s.get(user8)
	s.NoError(err)
	s.False(info.Banished)
	s.Equal(user1, info.Referrer)

	info, err = s.get(user1)
	s.NoError(err)
	s.Contains(info.Referrals, user8)

	info, err = s.get(user2)
	s.NoError(err)
	s.Zero(len(info.Referrals))

	info, err = s.get(user4)
	s.NoError(err)
	s.Zero(len(info.Referrals))
}

func (s Suite) TestComeBackViaDelegation() {
	user := app.DefaultGenesisUsers["user2"]
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, user, sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(8999)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(8999 + 2*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.True(info.Banished)

	s.ctx = s.ctx.WithBlockHeight(8999 + 3*util.BlocksOneMonth + util.BlocksOneDay)
	s.NoError(s.pk.PayForSubscription(s.ctx, parent, 5*util.GBSize))
	s.nextBlock()

	s.NoError(s.dk.Delegate(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(25_000000)))

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Banished)
	s.Zero(info.BanishmentAt)
	s.Equal(parent, info.Referrer)
	s.False(info.Active)

	info, err = s.get(parent)
	s.NoError(err)
	s.Contains(info.Referrals, user)
}

type StatusUpgradeSuite struct {
	BaseSuite
	heads [3]sdk.AccAddress
}

func (s *StatusUpgradeSuite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()

	data, err := ioutil.ReadFile("test-genesis-status-upgrade.json")
	if err != nil {
		panic(err)
	}
	s.setupTest(json.RawMessage(data))

	s.heads[0] = accAddr("artr1847dh25pq47cysckpa05lh7yt7ckuqs8r6gsgu")
	s.heads[1] = accAddr("artr1ykm27k4ju4whlmre776s9s55gjscvvrfzy9ejx")
	s.heads[2] = accAddr("artr1utkd2et99v496k973qvpgn7w6d6zr83feclmnc")
}

func (s *StatusUpgradeSuite) TestStatusUpgradeDowngrade() {
	root := app.DefaultGenesisUsers["user15"]

	var (
		status types.Status
		err    error
		data   types.R
	)

	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusChampion, status)

	// Jump to next level
	s.NoError(s.app.GetBankKeeper().SetCoins(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(150_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusBusinessman, status)

	// Jump several levels at once
	s.NoError(s.app.GetBankKeeper().SetCoins(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2_000_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusHero, status)

	// Step back
	s.NoError(s.app.GetBankKeeper().SetCoins(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusHero, status)

	data, err = s.get(root)
	s.NoError(err)
	s.Equal(referral.StatusHero, data.Status)
	s.Equal(int64(1+referral.StatusDowngradeAfter), data.StatusDowngradeAt)

	// Jump to the top (downgrade should be cancelled)
	s.NoError(s.app.GetBankKeeper().SetCoins(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(100_000_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, status)
	data, err = s.get(root)
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, data.Status)
	s.Equal(int64(-1), data.StatusDowngradeAt)

	// Loose one of teams - should fall to the bottom
	s.NoError(s.k.SetActive(s.ctx, s.heads[2], false))
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, status)
	data, err = s.get(root)
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, data.Status)
	s.Equal(int64(1+referral.StatusDowngradeAfter), data.StatusDowngradeAt)

	// One month later ...
	s.ctx = s.ctx.WithBlockHeight(referral.StatusDowngradeAfter)
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusHero, status)
	data, err = s.get(root)
	s.NoError(err)
	s.Equal(referral.StatusHero, data.Status)
	s.Equal(int64(1+2*referral.StatusDowngradeAfter), data.StatusDowngradeAt)
}

type Status3x3Suite struct {
	BaseSuite
}

func (s *Status3x3Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	data, err := ioutil.ReadFile("test-genesis-status-3x3.json")
	if err != nil {
		panic(err)
	}
	s.setupTest(json.RawMessage(data))
}

func (s *Status3x3Suite) TestStatusDowngrade_3x3() {
	var (
		root, _   = sdk.AccAddressFromBech32("artr1yhy6d3m4utltdml7w7zte7mqx5wyuskq9rr5vg")
		neck00, _ = sdk.AccAddressFromBech32("artr18mrcvv6qkmkx5uyjxy4lpl5fh7w08wgf2acuwt")
		neck02, _ = sdk.AccAddressFromBech32("artr1d8gc7e2mftlcgjgejtluw9uqem88jzj4yydxnw")
	)

	var (
		status types.Status
		err    error
		check  types.StatusCheckResult
	)

	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusChampion, status)

	check, err = s.k.AreStatusRequirementsFulfilled(s.ctx, root, referral.StatusMaster)
	s.NoError(err)
	s.True(check.Overall)
	check, err = s.k.AreStatusRequirementsFulfilled(s.ctx, root, referral.StatusChampion)
	s.NoError(err)
	s.True(check.Overall)

	s.NoError(s.k.RequestTransition(s.ctx, neck00, neck02))
	s.NoError(s.k.AffirmTransition(s.ctx, neck00))
	check, err = s.k.AreStatusRequirementsFulfilled(s.ctx, root, referral.StatusMaster)
	s.NoError(err)
	s.False(check.Overall)
	check, err = s.k.AreStatusRequirementsFulfilled(s.ctx, root, referral.StatusChampion)
	s.NoError(err)
	s.False(check.Overall)

	// One month later
	s.ctx = s.ctx.WithBlockHeight(86400)
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusMaster, status)

	// Two months later
	s.ctx = s.ctx.WithBlockHeight(172800)
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusLeader, status)
}

type StatusBonusSuite struct {
	BaseSuite

	bk bank.Keeper
}

func (s *StatusBonusSuite) SetupTest() {
	data, err := ioutil.ReadFile("test-genesis-status-bonus.json")
	if err != nil {
		panic(err)
	}
	s.setupTest(json.RawMessage(data))
	s.bk = s.app.GetBankKeeper()
}

func (s *StatusBonusSuite) TestStatusBonus() {
	lvl5 := app.DefaultGenesisUsers["user14"]
	lvl7 := app.DefaultGenesisUsers["user15"]
	topR := s.k.GetParams(s.ctx).CompanyAccounts.TopReferrer
	s.nextBlock()

	var (
		status types.Status
		err    error
	)
	status, err = s.k.GetStatus(s.ctx, lvl7)
	s.NoError(err)
	s.Equal(referral.StatusTopLeader, status)

	status, err = s.k.GetStatus(s.ctx, lvl5)
	s.NoError(err)
	s.Equal(referral.StatusBusinessman, status)

	err = s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user1"], 5*util.GBSize)
	s.NoError(err)
	course, price, _, _, _, _ := s.app.GetSubscriptionKeeper().GetPrices(s.ctx)
	payment := int64(course * price)
	total := util.Percent(5).MulInt64(payment - util.CalculateFee(sdk.NewInt(payment)).Int64()).Int64()

	s.Equal(total,
		s.bk.GetCoins(s.ctx, s.k.GetParams(s.ctx).CompanyAccounts.StatusBonuses).AmountOf(util.ConfigMainDenom).Int64(),
		"commission from subscription",
	)

	toLevel5 := total / 10
	toLevel7 := total/5*2 + total/10
	toTopRef := total / 5 * 2

	b0level5 := s.bk.GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	b0level7 := s.bk.GetCoins(s.ctx, lvl7).AmountOf(util.ConfigMainDenom).Int64()
	b0topRef := s.bk.GetCoins(s.ctx, topR).AmountOf(util.ConfigMainDenom).Int64()

	// On the week end
	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneWeek - 1)
	s.nextBlock()

	b1level5 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	b1level7 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl7).AmountOf(util.ConfigMainDenom).Int64()
	b1topRef := s.app.GetBankKeeper().GetCoins(s.ctx, topR).AmountOf(util.ConfigMainDenom).Int64()

	s.Equal(b0level5+toLevel5, b1level5, "Level 5: %d + %d", b0level5, toLevel5)
	s.Equal(b0level7+toLevel7, b1level7, "Level 7: %d + %d", b0level7, toLevel7)
	s.Equal(b0topRef+toTopRef, b1topRef, "Top referrer: %d + %d", b0topRef, toTopRef)
}

func (s *StatusBonusSuite) TestStatusBonus_AfterDowngrade() {
	lvl5 := app.DefaultGenesisUsers["user14"]
	somebody := app.DefaultGenesisUsers["user1"]
	s.nextBlock()

	status, err := s.k.GetStatus(s.ctx, lvl5)
	s.NoError(err)
	s.Equal(referral.StatusBusinessman, status)
	b0level5 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, somebody, 5*util.GBSize))
	total := s.bk.GetCoins(s.ctx, s.k.GetParams(s.ctx).CompanyAccounts.StatusBonuses).AmountOf(util.ConfigMainDenom).Int64()
	s.NotZero(total)
	toLevel5 := total / 10

	// On the week end
	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneWeek - 1)
	s.nextBlock()
	b1level5 := s.bk.GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	s.Equal(b0level5+toLevel5, b1level5, "Week 1: %d + %d", b0level5, toLevel5)

	// Fail status requirements
	s.NoError(s.bk.SendCoins(s.ctx, lvl5, somebody, util.Uartrs(100_000_000000)))
	check, err := s.k.AreStatusRequirementsFulfilled(s.ctx, lvl5, referral.StatusBusinessman)
	s.NoError(err)
	s.False(check.Overall)
	s.nextBlock()
	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, somebody, 5*util.GBSize))
	total = s.bk.GetCoins(s.ctx, s.k.GetParams(s.ctx).CompanyAccounts.StatusBonuses).AmountOf(util.ConfigMainDenom).Int64()
	s.NotZero(total)
	toLevel5 = total / 10

	s.ctx = s.ctx.WithBlockHeight(2*util.BlocksOneWeek - 1)
	s.nextBlock()
	b2level5 := s.bk.GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	s.Equal(b1level5-100_000_000000+toLevel5, b2level5, "Week 2: %d + %d", b1level5-100_000_000000, toLevel5)

	// One month later
	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneWeek + util.BlocksOneMonth - 1)
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, lvl5)
	s.NoError(err)
	s.Equal(referral.StatusChampion, status)

	s.NoError(s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, somebody, 5*util.GBSize))
	total = s.bk.GetCoins(s.ctx, s.k.GetParams(s.ctx).CompanyAccounts.StatusBonuses).AmountOf(util.ConfigMainDenom).Int64()
	s.NotZero(total)
	s.ctx = s.ctx.WithBlockHeight(5*util.BlocksOneWeek - 1)
	s.nextBlock()
	b5level5 := s.bk.GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	s.Equal(b2level5, b5level5, "Week 5: %d + 0", b2level5)
}

// ----- private functions ------------

func (s *BaseSuite) setBalance(acc sdk.AccAddress, coins sdk.Coins) error {
	item := s.accKeeper.GetAccount(s.ctx, acc)
	if item == nil {
		item = s.accKeeper.NewAccountWithAddress(s.ctx, acc)
	}
	err := item.SetCoins(coins)
	if err != nil {
		return err
	}
	s.accKeeper.SetAccount(s.ctx, item)
	return nil
}

func (s *BaseSuite) get(acc sdk.AccAddress) (types.R, error) {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.R
	err := s.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &value)
	return value, err
}

func (s *BaseSuite) set(acc sdk.AccAddress, value types.R) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := s.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (s *BaseSuite) update(acc sdk.AccAddress, callback func(*types.R)) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.R
	err := s.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &value)
	if err != nil {
		return err
	}
	callback(&value)
	valueBytes, err = s.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

var bbHeader = abci.RequestBeginBlock{
	Header: abci.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}

func (s *BaseSuite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}

func accAddr(s string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return addr
}
