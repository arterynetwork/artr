// +build testing

package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/noding"
)

func TestNodingKeeper(t *testing.T) {
	suite.Run(t, new(Suite))
}

type BaseSuite struct {
	suite.Suite

	app     *app.ArteryApp
	cleanup func()

	cdc codec.BinaryMarshaler
	ctx sdk.Context
	k   noding.Keeper
	bk  bank.Keeper
}

type Suite struct {
	BaseSuite
}

func (s *BaseSuite) setupTest(genesis []byte) {
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(genesis)

	s.cdc = s.app.Codec()
	s.k = s.app.GetNodingKeeper()
	s.bk = s.app.GetBankKeeper()
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.setupTest(nil)
}

func (s *Suite) TestSwitchOn() {
	var pubKeys [3]crypto.PubKey
	var tmPubKeys [3]tmcrypto.PublicKey
	for i := 0; i < 3; i++ {
		_, pubKeys[i], _ = app.NewTestConsPubAddress()
		tmPubKeys[i], _ = cryptocodec.ToTmProtoPublicKey(pubKeys[i])
	}
	s.Equal(noding.ErrNotQualified, s.k.SwitchOn(s.ctx, s.user(15), pubKeys[0]))
	s.NoError(s.k.SwitchOn(s.ctx, s.user(2), pubKeys[1]))
	s.NoError(s.k.SwitchOn(s.ctx, s.user(3), pubKeys[2]))

	resp := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.Equal(
		[]abci.ValidatorUpdate{
			{PubKey: tmPubKeys[1], Power: 15},
			{PubKey: tmPubKeys[2], Power: 15},
		},
		resp.ValidatorUpdates,
	)
}

func (s *Suite) TestAddToStaff() {
	s.NoError(s.k.AddToStaff(s.ctx, s.user(15)))

	qualified, _, _, err := s.k.IsQualified(s.ctx, s.user(15))
	s.NoError(err)
	s.True(qualified, "despite of rules")

	_, pubkey, _ := app.NewTestConsPubAddress()
	tmPubKey, _ := cryptocodec.ToTmProtoPublicKey(pubkey)
	s.NoError(s.k.SwitchOn(s.ctx, s.user(15), pubkey))
	resp := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.Equal(
		[]abci.ValidatorUpdate{
			{PubKey: tmPubKey, Power: 15},
		},
		resp.ValidatorUpdates,
	)
}

func (s *Suite) TestRemoveFromStaff() {
	var pubkeys [2]crypto.PubKey
	var tmPubKeys [2]tmcrypto.PublicKey
	for i := 0; i < 2; i++ {
		_, pubkeys[i], _ = app.NewTestConsPubAddress()
		tmPubKeys[i], _ = cryptocodec.ToTmProtoPublicKey(pubkeys[i])
	}
	_ = s.k.AddToStaff(s.ctx, s.user(2))
	_ = s.k.AddToStaff(s.ctx, s.user(15))
	_ = s.k.SwitchOn(s.ctx, s.user(2), pubkeys[0])
	_ = s.k.SwitchOn(s.ctx, s.user(15), pubkeys[1])

	s.nextBlock(pubkeys[0], []abci.VoteInfo{
		{
			Validator: abci.Validator{
				Address: pubkeys[0].Address().Bytes(),
				Power:   10,
			},
			SignedLastBlock: true,
		},
		{
			Validator: abci.Validator{
				Address: pubkeys[1].Address().Bytes(),
				Power:   10,
			},
			SignedLastBlock: true,
		},
	}, nil)

	s.NoError(s.k.RemoveFromStaff(s.ctx, s.user(2)))
	s.NoError(s.k.RemoveFromStaff(s.ctx, s.user(15)))
	resp := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.Equal(
		[]abci.ValidatorUpdate{
			// user2 is qualified, so it remains
			{PubKey: tmPubKeys[1], Power: 0},
		},
		resp.ValidatorUpdates,
	)
}

