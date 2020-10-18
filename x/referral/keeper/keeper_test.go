// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
)

func TestReferralKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()

	cdc       *codec.Codec
	ctx       sdk.Context
	k         referral.Keeper
	storeKey  sdk.StoreKey
	accKeeper auth.AccountKeeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)

	s.cdc       = s.app.Codec()
	s.ctx       = s.app.NewContext(true, abci.Header{Height: 1})
	s.k         = s.app.GetReferralKeeper()
	s.storeKey  = s.app.GetKeys()[referral.ModuleName]
	s.accKeeper = s.app.GetAccountKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

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
		{ 0x0001, 0x0002, 0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400 },
		{ 0x0002, 0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0008, 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0010, 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0020, 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0040, 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0080, 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0100, 0x0200, 0x0400, 0x0800 },
		{ 0x0200, 0x0400, 0x0800 },
		{ 0x0400, 0x0800 },
		{ 0x0800 },
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
				[]sdk.AccAddress{ accounts[i+1] },
				value.Referrals,
				"GetChildren #%d", i,
			)
		}

		expectedRefCount := [11]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		for j := 10; j > 11 - i; j-- {
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
			Status:   types.Hero,
			Referrer: parent,
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
					Amount: sdk.NewInt(1 << (2*i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i+1)),
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

	res, err := s.k.GetCoinsInNetwork(s.ctx, accounts[0])
	s.Nil(err, "GetCoinsInNetwork")
	s.Equal(uint64(0x00CFF3FF), res.Uint64(), "GetCoinsInNetwork")

	res, err = s.k.GetDelegatedInNetwork(s.ctx, accounts[0])
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
	s.Nil(
		s.set(accounts[0], types.R{
			Status:    types.Lucky,
			Coins:     [11]sdk.Int{sdk.NewInt(1)},
			Delegated: [11]sdk.Int{},
		}),
	)
	for i := 0; i < 12-1; i++ {
		s.Nil(s.k.AppendChild(s.ctx, accounts[i], accounts[i+1]))
	}
	for i := 0; i < 12; i++ {
		s.Nil(s.k.SetActive(s.ctx, accounts[i], true))
	}

	var companyAccs types.CompanyAccounts
	s.app.GetSubspaces()[referral.DefaultParamspace].Get(s.ctx, types.KeyCompanyAccounts, &companyAccs)

	res, err := s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating all newbies: no error")
	s.Equal(4, len(res), "GetReferralFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio: util.Percent(5),
	}, "GetReferralFesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all newbies: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio: util.Permille(5),
	}, "GetReferralFesForDelegating all newbies: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio: util.Permille(85),
	}, "GetReferralFesForDelegating all newbies: \"top referrer\"")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription all newbies: no error")
	s.Equal(7, len(res), "GetReferralFeesForSubscription all newbies: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio: util.Percent(15),
	}, "GetReferralFeesForSubscription all newbies: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio: util.Percent(10),
	}, "GetReferralFeesForSubscription all newbies: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio: util.Percent(10),
	}, "GetReferralFeesForSubscription all newbies: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio: util.Percent(44),
	}, "GetReferralFeesForSubscription all newbies: \"top referrer\"")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription all newbies: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription all newbies: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio: util.Percent(5),
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
		Ratio: util.Percent(5),
	}, "GetReferralFesForDelegating all pros: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[8],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 3")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[7],
		Ratio: util.Percent(2),
	}, "GetReferralFesForDelegating all pros: lvl 4")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[6],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 5")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[5],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 6")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[4],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 7")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[3],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 8")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[2],
		Ratio: util.Percent(1),
	}, "GetReferralFesForDelegating all pros: lvl 9")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[1],
		Ratio: util.Permille(5),
	}, "GetReferralFesForDelegating all pros: lvl 10")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio: util.Permille(5),
	}, "GetReferralFesForDelegating all pros: company")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription all pros: no error")
	s.Equal(14, len(res), "GetReferralFeesForSubscription all pros: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio: util.Percent(15),
	}, "GetReferralFeesForSubscription all pros: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[9],
		Ratio: util.Percent(10),
	}, "GetReferralFeesForSubscription all pros: lvl 2")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[8],
		Ratio: util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 3")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[7],
		Ratio: util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 4")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[6],
		Ratio: util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 5")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[5],
		Ratio: util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 6")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[4],
		Ratio: util.Percent(7),
	}, "GetReferralFeesForSubscription all pros: lvl 7")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[3],
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: lvl 8")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[2],
		Ratio: util.Percent(2),
	}, "GetReferralFeesForSubscription all pros: lvl 9")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[1],
		Ratio: util.Percent(2),
	}, "GetReferralFeesForSubscription all pros: lvl 10")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio: util.Percent(10),
	}, "GetReferralFeesForSubscription all pros: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription all pros: status bonus")

	s.Nil(s.update(accounts[10], func(value *types.R) {
		value.Referrer = nil
	}))

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating short chain: no error")
	s.Equal(3, len(res), "GetReferralFeesForDelegating short chain: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio: util.Percent(5),
	}, "GetReferralFesForDelegating short chain: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio: util.Permille(5),
	}, "GetReferralFesForDelegating short chain: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio: util.Permille(95),
	}, "GetReferralFesForDelegating short chain: \"top referrer\"")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription short chain: no error")
	s.Equal(6, len(res), "GetReferralFeesForSubscription short chain: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio: util.Percent(15),
	}, "GetReferralFeesForSubscription short chain: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio: util.Percent(10),
	}, "GetReferralFeesForSubscription short chain: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio: util.Percent(54),
	}, "GetReferralFeesForSubscription short chain: \"top referrer\"")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription short chain: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription short chain: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription short chain: status bonus")

	s.Nil(s.update(accounts[11], func(value *types.R) {
		value.Referrer = nil
	}))

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForDelegating top account: no error")
	s.Equal(2, len(res), "GetReferralFeesForDelegating top account: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForDelegating,
		Ratio: util.Permille(5),
	}, "GetReferralFesForDelegating top account: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio: util.Permille(145),
	}, "GetReferralFesForDelegating top account: \"top referrer\"")

	res, err = s.k.GetReferralFeesForSubscription(s.ctx, accounts[11])
	s.Nil(err, "GetReferralFeesForSubscription top account: no error")
	s.Equal(5, len(res), "GetReferralFeesForSubscription top account: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio: util.Percent(10),
	}, "GetReferralFeesForSubscription top account: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio: util.Percent(69),
	}, "GetReferralFeesForSubscription top account: \"top referrer\"")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.PromoBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription top account: promo bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.LeaderBonuses,
		Ratio: util.Percent(5),
	}, "GetReferralFeesForSubscription top account: leader bonus")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.StatusBonuses,
		Ratio: util.Percent(5),
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
					Amount: sdk.NewInt(1 << (2*i)),
				},
				sdk.Coin{
					Denom:  util.ConfigDelegatedDenom,
					Amount: sdk.NewInt(1 << (2*i+1)),
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
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             nil,
			Referrals:            []sdk.AccAddress{
				accounts[1],
				accounts[9],
			},
			ActiveReferralsCount: [11]int{1, 2, 3, 3},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x000003),
				sdk.NewInt(0x0C000C),
				sdk.NewInt(0x030F30),
				sdk.NewInt(0x00F0C0),
				zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x000002),
				sdk.NewInt(0x080008),
				sdk.NewInt(0x020A20),
				sdk.NewInt(0x00A080),
				zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #1
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[0],
			Referrals:            []sdk.AccAddress{
				accounts[2],
				accounts[4],
				accounts[5],
				accounts[8],
			},
			ActiveReferralsCount: [11]int{1, 3, 3},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x00000C),
				sdk.NewInt(0x030F30),
				sdk.NewInt(0x00F0C0),
				zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x000008),
				sdk.NewInt(0x020A20),
				sdk.NewInt(0x00A080),
				zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #2
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[1],
			Referrals:            []sdk.AccAddress{
				accounts[3],
			},
			ActiveReferralsCount: [11]int{1, 1},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x000030),
				sdk.NewInt(0x0000C0),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x000020),
				sdk.NewInt(0x000080),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #3
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[2],
			ActiveReferralsCount: [11]int{1},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x0000C0),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x000080),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #4
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               false,
			Referrer:             accounts[1],
			ActiveReferralsCount: [11]int{},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x000300),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x000200),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        1 + referral.CompressionPeriod,
		},
		{ // item #5
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[1],
			Referrals:            []sdk.AccAddress{
				accounts[6],
				accounts[7],
			},
			ActiveReferralsCount: [11]int{1, 2},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x000C00),
				sdk.NewInt(0x00F000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x000800),
				sdk.NewInt(0x00A000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #6
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[5],
			ActiveReferralsCount: [11]int{1},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x003000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x002000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #7
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[5],
			ActiveReferralsCount: [11]int{1},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x00C000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x008000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #8
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[1],
			ActiveReferralsCount: [11]int{1},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x030000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x020000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
		{ // item #9
			Status:               types.Lucky,
			StatusDowngradeAt:    -1,
			Active:               true,
			Referrer:             accounts[0],
			ActiveReferralsCount: [11]int{1},
			Coins:                [11]sdk.Int{
				sdk.NewInt(0x0C0000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated:            [11]sdk.Int{
				sdk.NewInt(0x080000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			CompressionAt:        -1,
		},
	} {
		value, err := s.get(accounts[i])
		s.Nilf(err, "get item #%d without error", i)
		s.Equalf(expected, value, "value of item #%d", i)
	}
}

func (s *Suite) TestStatusDowngrade() {
	if err := s.k.Compress(s.ctx, app.DefaultGenesisUsers["user4"]); err != nil { panic(err) }
	// After that, user2 does not fulfill level2 requirements anymore

	addr := app.DefaultGenesisUsers["user2"]
	if r, err := s.get(addr); err != nil { panic(err) } else {
		s.Equal(referral.StatusLeader, r.Status)
		s.Equal(int64(86401), r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil { panic(err) } else {
		s.Equal(referral.StatusLeader, status)
	}

	// Next block (nothing should happen) ...
	s.nextBlock()
	if r, err := s.get(addr); err != nil { panic(err) } else {
		s.Equal(referral.StatusLeader, r.Status)
		s.Equal(int64(86401), r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil { panic(err) } else {
		s.Equal(referral.StatusLeader, status)
	}

	// One month later
	s.ctx = s.ctx.WithBlockHeight(86400)
	s.nextBlock()
	if r, err := s.get(addr); err != nil { panic(err) } else {
		s.Equal(referral.StatusLucky, r.Status)
		s.Equal(int64(-1), r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil { panic(err) } else {
		s.Equal(referral.StatusLucky, status)
	}
}

func (s *Suite) TestStatusUpgradeDowngrade() {
	root := app.DefaultGenesisUsers["user15"]
	var heads [3]sdk.AccAddress
	for i := 0; i < 3; i++ {
		_, _, heads[i] = authtypes.KeyTestPubAddr()
		if err := s.setBalance(heads[i], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1)))); err != nil { panic(err) }
		if err := s.k.AppendChild(s.ctx, root, heads[i]); err != nil { panic(err) }
		if err := s.k.SetActive(s.ctx, heads[i], true); err != nil { panic(err) }
		for j := 0; j < 2_000; j++ {
			_, _, tail := authtypes.KeyTestPubAddr()
			if err := s.setBalance(tail, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1)))); err != nil { panic(err) }
			if err := s.k.AppendChild(s.ctx, heads[i], tail); err != nil { panic(err) }
			if err := s.k.SetActive(s.ctx, tail, true); err != nil { panic(err) }
		}
	}

	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusChampion, status)
	}

	// Jump to next level
	if err := s.app.GetBankKeeper().SetCoins(s.ctx, heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(150_000_000000)))); err != nil { panic(err) }
	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusBusinessman, status)
	}

	// Jump several levels at once
	if err := s.app.GetBankKeeper().SetCoins(s.ctx, heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2_000_000_000000)))); err != nil { panic(err) }
	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusHero, status)
	}

	// Step back
	if err := s.app.GetBankKeeper().SetCoins(s.ctx, heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000_000000)))); err != nil { panic(err) }
	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusHero, status)
	}
	if data, err := s.get(root); err != nil { panic(err) } else {
		s.Equal(referral.StatusHero, data.Status)
		s.Equal(int64(1+referral.StatusDowngradeAfter), data.StatusDowngradeAt)
	}

	// Jump to the top (downgrade should be cancelled)
	if err := s.app.GetBankKeeper().SetCoins(s.ctx, heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(100_000_000_000000)))); err != nil { panic(err) }
	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusAbsoluteChampion, status)
	}
	if data, err := s.get(root); err != nil { panic(err) } else {
		s.Equal(referral.StatusAbsoluteChampion, data.Status)
		s.Equal(int64(-1), data.StatusDowngradeAt)
	}

	// Loose one of teams - should fall to the bottom
	if err := s.k.SetActive(s.ctx, heads[2], false); err != nil { panic(err) }
	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusAbsoluteChampion, status)
	}
	if data, err := s.get(root); err != nil { panic(err) } else {
		s.Equal(referral.StatusAbsoluteChampion, data.Status)
		s.Equal(int64(1+referral.StatusDowngradeAfter), data.StatusDowngradeAt)
	}

	// One month later ...
	s.ctx = s.ctx.WithBlockHeight(referral.StatusDowngradeAfter)
	s.nextBlock()
	if status, err := s.k.GetStatus(s.ctx, root); err != nil { panic(err) } else {
		s.Equal(referral.StatusHero, status)
	}
	if data, err := s.get(root); err != nil { panic(err) } else {
		s.Equal(referral.StatusHero, data.Status)
		s.Equal(int64(1+ 2* referral.StatusDowngradeAfter), data.StatusDowngradeAt)
	}
}

