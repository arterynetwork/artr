// +build testing

package keeper_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authK "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	delegatingK "github.com/arterynetwork/artr/x/delegating/keeper"
	profileK "github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
)

func TestReferralKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
	suite.Run(t, new(TransitionBorderlineSuite))
	suite.Run(t, new(StatusUpgradeSuite))
	suite.Run(t, new(Status3x3Suite))
}

type BaseSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc      codec.BinaryMarshaler
	ctx      sdk.Context
	k        referral.Keeper
	ak       authK.AccountKeeper
	bk       bank.Keeper
	storeKey sdk.StoreKey
}

func (s *BaseSuite) setupTest(genesis json.RawMessage) {
	defer func() {
		if err := recover(); err != nil {
			s.FailNow("panic in setup", err)
		}
	}()

	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(genesis)

	s.cdc = s.app.Codec()
	s.k = s.app.GetReferralKeeper()
	s.ak = s.app.GetAccountKeeper()
	s.bk = s.app.GetBankKeeper()
	s.storeKey = s.app.GetKeys()[referral.ModuleName]
}

func (s *BaseSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

type Suite struct {
	BaseSuite

	pk profileK.Keeper
	dk delegatingK.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.setupTest(nil)

	s.pk = s.app.GetProfileKeeper()
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
	accounts := [12]string{}
	for i := 0; i < 12; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
		s.NoError(
			s.setBalance(addr, sdk.Coins{sdk.Coin{
				Denom:  util.ConfigMainDenom,
				Amount: sdk.NewInt(1 << i),
			}}),
		)
	}

	s.NoError(s.set(accounts[0], types.NewInfo("", sdk.NewInt(1), sdk.ZeroInt())))
	s.NoError(s.k.SetActive(s.ctx, accounts[0], true, true))

	for i := 0; i <= 10; i++ {
		s.NoError(s.k.AppendChild(s.ctx, accounts[i], accounts[i+1]))
		s.NoError(s.k.SetActive(s.ctx, accounts[i+1], true, true))
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
		s.NoErrorf(err, "Get account #%d", i)
		for j := 0; j <= 10; j++ {
			s.Equalf(
				expected[j], value.Coins[j].Int64(),
				"Coins at lvl #%d for item #%d", j, i)
		}

		if i == 0 {
			s.Equal("", value.Referrer, "GetParent #0")
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
				[]string{accounts[i+1]},
				value.Referrals,
				"GetChildren #%d", i,
			)
		}

		expectedRefCount := []uint64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		for j := 10; j > 11-i; j-- {
			expectedRefCount[j] = 0
		}
		s.Equalf(
			expectedRefCount,
			value.ActiveRefCounts,
			"ActiveReferralsCount #%d", i,
		)
	}
}

func (s *Suite) TestGetters() {
	_, _, acc := testdata.KeyTestPubAddr()
	_, _, parent := testdata.KeyTestPubAddr()
	_, _, child1 := testdata.KeyTestPubAddr()
	_, _, child2 := testdata.KeyTestPubAddr()
	s.NoError(
		s.set(acc.String(), types.Info{
			Status:    types.STATUS_HERO,
			Referrer:  parent.String(),
			Referrals: []string{child1.String(), child2.String()},
			//			Coins:                [11]sdk.Int{},
			//			Delegated:            [11]sdk.Int{},
			//			Active:               false,
			//			ActiveReferralsCount: [11]int{},
		}),
	)

	resultStatus, err := s.k.GetStatus(s.ctx, acc.String())
	s.NoError(err, "GetStatus without error")
	s.Equal(types.STATUS_HERO, resultStatus, "GetStatus")

	resultParent, err := s.k.GetParent(s.ctx, acc.String())
	s.NoError(err, "GetParent without error")
	s.Equal(parent.String(), resultParent, "GetParent")

	resultChildren, err := s.k.GetChildren(s.ctx, acc.String())
	s.NoError(err, "GetChildren without error")
	s.Equal([]string{child1.String(), child2.String()}, resultChildren, "GetChildren")
}