func (s *Suite) TestProposerAward() {
	balance0 := s.bk.GetBalance(s.ctx, s.user(2)).AmountOf(util.ConfigMainDenom).Int64()

	_, pubkey, _ := app.NewTestConsPubAddress()
	if err := s.k.SwitchOn(s.ctx, s.user(2), pubkey); err != nil {
		panic(err)
	}
	if err := s.bk.SendCoinsFromAccountToModule(
		s.ctx, s.user(1), auth.FeeCollectorName,
		sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, sdk.NewInt(10_000000))),
	); err != nil {
		panic(err)
	}

	s.nextBlock(pubkey, nil, nil)

	balance := s.bk.GetBalance(s.ctx, s.user(2)).AmountOf(util.ConfigMainDenom).Int64()
	s.Equal(int64(10_000000), balance-balance0)
	if data, err := s.k.Get(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.Equal(int64(1), data.ProposedCount)
	}
}

func (s *Suite) TestByzantine() {
	_, pubkey, _ := app.NewTestConsPubAddress()
	tmPubKey, _ := cryptocodec.ToTmProtoPublicKey(pubkey)
	if err := s.k.SwitchOn(s.ctx, s.user(2), pubkey); err != nil {
		panic(err)
	}

	validator := abci.Validator{
		Address: pubkey.Address().Bytes(),
		Power:   10,
	}
	votes := []abci.VoteInfo{{Validator: validator, SignedLastBlock: true}}

	// First infraction
	s.nextBlock(pubkey, votes, []abci.Evidence{{
		Type:             abci.EvidenceType_DUPLICATE_VOTE,
		Validator:        validator,
		Height:           s.ctx.BlockHeight(),
		TotalVotingPower: 20,
	}})

	// Just warning for the first time
	if isValidator, err := s.k.IsValidator(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.True(isValidator)
	}

	// Second infraction
	s.nextBlock(pubkey, votes, []abci.Evidence{{
		Type:             abci.EvidenceType_DUPLICATE_VOTE,
		Validator:        validator,
		Height:           s.ctx.BlockHeight(),
		TotalVotingPower: 20,
	}})

	// After that, a validator's banned for a lifetime
	if isValidator, err := s.k.IsValidator(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.False(isValidator)
	}
	if isBanned, err := s.k.IsBanned(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.True(isBanned)
	}
	resp := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.Equal([]abci.ValidatorUpdate{{PubKey: tmPubKey, Power: 0}}, resp.ValidatorUpdates)

	// Banned node cannot be switched on by any means
	s.Equal(noding.ErrBannedForLifetime, s.k.SwitchOn(s.ctx, s.user(2), pubkey))
	if err := s.k.AddToStaff(s.ctx, s.user(2)); err != nil {
		panic(err)
	}
	s.Equal(noding.ErrBannedForLifetime, s.k.SwitchOn(s.ctx, s.user(2), pubkey))
	if isValidator, err := s.k.IsValidator(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.False(isValidator)
	}
}

