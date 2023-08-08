// +build testing

package referral_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	profileK "github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
	scheduleT "github.com/arterynetwork/artr/x/schedule/types"
)

func TestReferralGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         referral.Keeper
	subKeeper profileK.Keeper

	bbHeader abci.RequestBeginBlock
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)
	s.k = s.app.GetReferralKeeper()
	s.subKeeper = s.app.GetProfileKeeper()

	s.bbHeader = abci.RequestBeginBlock{
		Header: tmproto.Header{
			ProposerAddress: sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey).Address().Bytes(),
		},
	}
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestStatusDowngrade() {
	var (
		status referral.Status
		check  referral.StatusCheckResult
		err    error
	)
	var (
		user2 = app.DefaultGenesisUsers["user2"].String()
		user8 = app.DefaultGenesisUsers["user8"]
	)

	p := s.subKeeper.GetProfile(s.ctx, user8)
	*p.ActiveUntil = s.ctx.BlockTime().Add(-time.Hour)
	s.NoError(s.subKeeper.SetProfile(s.ctx, user8, *p))

	// so, user2 loses its status
	if status, err = s.k.GetStatus(s.ctx, user2); err != nil {
		panic(err)
	}
	if check, err = s.k.AreStatusRequirementsFulfilled(s.ctx, user2, status); err != nil {
		panic(err)
	}
	s.False(check.Overall)

	s.checkExportImport()
}

func (s Suite) TestCompression() {
	user1 := app.DefaultGenesisUsers["user1"]

	p := s.subKeeper.GetProfile(s.ctx, user1)
	*p.ActiveUntil = s.ctx.BlockTime().Add(-time.Hour)
	s.NoError(s.subKeeper.SetProfile(s.ctx, user1, *p))

	// so, compression is scheduled
	info, err := s.k.Get(s.ctx, user1.String())
	s.NoError(err)
	s.NotNil(info.CompressionAt)

	s.checkExportImport()
}

func (s Suite) TestAlreadyCompressed() {
	user1 := app.DefaultGenesisUsers["user1"]

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 8640).WithBlockTime(s.ctx.BlockTime().Add(4 * 24 * time.Hour))
	s.nextBlock()
	info, err := s.k.Get(s.ctx, user1.String())
	s.NoError(err)
	s.False(info.Active)
	s.NotNil(info.CompressionAt)

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(*info.CompressionAt)
	s.nextBlock()
	info, err = s.k.Get(s.ctx, user1.String())
	s.NoError(err)
	s.Nil(info.CompressionAt)

	s.checkExportImport()
}

func (s *Suite) TestParams() {
	s.k.SetParams(s.ctx, referral.Params{
		CompanyAccounts: referral.CompanyAccounts{
			TopReferrer:     user(10),
			ForSubscription: user(11),
			ForDelegating:   user(15),
		},
		DelegatingAward: referral.NetworkAward{
			Network: []util.Fraction{
				util.Permille(1),
				util.Permille(3),
				util.Permille(5),
				util.Permille(7),
				util.Permille(9),
				util.Permille(11),
				util.Permille(13),
				util.Permille(15),
				util.Permille(17),
				util.Permille(19),
			},
			Company: util.Percent(21),
		},
		SubscriptionAward: referral.NetworkAward{
			Network: []util.Fraction{
				util.Permille(2),
				util.Permille(4),
				util.Permille(6),
				util.Permille(8),
				util.Permille(10),
				util.Permille(12),
				util.Permille(14),
				util.Permille(16),
				util.Permille(18),
				util.Permille(20),
			},
			Company: util.Percent(22),
		},
	})
	s.checkExportImport()
}

func (s Suite) TestBanished() {
	user := app.DefaultGenesisUsers["user14"].String()

	if err := s.k.SetActive(s.ctx, user, false, true); err != nil {
		panic(err)
	}
	s.NoError(s.k.Compress(s.ctx, user))
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + util.BlocksOneMonth).WithBlockTime(s.ctx.BlockTime().Add(30 * 24 * time.Hour))
	s.nextBlock()
	for i := 1; i <= 7; i++ {
		user := fmt.Sprintf("user%d", i)
		addr := app.DefaultGenesisUsers[user]
		s.NoError(s.subKeeper.PayTariff(s.ctx, addr, 5), "pay tariff for %s (%s)", user, addr.String())
	}

	r, err := s.k.Get(s.ctx, user)
	s.NoError(err)
	s.True(r.Banished)
	s.Zero(r.Status)

	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		s.ctx.BlockTime(),
		[]string{
			referral.StoreKey,
			scheduleT.StoreKey,
			params.StoreKey,
		},
		map[string]app.Decoder{
			referral.StoreKey:  app.StringDecoder,
			scheduleT.StoreKey: app.Uint64Decoder,
			params.StoreKey:    app.DummyDecoder,
		},
		map[string]app.Decoder{
			referral.StoreKey: func(bz []byte) (string, error) {
				var result types.Info
				err := s.app.Codec().UnmarshalBinaryBare(bz, &result)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%+v", result), nil
			},
			scheduleT.StoreKey: app.DummyDecoder,
			params.StoreKey:    app.DummyDecoder,
		},
		make(map[string][][]byte, 0),
	)
}

func user(n int) string {
	return app.DefaultGenesisUsers[fmt.Sprintf("user%d", n)].String()
}

func (s *Suite) nextBlock() (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(30 * time.Second))
	bbr := s.app.BeginBlocker(s.ctx, s.bbHeader)
	return ebr, bbr
}