func (s *Suite) TestGetCoinsInNetwork() {
	accounts := [12]string{}
	for i := 0; i < 12; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
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
	s.NoError(s.set(accounts[0], types.Info{
		Status:          types.STATUS_LEADER,
		Active:          true,
		ActiveRefCounts: []uint64{1},
		Coins:           []sdk.Int{sdk.NewInt(3)},
		Delegated:       []sdk.Int{sdk.NewInt(2)},
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
	s.NoError(s.k.AppendChild(s.ctx, accounts[0], accounts[1]))
	s.NoError(s.k.SetActive(s.ctx, accounts[1], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[1], accounts[2]))
	s.NoError(s.k.SetActive(s.ctx, accounts[2], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[2], accounts[3]))
	s.NoError(s.k.SetActive(s.ctx, accounts[3], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[3], accounts[4]))
	s.NoError(s.k.SetActive(s.ctx, accounts[4], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[4], accounts[5]))
	s.NoError(s.k.SetActive(s.ctx, accounts[5], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[3], accounts[6]))
	s.NoError(s.k.SetActive(s.ctx, accounts[6], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[1], accounts[7]))
	s.NoError(s.k.SetActive(s.ctx, accounts[7], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[7], accounts[8]))
	s.NoError(s.k.SetActive(s.ctx, accounts[8], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[8], accounts[9]))
	s.NoError(s.k.SetActive(s.ctx, accounts[9], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[9], accounts[10]))
	s.NoError(s.k.SetActive(s.ctx, accounts[10], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[0], accounts[11]))
	s.NoError(s.k.SetActive(s.ctx, accounts[11], true, true))

	res, err := s.k.GetCoinsInNetwork(s.ctx, accounts[0], 10)
	s.NoError(err, "GetCoinsInNetwork")
	s.Equal(uint64(0x00CFF3FF), res.Uint64(), "GetCoinsInNetwork")

	res, err = s.k.GetDelegatedInNetwork(s.ctx, accounts[0], 10)
	s.NoError(err, "GetDelegatedInNetwork")
	s.Equal(uint64(0x008AA2AA), res.Uint64(), "GetDelegatedInNetwork")
}

func (s *Suite) TestReferralFees() {
	accounts := [12]string{}
	for i := 0; i < 12; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
		s.NoError(
			s.setBalance(addr, sdk.Coins{sdk.Coin{
				Denom:  util.ConfigMainDenom,
				Amount: sdk.NewInt(1),
			}}),
		)
	}
	s.NoError(
		s.set(accounts[0], types.Info{
			Status:    types.STATUS_LUCKY,
			Coins:     []sdk.Int{sdk.NewInt(1)},
			Delegated: []sdk.Int{},
		}),
	)
	s.NoError(s.k.SetActive(s.ctx, accounts[0], true, true))
	for i := 1; i < 12; i++ {
		s.NoError(s.k.AppendChild(s.ctx, accounts[i-1], accounts[i]))
		s.NoError(s.k.SetActive(s.ctx, accounts[i], true, true))
	}

	var companyAccs types.CompanyAccounts
	s.app.GetSubspaces()[referral.DefaultParamspace].Get(s.ctx, types.KeyCompanyAccounts, &companyAccs)

	res, err := s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.NoError(err, "GetReferralFeesForDelegating all newbies: no error")
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
	s.NoError(err, "GetReferralFeesForSubscription all newbies: no error")
	s.Equal(4, len(res), "GetReferralFeesForSubscription all newbies: len")
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
		Ratio:       util.Percent(25),
	}, "GetReferralFeesForSubscription all newbies: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Percent(44),
	}, "GetReferralFeesForSubscription all newbies: \"top referrer\"")

	for i := 0; i < 12; i++ {
		s.NoError(s.update(accounts[i], func(value *types.Info) {
			value.Status = types.STATUS_ABSOLUTE_CHAMPION
		}))
	}

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.NoError(err, "GetReferralFeesForDelegating all pros: no error")
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
	s.NoError(err, "GetReferralFeesForSubscription all pros: no error")
	s.Equal(11, len(res), "GetReferralFeesForSubscription all pros: len")
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
		Ratio:       util.Percent(25),
	}, "GetReferralFeesForSubscription all pros: company")

	s.NoError(s.update(accounts[10], func(value *types.Info) {
		value.Referrer = ""
	}))

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.NoError(err, "GetReferralFeesForDelegating short chain: no error")
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
	s.NoError(err, "GetReferralFeesForSubscription short chain: no error")
	s.Equal(3, len(res), "GetReferralFeesForSubscription short chain: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: accounts[10],
		Ratio:       util.Percent(15),
	}, "GetReferralFeesForSubscription short chain: lvl 1")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio:       util.Percent(25),
	}, "GetReferralFeesForSubscription short chain: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Percent(54),
	}, "GetReferralFeesForSubscription short chain: \"top referrer\"")

	s.NoError(s.update(accounts[11], func(value *types.Info) {
		value.Referrer = ""
	}))

	res, err = s.k.GetReferralFeesForDelegating(s.ctx, accounts[11])
	s.NoError(err, "GetReferralFeesForDelegating top account: no error")
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
	s.NoError(err, "GetReferralFeesForSubscription top account: no error")
	s.Equal(2, len(res), "GetReferralFeesForSubscription top account: len")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.ForSubscription,
		Ratio:       util.Percent(25),
	}, "GetReferralFeesForSubscription top account: company")
	s.Contains(res, types.ReferralFee{
		Beneficiary: companyAccs.TopReferrer,
		Ratio:       util.Percent(69),
	}, "GetReferralFeesForSubscription top account: \"top referrer\"")
}