func (s *Suite) TestJailing() {
	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	_, pubkey, _ := app.NewTestConsPubAddress()
	tmPubKey, _ := cryptocodec.ToTmProtoPublicKey(pubkey)
	if err := s.k.SwitchOn(s.ctx, s.user(2), pubkey); err != nil {
		panic(err)
	}

	validator := abci.Validator{
		Address: pubkey.Address().Bytes(),
		Power:   10,
	}
	votes := []abci.VoteInfo{{Validator: validator, SignedLastBlock: false}}

	// First missed block
	s.nextBlock(proposerKey, votes, nil)

	if data, err := s.k.Get(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.Equal(int64(0), data.OkBlocksInRow)
		s.Equal(int64(1), data.MissedBlocksInRow)
		s.Equal(int64(1), data.Strokes)
		s.False(data.Jailed)
		s.Equal(int64(0), data.JailCount)
	}

	// Second missed block
	s.nextBlock(proposerKey, votes, nil)

	if data, err := s.k.Get(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.Equal(int64(0), data.OkBlocksInRow)
		s.Equal(int64(0), data.MissedBlocksInRow) // zeroed because of jail
		s.Equal(int64(2), data.Strokes)
		s.True(data.Jailed)
		s.Equal(int64(123), data.UnjailAt)
		s.Equal(int64(1), data.JailCount)
	}
	if isValidator, err := s.k.IsValidator(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.False(isValidator)
	}
	resp, _ := s.nextBlock(proposerKey, votes, nil)
	s.Equal([]abci.ValidatorUpdate{{PubKey: tmPubKey, Power: 0}}, resp.ValidatorUpdates)

	s.Equal(noding.ErrJailPeriodNotOver, s.k.Unjail(s.ctx, s.user(2)))
	if isValidator, err := s.k.IsValidator(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.False(isValidator)
	}

	// One hour later
	s.ctx = s.ctx.WithBlockHeight(123)
	s.NoError(s.k.Unjail(s.ctx, s.user(2)))
	if isValidator, err := s.k.IsValidator(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.True(isValidator)
	}
	resp, _ = s.nextBlock(proposerKey, nil, nil)
	s.Equal([]abci.ValidatorUpdate{{PubKey: tmPubKey, Power: 15}}, resp.ValidatorUpdates)
}

func (s *Suite) TestSwitchOnAfterSwitchOffWhileJailed() {
	user2 := s.user(2)
	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	_, pubkey, _ := app.NewTestConsPubAddress()
	s.NoError(s.k.SwitchOn(s.ctx, user2, pubkey))

	validator := abci.Validator{
		Address: pubkey.Address().Bytes(),
		Power:   10,
	}
	votes := []abci.VoteInfo{{Validator: validator, SignedLastBlock: false}}

	s.nextBlock(proposerKey, votes, nil)
	s.nextBlock(proposerKey, votes, nil)

	isValidator, err := s.k.IsValidator(s.ctx, user2)
	s.NoError(err)
	s.False(isValidator)

	s.NoError(s.k.SwitchOff(s.ctx, user2))
	s.NoError(s.k.SwitchOn(s.ctx, user2, pubkey))

	isValidator, err = s.k.IsValidator(s.ctx, user2)
	s.NoError(err)
	s.False(isValidator)
	data, err := s.k.Get(s.ctx, s.user(2))
	s.NoError(err)
	s.True(data.Jailed)
	s.Equal(int64(123), data.UnjailAt)
}

func (s Suite) TestDoubleSwitchOn() {
	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)

	user := s.user(2)
	_, pubkey1, _ := app.NewTestConsPubAddress()
	_, pubkey2, _ := app.NewTestConsPubAddress()
	s.NoError(s.k.SwitchOn(s.ctx, user, pubkey1))

	s.nextBlock(proposerKey, nil, nil)
	s.Equal(noding.ErrAlreadyOn, s.k.SwitchOn(s.ctx, user, pubkey2))
	reb, _ := s.nextBlock(proposerKey, nil, nil)
	s.Empty(reb.ValidatorUpdates)
}

func (s Suite) TestDoubleSwitchOnWithJail() {
	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)

	user := s.user(2)
	_, pubkey1, consAddr := app.NewTestConsPubAddress()
	tmPubKey1, _ := cryptocodec.ToTmProtoPublicKey(pubkey1)
	_, pubkey2, _ := app.NewTestConsPubAddress()
	s.NoError(s.k.SwitchOn(s.ctx, user, pubkey1))

	votes := []abci.VoteInfo{{Validator: abci.Validator{Address: consAddr}, SignedLastBlock: false}}
	s.nextBlock(proposerKey, votes, nil)
	s.nextBlock(proposerKey, votes, nil)
	data, err := s.k.Get(s.ctx, s.user(2))
	s.NoError(err)
	s.True(data.Jailed)
	reb, _ := s.nextBlock(proposerKey, nil, nil)
	s.Equal([]abci.ValidatorUpdate{{PubKey: tmPubKey1, Power: 0}}, reb.ValidatorUpdates)

	s.Equal(noding.ErrAlreadyOn, s.k.SwitchOn(s.ctx, user, pubkey2))
	reb, _ = s.nextBlock(proposerKey, nil, nil)
	s.Empty(reb.ValidatorUpdates)
}

