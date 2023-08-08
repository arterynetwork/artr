// +build testing

package keeper_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authK "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	delegatingK "github.com/arterynetwork/artr/x/delegating/keeper"
	nodingK "github.com/arterynetwork/artr/x/noding/keeper"
	profileK "github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/referral/types"
)

func TestValidatorAccrueReferralKeeper(t *testing.T) {
	suite.Run(t, new(VASuite))
}

type VASuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc      codec.BinaryMarshaler
	ctx      sdk.Context
	k        referral.Keeper
	ak       authK.AccountKeeper
	bk       bank.Keeper
	storeKey sdk.StoreKey

	pk            profileK.Keeper
	dk            delegatingK.Keeper
	nk            nodingK.Keeper
	indexStoreKey sdk.StoreKey
}

func (s *VASuite) SetupTest() {
	defer func() {
		if err := recover(); err != nil {
			s.FailNow("panic in setup", err)
		}
	}()

	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)

	s.cdc = s.app.Codec()
	s.k = s.app.GetReferralKeeper()
	s.ak = s.app.GetAccountKeeper()
	s.bk = s.app.GetBankKeeper()
	s.storeKey = s.app.GetKeys()[referral.ModuleName]

	s.pk = s.app.GetProfileKeeper()
	s.dk = s.app.GetDelegatingKeeper()
	s.nk = s.app.GetNodingKeeper()
	s.indexStoreKey = s.app.GetKeys()[types.IndexStoreKey]
}

func (s *VASuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

var (
	minIndexedStatus = types.STATUS_BUSINESSMAN
)

func (s *VASuite) TestReferralValidatorFees() {
	pz := s.nk.GetParams(s.ctx)
	s.ctx = s.ctx.WithBlockHeight(6*int64(pz.UnjailAfter) + 1)

	topReferrer, _ := app.DefaultGenesisUsers["root"]

	accounts := [22]string{}
	referrer := topReferrer.String()
	for i := 0; i < 22; i++ {
		_, _, addr := testdata.KeyTestPubAddr()
		s.NoError(
			s.k.AppendChild(s.ctx, referrer, addr.String()),
			s.k.SetActive(s.ctx, addr.String(), true, true),
			s.bk.SetBalance(s.ctx, addr, sdk.Coins{sdk.Coin{
				Denom:  util.ConfigMainDenom,
				Amount: sdk.NewInt(1),
			}, sdk.Coin{
				Denom:  util.ConfigDelegatedDenom,
				Amount: sdk.NewInt(1),
			}}),
		)
		accounts[i] = addr.String()
		referrer = addr.String()
	}

	validatorOn := func(acc string, status types.Status) {
		_, consPubKey, _ := app.NewTestConsPubAddress()
		addr, err := sdk.AccAddressFromBech32(acc)
		s.NoError(
			err,
			s.setStatusHelper(addr.String(), status),
			s.bk.SetBalance(s.ctx, addr, sdk.Coins{sdk.Coin{
				Denom:  util.ConfigDelegatedDenom,
				Amount: sdk.NewInt(10_000_000000),
			}}),
			s.nk.SwitchOn(s.ctx, addr, consPubKey),
		)
	}

	validatorOff := func(acc string) {
		addr, err := sdk.AccAddressFromBech32(acc)
		s.NoError(
			err,
			s.nk.SwitchOff(s.ctx, addr),
		)
	}

	res, err := s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[0])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(0, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")

	validatorOn(accounts[0], types.STATUS_MASTER)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[0])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(0, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	validatorOff(accounts[0])

	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[1])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(0, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")

	validatorOn(accounts[0], types.STATUS_MASTER)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[1])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[0],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[0])

	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(0, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")

	validatorOn(accounts[20], types.STATUS_MASTER)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])

	validatorOn(accounts[20], types.STATUS_BUSINESSMAN)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(3),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])

	validatorOn(accounts[20], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(6),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])

	validatorOn(accounts[17], types.STATUS_MASTER)
	validatorOn(accounts[15], types.STATUS_CHAMPION)
	validatorOn(accounts[11], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[9], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[7], types.STATUS_TOP_LEADER)
	validatorOn(accounts[1], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(6, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[17],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[15],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 2")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[11],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 3")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[9],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 4")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[7],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 5")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[1],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 6")
	validatorOff(accounts[17])
	validatorOff(accounts[15])
	validatorOff(accounts[11])
	validatorOff(accounts[9])
	validatorOff(accounts[7])
	validatorOff(accounts[1])

	validatorOn(accounts[16], types.STATUS_MASTER)
	validatorOn(accounts[14], types.STATUS_CHAMPION)
	validatorOn(accounts[10], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[8], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[6], types.STATUS_TOP_LEADER)
	validatorOn(accounts[0], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(0, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	validatorOff(accounts[16])
	validatorOff(accounts[14])
	validatorOff(accounts[10])
	validatorOff(accounts[8])
	validatorOff(accounts[6])
	validatorOff(accounts[0])

	validatorOn(accounts[20], types.STATUS_MASTER)
	validatorOn(accounts[19], types.STATUS_CHAMPION)
	validatorOn(accounts[18], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[17], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[16], types.STATUS_TOP_LEADER)
	validatorOn(accounts[15], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(6, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[19],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 2")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[18],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 3")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[17],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 4")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[16],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 5")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[15],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 6")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[19], types.STATUS_TOP_LEADER)
	validatorOn(accounts[18], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[17], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[16], types.STATUS_CHAMPION)
	validatorOn(accounts[15], types.STATUS_MASTER)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(6),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[19], types.STATUS_MASTER)
	validatorOn(accounts[18], types.STATUS_CHAMPION)
	validatorOn(accounts[17], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[16], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[15], types.STATUS_TOP_LEADER)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(6),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_MASTER)
	validatorOn(accounts[19], types.STATUS_MASTER)
	validatorOn(accounts[18], types.STATUS_MASTER)
	validatorOn(accounts[17], types.STATUS_MASTER)
	validatorOn(accounts[16], types.STATUS_MASTER)
	validatorOn(accounts[15], types.STATUS_MASTER)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[19], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[18], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[17], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[16], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[15], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(1, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(6),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_MASTER)
	validatorOn(accounts[19], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[18], types.STATUS_CHAMPION)
	validatorOn(accounts[17], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[16], types.STATUS_TOP_LEADER)
	validatorOn(accounts[15], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(5, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[19],
		Ratio:       util.Permille(2),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 2")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[17],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 4")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[16],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 5")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[15],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 6")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_MASTER)
	validatorOn(accounts[19], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[18], types.STATUS_CHAMPION)
	validatorOn(accounts[17], types.STATUS_TOP_LEADER)
	validatorOn(accounts[16], types.STATUS_PROFESSIONAL)
	validatorOn(accounts[15], types.STATUS_ABSOLUTE_CHAMPION)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(4, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[19],
		Ratio:       util.Permille(2),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 2")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[17],
		Ratio:       util.Permille(2),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 3")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[15],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 4")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])

	validatorOn(accounts[20], types.STATUS_MASTER)
	validatorOn(accounts[19], types.STATUS_ABSOLUTE_CHAMPION)
	validatorOn(accounts[18], types.STATUS_BUSINESSMAN)
	validatorOn(accounts[17], types.STATUS_CHAMPION)
	validatorOn(accounts[16], types.STATUS_TOP_LEADER)
	validatorOn(accounts[15], types.STATUS_PROFESSIONAL)
	res, err = s.k.GetReferralValidatorFeesForDelegating(s.ctx, accounts[21])
	s.NoError(err, "GetReferralValidatorFeesForDelegating all newbies: no error")
	s.Equal(2, len(res), "GetReferralValidatorFeesForDelegating all newbies: len")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[20],
		Ratio:       util.Permille(1),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 1")
	s.Contains(res, types.ReferralValidatorFee{
		Beneficiary: accounts[19],
		Ratio:       util.Permille(5),
	}, "GetReferralValidatorFeesForDelegating all newbies: lvl 2")
	validatorOff(accounts[20])
	validatorOff(accounts[19])
	validatorOff(accounts[18])
	validatorOff(accounts[17])
	validatorOff(accounts[16])
	validatorOff(accounts[15])
}