func (s *Suite) TestCompression() {
	accounts := [10]string{}
	for i := 0; i < 10; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
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
	zero := sdk.ZeroInt()
	s.NoError(s.set(accounts[0], types.Info{
		Status:          types.STATUS_LUCKY,
		Active:          true,
		ActiveRefCounts: []uint64{1},
		Coins: []sdk.Int{
			sdk.NewInt(3),
			zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
		},
		Delegated: []sdk.Int{
			sdk.NewInt(2),
			zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
		},
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
	s.NoError(s.k.AppendChild(s.ctx, accounts[0], accounts[1]))
	s.NoError(s.k.SetActive(s.ctx, accounts[1], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[1], accounts[2]))
	s.NoError(s.k.SetActive(s.ctx, accounts[2], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[2], accounts[3]))
	s.NoError(s.k.SetActive(s.ctx, accounts[3], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[1], accounts[4]))
	s.NoError(s.k.SetActive(s.ctx, accounts[4], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[4], accounts[5]))
	s.NoError(s.k.SetActive(s.ctx, accounts[5], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[5], accounts[6]))
	s.NoError(s.k.SetActive(s.ctx, accounts[6], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[5], accounts[7]))
	s.NoError(s.k.SetActive(s.ctx, accounts[7], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[4], accounts[8]))
	s.NoError(s.k.SetActive(s.ctx, accounts[8], true, true))
	s.NoError(s.k.AppendChild(s.ctx, accounts[0], accounts[9]))
	s.NoError(s.k.SetActive(s.ctx, accounts[9], true, true))

	s.NoError(s.k.SetActive(s.ctx, accounts[4], false, true))
	s.NoError(s.k.Compress(s.ctx, accounts[4]))

	for i, expected := range [10]types.Info{
		{ // item #0
			Status: types.STATUS_LUCKY,
			Active: true,
			Referrals: []string{
				accounts[1],
				accounts[9],
			},
			ActiveRefCounts: []uint64{1, 2, 3, 3, 0, 0, 0, 0, 0, 0, 0},
			ActiveReferrals: []string{
				accounts[1],
				accounts[9],
			},
			Coins: []sdk.Int{
				sdk.NewInt(0x000003),
				sdk.NewInt(0x0C000C),
				sdk.NewInt(0x030F30),
				sdk.NewInt(0x00F0C0),
				zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x000002),
				sdk.NewInt(0x080008),
				sdk.NewInt(0x020A20),
				sdk.NewInt(0x00A080),
				zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #1
			Status:   types.STATUS_LUCKY,
			Active:   true,
			Referrer: accounts[0],
			Referrals: []string{
				accounts[2],
				accounts[4],
				accounts[5],
				accounts[8],
			},
			ActiveRefCounts: []uint64{1, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0},
			ActiveReferrals: []string{
				accounts[2],
				accounts[5],
				accounts[8],
			},
			Coins: []sdk.Int{
				sdk.NewInt(0x00000C),
				sdk.NewInt(0x030F30),
				sdk.NewInt(0x00F0C0),
				zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x000008),
				sdk.NewInt(0x020A20),
				sdk.NewInt(0x00A080),
				zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #2
			Status:   types.STATUS_LUCKY,
			Active:   true,
			Referrer: accounts[1],
			Referrals: []string{
				accounts[3],
			},
			ActiveRefCounts: []uint64{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			ActiveReferrals: []string{
				accounts[3],
			},
			Coins: []sdk.Int{
				sdk.NewInt(0x000030),
				sdk.NewInt(0x0000C0),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x000020),
				sdk.NewInt(0x000080),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #3
			Status:          types.STATUS_LUCKY,
			Active:          true,
			Referrer:        accounts[2],
			ActiveRefCounts: []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Coins: []sdk.Int{
				sdk.NewInt(0x0000C0),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x000080),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #4
			Status:          types.STATUS_LUCKY,
			Active:          false,
			Referrer:        accounts[1],
			ActiveRefCounts: []uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Coins: []sdk.Int{
				sdk.NewInt(0x000300),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x000200),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #5
			Status:   types.STATUS_LUCKY,
			Active:   true,
			Referrer: accounts[1],
			Referrals: []string{
				accounts[6],
				accounts[7],
			},
			ActiveRefCounts: []uint64{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			ActiveReferrals: []string{
				accounts[6],
				accounts[7],
			},
			Coins: []sdk.Int{
				sdk.NewInt(0x000C00),
				sdk.NewInt(0x00F000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x000800),
				sdk.NewInt(0x00A000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #6
			Status:          types.STATUS_LUCKY,
			Active:          true,
			Referrer:        accounts[5],
			ActiveRefCounts: []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Coins: []sdk.Int{
				sdk.NewInt(0x003000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x002000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #7
			Status:          types.STATUS_LUCKY,
			Active:          true,
			Referrer:        accounts[5],
			ActiveRefCounts: []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Coins: []sdk.Int{
				sdk.NewInt(0x00C000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x008000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #8
			Status:          types.STATUS_LUCKY,
			Active:          true,
			Referrer:        accounts[1],
			ActiveRefCounts: []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Coins: []sdk.Int{
				sdk.NewInt(0x030000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x020000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
		{ // item #9
			Status:          types.STATUS_LUCKY,
			Active:          true,
			Referrer:        accounts[0],
			ActiveRefCounts: []uint64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Coins: []sdk.Int{
				sdk.NewInt(0x0C0000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
			Delegated: []sdk.Int{
				sdk.NewInt(0x080000),
				zero, zero, zero, zero, zero, zero, zero, zero, zero, zero,
			},
		},
	} {
		sort.Strings(expected.ActiveReferrals)
		value, err := s.get(accounts[i])
		s.NoErrorf(err, "get item #%d without error", i)
		s.Equalf(expected, value, "value of item #%d", i)
	}
}

func (s *Suite) TestAddChildJustBeforeCompression() {
	user1 := app.DefaultGenesisUsers["user1"].String()
	genesisTime := s.ctx.BlockTime()

	accounts := [3]string{}
	for i := 0; i < 3; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
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
	s.ctx = s.ctx.WithBlockHeight(1 + 9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.nextBlock()
	info, err := s.get(user1)
	s.NoError(err)
	s.False(info.Active)
	s.False(info.RegistrationClosed(s.ctx, s.app.GetScheduleKeeper()))
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[1]))

	// One month later
	s.ctx = s.ctx.WithBlockHeight(9002 + 86400).WithBlockTime(genesisTime.Add(9001*30*time.Second + 30*24*time.Hour))
	s.nextBlock()
	info, err = s.get(user1)
	s.NoError(err)
	s.False(info.Active)
	s.True(info.RegistrationClosed(s.ctx, s.app.GetScheduleKeeper()))
	s.Error(s.k.AppendChild(s.ctx, user1, accounts[2]))
}

func (s *Suite) TestAddChildAfterCompression() {
	user1 := app.DefaultGenesisUsers["user1"].String()
	genesisTime := s.ctx.BlockTime()

	accounts := [2]string{}
	for i := 0; i < 2; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
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
	s.ctx = s.ctx.WithBlockHeight(1 + 9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.nextBlock()
	info, err := s.get(user1)
	s.NoError(err)
	s.False(info.Active)
	s.False(info.RegistrationClosed(s.ctx, s.app.GetScheduleKeeper()))
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[1]))

	// After compression
	s.ctx = s.ctx.WithBlockHeight(9002 + 2*86400).WithBlockTime(genesisTime.Add(9001*30*time.Second + 2*30*24*time.Hour))
	s.nextBlock()
	info, err = s.get(user1)
	s.NoError(err)
	s.Zero(len(info.Referrals))
	s.True(info.RegistrationClosed(s.ctx, s.app.GetScheduleKeeper()))
	s.Error(s.k.AppendChild(s.ctx, user1, accounts[1]))
}

func (s *Suite) TestAddChildAfterReactivation() {
	user1 := app.DefaultGenesisUsers["user1"].String()
	genesisTime := s.ctx.BlockTime()

	accounts := [1]string{}
	for i := 0; i < 1; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		accounts[i] = addr.String()
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
	s.ctx = s.ctx.WithBlockHeight(1 + 9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.nextBlock()
	info, err := s.get(user1)
	s.NoError(err)
	s.False(info.Active)

	// Compression
	s.ctx = s.ctx.WithBlockHeight(9002 + 2*86400).WithBlockTime(genesisTime.Add(9001*30*time.Second + 2*30*24*time.Hour))
	s.nextBlock()
	info, err = s.get(user1)
	s.NoError(err)
	s.Zero(len(info.Referrals))

	// Pay tariff
	s.NoError(s.app.GetProfileKeeper().PayTariff(s.ctx, app.DefaultGenesisUsers["user1"], 5))
	info, err = s.get(user1)
	s.NoError(err)
	s.True(info.Active)
	s.False(info.RegistrationClosed(s.ctx, s.app.GetScheduleKeeper()))
	s.NoError(s.k.AppendChild(s.ctx, user1, accounts[0]))
}

func (s *Suite) TestStatusDowngrade() {
	genesis_time := s.ctx.BlockTime()
	if err := s.k.Compress(s.ctx, app.DefaultGenesisUsers["user4"].String()); err != nil {
		panic(err)
	}
	// After that, user2 does not fulfill level2 requirements anymore

	addr := app.DefaultGenesisUsers["user2"].String()
	if r, err := s.get(addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLeader, r.Status)
		s.NotNil(r.StatusDowngradeAt)
		s.Equal(genesis_time.Add(2*24*time.Hour), *r.StatusDowngradeAt)
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
		s.NotNil(r.StatusDowngradeAt)
		s.Equal(genesis_time.Add(2*24*time.Hour), *r.StatusDowngradeAt)
	}
	if status, err := s.k.GetStatus(s.ctx, addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLeader, status)
	}

	// One month later
	s.ctx = s.ctx.WithBlockHeight(86400).WithBlockTime(genesis_time.Add(2 * 24 * time.Hour))
	s.nextBlock()
	if r, err := s.get(addr); err != nil {
		panic(err)
	} else {
		s.Equal(referral.StatusLucky, r.Status)
		s.Nil(r.StatusDowngradeAt)
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

	s.NoError(s.k.RequestTransition(s.ctx, subj.String(), dest.String()), "request transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.bk.GetBalance(s.ctx, subj),
	)

	s.NoError(s.k.AffirmTransition(s.ctx, subj.String()), "affirm transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.bk.GetBalance(s.ctx, subj),
	)

	acc, err := s.k.GetParent(s.ctx, subj.String())
	s.NoError(err, "get parent")
	s.Equal(dest.String(), acc, "new parent")

	info, err := s.k.Get(s.ctx, oldParent.String())
	s.NoError(err, "get old parent info")
	s.Equal(
		[]string{app.DefaultGenesisUsers["user5"].String()},
		info.Referrals, "old parent's children",
	)
	s.Equal(
		[]string{app.DefaultGenesisUsers["user5"].String()},
		info.ActiveReferrals, "old parent's active children",
	)

	info, err = s.k.Get(s.ctx, dest.String())
	s.NoError(err, "get new parent info")
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user6"].String(),
			app.DefaultGenesisUsers["user7"].String(),
			subj.String(),
		},
		info.Referrals, "new parent's children",
	)
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user6"].String(),
			app.DefaultGenesisUsers["user7"].String(),
			subj.String(),
		},
		info.ActiveReferrals, "new parent's active children",
	)

	info, err = s.k.Get(s.ctx, subj.String())
	s.NoError(err, "get subject info")
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user8"].String(),
			app.DefaultGenesisUsers["user9"].String(),
		},
		info.Referrals, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj.String())
	s.NoError(err, "get pending transition")
	s.Equal("", acc, "pending transition")

	for i, n := range []int64{
		34_990_000000,
		14_000_000000, 19_990_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)].String(), 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

func (s Suite) TestTransition_Decline() {
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.NoError(s.k.RequestTransition(s.ctx, subj.String(), dest.String()), "request transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.bk.GetBalance(s.ctx, subj),
	)

	s.NoError(s.k.CancelTransition(s.ctx, subj.String(), false), "decline transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.bk.GetBalance(s.ctx, subj),
	)

	acc, err := s.k.GetParent(s.ctx, subj.String())
	s.NoError(err, "get parent")
	s.Equal(oldParent.String(), acc, "new parent")

	accz, err := s.k.GetChildren(s.ctx, oldParent.String())
	s.NoError(err, "get old parent's children")
	s.Equal(
		[]string{
			subj.String(),
			app.DefaultGenesisUsers["user5"].String(),
		},
		accz, "old parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, dest.String())
	s.NoError(err, "get new parent's children")
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user6"].String(),
			app.DefaultGenesisUsers["user7"].String(),
		},
		accz, "new parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, subj.String())
	s.NoError(err, "get subject's children")
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user8"].String(),
			app.DefaultGenesisUsers["user9"].String(),
		},
		accz, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj.String())
	s.NoError(err, "get pending transition")
	s.Equal("", acc, "pending transition")

	for i, n := range []int64{
		34_990_000000,
		16_990_000000, 17_000_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)].String(), 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

func (s Suite) TestTransition_Timeout() {
	genesisTime := s.ctx.BlockTime()
	subj := app.DefaultGenesisUsers["user4"]
	dest := app.DefaultGenesisUsers["user3"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.NoError(s.k.RequestTransition(s.ctx, subj.String(), dest.String()), "request transition")
	s.Equal(
		util.Uartrs(990_000000),
		s.bk.GetBalance(s.ctx, subj),
	)

	for i, n := range []sdk.Coins{
		THOUSAND,
		STAKE, STAKE,
		util.Uartrs(990_000000), THOUSAND, THOUSAND, THOUSAND, // transition fee
		THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND,
	} {
		cz := s.bk.GetBalance(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)])
		s.Equal(n, cz)
	}
	for i, n := range []int64{
		34_990_000000,
		16_990_000000, 17_000_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)].String(), 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}

	s.ctx = s.ctx.WithBlockHeight(util.BlocksOneDay).WithBlockTime(genesisTime.Add(24 * time.Hour))
	s.nextBlock()
	for i, n := range []sdk.Coins{
		util.Uartrs(1_010_000000), // validator's award
		STAKE, STAKE,
		util.Uartrs(990_000000), THOUSAND, THOUSAND, THOUSAND,
		THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND, THOUSAND,
	} {
		cz := s.bk.GetBalance(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)])
		s.Equal(n, cz)
	}

	acc, err := s.k.GetParent(s.ctx, subj.String())
	s.NoError(err, "get parent")
	s.Equal(oldParent.String(), acc, "new parent")

	accz, err := s.k.GetChildren(s.ctx, oldParent.String())
	s.NoError(err, "get old parent's children")
	s.Equal(
		[]string{
			subj.String(),
			app.DefaultGenesisUsers["user5"].String(),
		},
		accz, "old parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, dest.String())
	s.NoError(err, "get new parent's children")
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user6"].String(),
			app.DefaultGenesisUsers["user7"].String(),
		},
		accz, "new parent's children",
	)

	accz, err = s.k.GetChildren(s.ctx, subj.String())
	s.NoError(err, "get subject's children")
	s.Equal(
		[]string{
			app.DefaultGenesisUsers["user8"].String(),
			app.DefaultGenesisUsers["user9"].String(),
		},
		accz, "subject's children",
	)

	acc, err = s.k.GetPendingTransition(s.ctx, subj.String())
	s.NoError(err, "get pending transition")
	s.Equal("", acc, "pending transition")

	for i, n := range []int64{
		35_000_000000,
		16_990_000000, 17_000_000000,
		2_990_000000, 3_000_000000, 3_000_000000, 3_000_000000,
		1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000, 1_000_000000,
	} {
		cz, err := s.k.GetCoinsInNetwork(s.ctx, app.DefaultGenesisUsers[fmt.Sprintf("user%d", i+1)].String(), 10)
		s.NoError(err, "get coins of user%d", i+1)
		s.Equal(sdk.NewInt(n), cz, "coins of user%d", i+1)
	}
}