func (s Suite) TestNodeNodeLeap() {
	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	user := s.user(2)
	_, pubkey, _ := app.NewTestConsPubAddress()
	tmPubKey, _ := cryptocodec.ToTmProtoPublicKey(pubkey)

	s.NoError(s.k.SwitchOn(s.ctx, user, pubkey))

	validator := abci.Validator{
		Address: pubkey.Address().Bytes(),
		Power:   10,
	}
	votes := []abci.VoteInfo{{Validator: validator, SignedLastBlock: true}}

	s.nextBlock(proposerKey, votes, nil)
	s.nextBlock(proposerKey, votes, nil)

	data, err := s.k.Get(s.ctx, user)
	s.NoError(err)
	s.Equal(int64(2), data.OkBlocksInRow)

	_, newPubkey, _ := app.NewTestConsPubAddress()
	tmNewPubKey, _ := cryptocodec.ToTmProtoPublicKey(newPubkey)
	s.NoError(s.k.SwitchOff(s.ctx, user))
	s.NoError(s.k.SwitchOn(s.ctx, user, newPubkey))

	ebr, _ := s.nextBlock(proposerKey, votes, nil)
	s.Equal(
		[]abci.ValidatorUpdate{
			{
				PubKey: tmPubKey,
				Power:  0,
			},
			{
				PubKey: tmNewPubKey,
				Power:  15,
			},
		},
		ebr.ValidatorUpdates,
	)

	data, err = s.k.Get(s.ctx, user)
	s.NoError(err)
	s.Equal(int64(3), data.OkBlocksInRow)

	validator.Address = newPubkey.Address().Bytes()
	votes = []abci.VoteInfo{{Validator: validator, SignedLastBlock: true}}
	s.nextBlock(proposerKey, votes, nil)

	data, err = s.k.Get(s.ctx, user)
	s.NoError(err)
	s.Equal(int64(4), data.OkBlocksInRow)
}

func (s *Suite) TestDoubleJail() {
	proposerKey := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	_, pubkey, _ := app.NewTestConsPubAddress()
	if err := s.k.SwitchOn(s.ctx, s.user(2), pubkey); err != nil {
		panic(err)
	}

	validator := abci.Validator{
		Address: pubkey.Address().Bytes(),
		Power:   10,
	}
	votes := []abci.VoteInfo{{Validator: validator, SignedLastBlock: false}}

	// 4 missed blocks in row (suppose Tendermint lagged and didn't exclude the validator in time)
	for i := 0; i < 4; i++ {
		s.nextBlock(proposerKey, votes, nil)
	}

	if data, err := s.k.Get(s.ctx, s.user(2)); err != nil {
		panic(err)
	} else {
		s.Equal(int64(0), data.OkBlocksInRow)
		s.Equal(int64(0), data.MissedBlocksInRow)
		s.Equal(int64(2), data.Strokes)
		s.True(data.Jailed)
		s.Equal(int64(1), data.JailCount)
	}
}

func (s *BaseSuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *BaseSuite) nextBlock(proposer crypto.PubKey, votes []abci.VoteInfo, byzantine []abci.Evidence) (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})

	s.ctx = s.ctx.
		WithBlockHeight(s.ctx.BlockHeight() + 1).
		WithBlockTime(s.ctx.BlockTime().Add(30*time.Second))

	bbr := s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{
		Header: tmproto.Header{
			ProposerAddress: proposer.Address().Bytes(),
		},
		LastCommitInfo: abci.LastCommitInfo{
			Votes: votes,
		},
		ByzantineValidators: byzantine,
	})
	return ebr, bbr
}

func (s *BaseSuite) user(n int) sdk.AccAddress {
	return app.DefaultGenesisUsers[fmt.Sprintf("user%d", n)]
}