func (s *Suite) TestStatusBonus() {
	lvl5 := app.DefaultGenesisUsers["user14"]
	lvl7 := app.DefaultGenesisUsers["user15"]
	topR := s.k.GetParams(s.ctx).CompanyAccounts.TopReferrer

	for _, root := range []sdk.AccAddress{lvl5, lvl7} {
		for i := 0; i < 3; i++ {
			_, _, head := authtypes.KeyTestPubAddr()
			if err := s.setBalance(head, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1)))); err != nil { panic(err) }
			if err := s.k.AppendChild(s.ctx, root, head); err != nil { panic(err) }
			if err := s.k.SetActive(s.ctx, head, true); err != nil { panic(err) }
			for j := 0; j < 2_000; j++ {
				_, _, tail := authtypes.KeyTestPubAddr()
				if err := s.setBalance(tail, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1)))); err != nil { panic(err) }
				if err := s.k.AppendChild(s.ctx, head, tail); err != nil { panic(err) }
				if err := s.k.SetActive(s.ctx, tail, true); err != nil { panic(err) }
			}
		}
	}
	if err := s.app.GetBankKeeper().SetCoins(s.ctx, lvl5, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(150_000_000000)))); err != nil { panic(err) }
	if err := s.app.GetBankKeeper().SetCoins(s.ctx, lvl7, sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000_000000)))); err != nil { panic(err) }

	if status, err := s.k.GetStatus(s.ctx, lvl5); err != nil { panic(err) } else {
		s.Equal(referral.StatusBusinessman, status)
	}
	if status, err := s.k.GetStatus(s.ctx, lvl7); err != nil { panic(err) } else {
		s.Equal(referral.StatusTopLeader, status)
	}

	if err := s.app.GetSubscriptionKeeper().PayForSubscription(s.ctx, app.DefaultGenesisUsers["user1"], 5 * util.GBSize); err != nil { panic(err) }
	course, price, _, _, _, _ := s.app.GetSubscriptionKeeper().GetPrices(s.ctx)
	payment := int64(course * price)
	total := util.Percent(5).MulInt64(payment - util.CalculateFee(sdk.NewInt(payment)).Int64()).Int64()

	s.Equal(total,
		s.app.GetBankKeeper().GetCoins(s.ctx, s.k.GetParams(s.ctx).CompanyAccounts.StatusBonuses).AmountOf(util.ConfigMainDenom).Int64(),
		"commission from subscription",
	)

	toLevel5 := total / 10
	toLevel7 := total / 5 * 2 + total / 10
	toTopRef := total / 5 * 2

	b0level5 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	b0level7 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl7).AmountOf(util.ConfigMainDenom).Int64()
	b0topRef := s.app.GetBankKeeper().GetCoins(s.ctx, topR).AmountOf(util.ConfigMainDenom).Int64()

	// On the week end
	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneWeek - 1)
	s.nextBlock()

	b1level5 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl5).AmountOf(util.ConfigMainDenom).Int64()
	b1level7 := s.app.GetBankKeeper().GetCoins(s.ctx, lvl7).AmountOf(util.ConfigMainDenom).Int64()
	b1topRef := s.app.GetBankKeeper().GetCoins(s.ctx, topR).AmountOf(util.ConfigMainDenom).Int64()

	s.Equal(b0level5 + toLevel5, b1level5, "Level 5: %d + %d", b0level5, toLevel5)
	s.Equal(b0level7 + toLevel7, b1level7, "Level 7: %d + %d", b0level7, toLevel7)
	s.Equal(b0topRef + toTopRef, b1topRef, "Top referrer: %d + %d", b0topRef, toTopRef)
}

// ----- private functions ------------

func (s *Suite) setBalance(acc sdk.AccAddress, coins sdk.Coins) error {
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

func (s *Suite) get(acc sdk.AccAddress) (types.R, error) {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.R
	err := s.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &value)
	return value, err
}

func (s *Suite) set(acc sdk.AccAddress, value types.R) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := s.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (s *Suite) update(acc sdk.AccAddress, callback func(*types.R)) error {
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
func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, bbHeader)
	return ebr, bbr
}