func (s Suite) TestTransition_Validate_Circle() {
	subj := app.DefaultGenesisUsers["user2"]
	dest := app.DefaultGenesisUsers["user5"]

	s.EqualError(
		s.k.RequestTransition(s.ctx, subj.String(), dest.String()),
		"transition is invalid: cycles are not allowed",
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(10_000_000000)),
		),
		s.bk.GetBalance(s.ctx, subj),
	)
}

func (s Suite) TestTransition_Validate_Self() {
	subj := app.DefaultGenesisUsers["user2"]

	s.EqualError(
		s.k.RequestTransition(s.ctx, subj.String(), subj.String()),
		"transition is invalid: subject cannot be their own referral",
	)
	s.Equal(
		sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000000)),
			sdk.NewCoin(util.ConfigDelegatedDenom, sdk.NewInt(10_000_000000)),
		),
		s.bk.GetBalance(s.ctx, subj),
	)
}

func (s Suite) TestTransition_Validate_OldParent() {
	subj := app.DefaultGenesisUsers["user4"]
	oldParent := app.DefaultGenesisUsers["user2"]

	s.EqualError(
		s.k.RequestTransition(s.ctx, subj.String(), oldParent.String()),
		"transition is invalid: destination address is already subject's referrer",
	)
	s.Equal(
		util.Uartrs(1_000_000000),
		s.bk.GetBalance(s.ctx, subj),
	)
}