// ----- private functions ------------

func (s *VASuite) setBalance(acc sdk.AccAddress, coins sdk.Coins) error {
	return s.bk.SetBalance(s.ctx, acc, coins)
}

func (s *VASuite) get(acc string) (types.Info, error) {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes := store.Get(keyBytes)
	var value types.Info
	err := s.cdc.UnmarshalBinaryBare(valueBytes, &value)
	return value, err
}

func (s *VASuite) set(acc string, value types.Info) error {
	store := s.ctx.KVStore(s.storeKey)
	keyBytes := []byte(acc)
	valueBytes, err := s.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return err
	}
	store.Set(keyBytes, valueBytes)
	return nil
}

func (s *VASuite) update(acc string, callback func(*types.Info)) error {
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

func (s *VASuite) setStatus(target *types.Info, value types.Status, acc string) {
	if target.Status == value {
		return
	}

	store := s.ctx.KVStore(s.indexStoreKey)
	key := make([]byte, len([]byte(acc))+1)
	copy(key[1:], acc)

	if target.Status >= minIndexedStatus {
		key[0] = uint8(target.Status)
		store.Delete(key)
	}

	target.Status = value
	if value >= minIndexedStatus {
		key[0] = uint8(value)
		store.Set(key, []byte{0x01})
	}
}

func (s *VASuite) setStatusHelper(acc string, nexStatus types.Status) error {
	value, err := s.get(acc)
	if err != nil {
		return err
	}

	s.setStatus(&value, nexStatus, acc)
	err = s.set(acc, value)
	if err != nil {
		return err
	}
	return nil
}