func (s Suite) TestBanishment() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"].String()
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.NotZero(len(info.Referrals))
	s.Equal(types.STATUS_LEADER, info.Status)

	s.ctx = s.ctx.WithBlockHeight(9000 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.Equal(types.STATUS_LUCKY, info.Status)

	s.ctx = s.ctx.WithBlockHeight(9000 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.True(info.Banished)
	s.Equal(types.STATUS_UNSPECIFIED, info.Status)
	s.Equal(parent.String(), info.Referrer)

	info, err = s.get(parent.String())
	s.NotContains(info.Referrals, user)
	s.NoError(err)
}

func (s Suite) TestBanishment_Undelegation() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"].String()
	parent := app.DefaultGenesisUsers["user1"]

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.NotZero(len(info.Referrals))
	s.Equal(types.STATUS_LEADER, info.Status)

	s.ctx = s.ctx.WithBlockHeight(9000 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.Equal(types.STATUS_LUCKY, info.Status)

	s.ctx = s.ctx.WithBlockHeight(9000 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.False(info.Banished)
	s.Equal(types.STATUS_LUCKY, info.Status)
	s.Equal(parent.String(), info.Referrer)
	s.Nil(info.BanishmentAt)

	s.NoError(s.dk.Revoke(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(10_000_000000)))

	info, err = s.get(user)
	s.NoError(err)
	s.NotNil(info.BanishmentAt)

	s.ctx = s.ctx.WithBlockHeight(9000 + 4*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 4*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.True(info.Banished)
	s.Equal(types.STATUS_UNSPECIFIED, info.Status)
	s.Equal(parent.String(), info.Referrer)
}

func (s Suite) TestBanishment_DelegationAfterCompression() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"].String()
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.NotZero(len(info.Referrals))
	s.Equal(types.STATUS_LEADER, info.Status)

	s.ctx = s.ctx.WithBlockHeight(9000 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.Equal(types.STATUS_LUCKY, info.Status)
	s.NotNil(info.BanishmentAt)

	s.ctx = s.ctx.WithBlockHeight(9000 + 2*util.BlocksOneMonth + util.BlocksOneDay).WithBlockTime(genesisTime.Add(9000*30*time.Second + 2*30*24*time.Hour + 24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.NoError(s.dk.Delegate(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(1_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(9000 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Active)
	s.Zero(len(info.Referrals))
	s.False(info.Banished)
	s.Equal(types.STATUS_LUCKY, info.Status)
	s.Equal(parent.String(), info.Referrer)
	s.Nil(info.BanishmentAt)
}

func (s Suite) TestComeBack() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"].String()
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, app.DefaultGenesisUsers["user2"], sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(9000 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(9000 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err := s.get(user)
	s.NoError(err)
	s.True(info.Banished)

	s.ctx = s.ctx.WithBlockHeight(9000 + 3*util.BlocksOneMonth + util.BlocksOneDay).WithBlockTime(genesisTime.Add(9000*30*time.Second + 3*30*24*time.Hour + 24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.NoError(s.pk.PayTariff(s.ctx, app.DefaultGenesisUsers["user2"], 5))

	info, err = s.get(user)
	s.NoError(err)
	s.False(info.Banished)
	s.Nil(info.BanishmentAt)
	s.Equal(parent.String(), info.Referrer)
	s.True(info.Active)

	info, err = s.get(parent.String())
	s.NoError(err)
	s.Contains(info.Referrals, user)
}

func (s Suite) TestComeBack_BubbleUp() {
	var (
		genesisTime = s.ctx.BlockTime()

		user1 = app.DefaultGenesisUsers["user1"]
		user2 = app.DefaultGenesisUsers["user2"]
		user4 = app.DefaultGenesisUsers["user4"]
		user8 = app.DefaultGenesisUsers["user8"]
	)

	s.ctx = s.ctx.WithBlockHeight(9000).WithBlockTime(genesisTime.Add(9000 * 30 * time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, user1, 5))
	s.NoError(s.pk.PayTariff(s.ctx, user2, 5))
	s.NoError(s.pk.PayTariff(s.ctx, user4, 5))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(9000 + 2*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 2*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, user1, 5))
	s.nextBlock()

	s.ctx = s.ctx.WithBlockHeight(9000 + 3*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 3*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, user1, 5))
	s.nextBlock()

	info, err := s.get(user8.String())
	s.NoError(err)
	s.True(info.Banished)
	s.Equal(user4.String(), info.Referrer)

	s.ctx = s.ctx.WithBlockHeight(9000 + 4*util.BlocksOneMonth).WithBlockTime(genesisTime.Add(9000*30*time.Second + 4*30*24*time.Hour))
	s.NoError(s.pk.PayTariff(s.ctx, user1, 5))
	s.nextBlock()
	s.NoError(s.pk.PayTariff(s.ctx, user8, 5))

	info, err = s.get(user8.String())
	s.NoError(err)
	s.False(info.Banished)
	s.Equal(user1.String(), info.Referrer)

	info, err = s.get(user1.String())
	s.NoError(err)
	s.Contains(info.Referrals, user8.String())

	info, err = s.get(user2.String())
	s.NoError(err)
	s.Zero(len(info.Referrals))

	info, err = s.get(user4.String())
	s.NoError(err)
	s.Zero(len(info.Referrals))
}

type TransitionBorderlineSuite struct {
	BaseSuite

	accounts map[string]string
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

	s.accounts = map[string]string{
		"1":     "artr1qq9gvskgjkwfkqexeapwps0cnqj6pxkz4nevre",
		"1.1":   "artr1qqxwvzmhjsrwa9fuyafu2jcxcrv2fclwrpy33g",
		"1.1.1": "artr1qqvnckqa5yqaps2v9wfeqpzkum4cmexcmr38kj",
		"1.1.2": "artr1pg635yjdpg62pjvsxfz5xyhxcxk2ss4lkepp7x",
		"2":     "artr1sxwwflxyj2wl0l3ltl83kn7sxvrkfalymmhvf0",
		"2.1":   "artr1sxnhvuyuac9x52lmpduyf9uaz763nw0wwdu5qm",
		"2.1.1": "artr1sx48ywhy3yqyhf4h4yxc4n2ucz62xkzva3e7d8",
		"2.1.2": "artr13366fwedzhlu7l66kmrq3utq9x5y0f7f46yzj9",
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
	s.Nil(data.StatusDowngradeAt)
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
	s.Nil(data.StatusDowngradeAt)
}

func (s Suite) TestComeBackViaDelegation() {
	genesisTime := s.ctx.BlockTime()
	user := app.DefaultGenesisUsers["user2"]
	parent := app.DefaultGenesisUsers["user1"]

	s.NoError(s.dk.Revoke(s.ctx, user, sdk.NewInt(10_000_000000)))

	s.ctx = s.ctx.WithBlockHeight(8999).WithBlockTime(genesisTime.Add(8999 * 30*time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.ctx = s.ctx.
		WithBlockHeight(8999 + 2*util.BlocksOneMonth).
		WithBlockTime(genesisTime.Add((8999 + 2*util.BlocksOneMonth) * 30*time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.ctx = s.ctx.
		WithBlockHeight(8999 + 3*util.BlocksOneMonth).
		WithBlockTime(genesisTime.Add((8999 + 3*util.BlocksOneMonth) * 30*time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	info, err := s.get(user.String())
	s.NoError(err)
	s.True(info.Banished)

	s.ctx = s.ctx.
		WithBlockHeight(8999 + 3*util.BlocksOneMonth + util.BlocksOneDay).
		WithBlockTime(genesisTime.Add((8999 + 3*util.BlocksOneMonth + util.BlocksOneDay) * 30*time.Second))
	s.NoError(s.pk.PayTariff(s.ctx, parent, 5))
	s.nextBlock()

	s.NoError(s.dk.Delegate(s.ctx, user, sdk.NewInt(25_000000)))

	info, err = s.get(user.String())
	s.NoError(err)
	s.False(info.Banished)
	s.Nil(info.BanishmentAt)
	s.Equal(parent.String(), info.Referrer)
	s.False(info.Active)

	info, err = s.get(parent.String())
	s.NoError(err)
	s.Contains(info.Referrals, user.String())
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
	genesisTime := s.ctx.BlockTime()
	root := app.DefaultGenesisUsers["user15"]

	var (
		status types.Status
		err    error
		data   types.Info
	)

	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusChampion, status)

	// Jump to next level
	s.NoError(s.bk.SetBalance(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(150_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusBusinessman, status)

	// Jump several levels at once
	s.NoError(s.bk.SetBalance(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(2_000_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusHero, status)

	// Step back
	s.NoError(s.bk.SetBalance(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(1_000_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusHero, status)

	data, err = s.get(root.String())
	s.NoError(err)
	s.Equal(referral.StatusHero, data.Status)
	s.NotNil(data.StatusDowngradeAt)
	s.Equal(genesisTime.Add(2*24*time.Hour), *data.StatusDowngradeAt)

	// Jump to the top (downgrade should be cancelled)
	s.NoError(s.bk.SetBalance(s.ctx, s.heads[0], sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(100_000_000_000000)))))
	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, status)
	data, err = s.get(root.String())
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, data.Status)
	s.Nil(data.StatusDowngradeAt)

	// Loose one of teams - should fall to the bottom
	s.NoError(s.k.SetActive(s.ctx, s.heads[2].String(), false, true))
	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, status)
	data, err = s.get(root.String())
	s.NoError(err)
	s.Equal(referral.StatusAbsoluteChampion, data.Status)
	s.NotNil(data.StatusDowngradeAt)
	s.Equal(genesisTime.Add(2*24*time.Hour), *data.StatusDowngradeAt)

	// One month later ...
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 2*2880 - 1).WithBlockTime(s.ctx.BlockTime().Add(2*24*time.Hour - 30*time.Second))
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, root.String())
	s.NoError(err)
	s.Equal(referral.StatusHero, status)
	data, err = s.get(root.String())
	s.NoError(err)
	s.Equal(referral.StatusHero, data.Status)
	s.NotNil(data.StatusDowngradeAt)
	s.Equal(genesisTime.Add(2*2*24*time.Hour), *data.StatusDowngradeAt)
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
	genesisTime := s.ctx.BlockTime()
	const (
		root   = "artr1yhy6d3m4utltdml7w7zte7mqx5wyuskq9rr5vg"
		neck00 = "artr18mrcvv6qkmkx5uyjxy4lpl5fh7w08wgf2acuwt"
		neck02 = "artr1d8gc7e2mftlcgjgejtluw9uqem88jzj4yydxnw"
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
	s.ctx = s.ctx.WithBlockHeight(86400).WithBlockTime(genesisTime.Add(2 * 24 * time.Hour))
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusMaster, status)

	// Two months later
	s.ctx = s.ctx.WithBlockHeight(172800).WithBlockTime(genesisTime.Add(2 * 2 * 24 * time.Hour))
	s.nextBlock()
	status, err = s.k.GetStatus(s.ctx, root)
	s.NoError(err)
	s.Equal(referral.StatusLeader, status)
}

// ----- private functions ------------

func (s *BaseSuite) setBalance(acc sdk.AccAddress, coins sdk.Coins) error {
	return s.bk.SetBalance(s.ctx, acc, coins)
}

func (s *BaseSuite) get(acc string) (types.Info, error) {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.Info
	err := s.cdc.UnmarshalBinaryBare(valueBytes, &value)
	return value, err
}

func (s *BaseSuite) set(acc string, value types.Info) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := s.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (s *BaseSuite) update(acc string, callback func(*types.Info)) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.Info
	err := s.cdc.UnmarshalBinaryBare(valueBytes, &value)
	if err != nil {
		return errors.Wrap(err, "cannot unmarshal value")
	}
	callback(&value)
	valueBytes, err = s.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return errors.Wrap(err, "cannot marshal value")
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

var bbHeader = abci.RequestBeginBlock{
	Header: tmproto.Header{
		ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
	},
}

func (s *BaseSuite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
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
