package noding_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/noding"
	"github.com/arterynetwork/artr/x/noding/types"
)

func TestNodingGenesis(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite

	app       *app.ArteryApp
	cleanup   func()
	ctx       sdk.Context
	k         noding.Keeper
}

func (s *Suite) SetupTest() {
	s.app, s.cleanup = app.NewAppFromGenesis(nil)
	s.ctx = s.app.NewContext(true, abci.Header{Height: 1})
	s.k   = s.app.GetNodingKeeper()
}

func (s *Suite) TearDownTest() {
	s.cleanup()
}

func (s Suite) TestCleanGenesis() {
	s.checkExportImport()
}

func (s Suite) TestBlocksInRowAndJail() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	user1key := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	user1ca := sdk.ConsAddress(user1key.Address().Bytes())
	_, user2key, user2ca := app.NewTestConsPubAddress()
	_, user3key, user3ca := app.NewTestConsPubAddress()

	if err := s.k.SwitchOn(s.ctx, user2, user2key, false); err != nil { panic(err) }
	if err := s.k.SwitchOn(s.ctx, user3, user3key, true); err != nil { panic(err) }

	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user3ca, Power: 1}, SignedLastBlock: false},
	}, nil)
	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: false},
		{Validator: abci.Validator{Address: user3ca, Power: 1}, SignedLastBlock: false},
	}, nil)
	{
		d, _ := s.k.Get(s.ctx, user3)
		s.True(d.Jailed)
	}
	s.checkExportImport()
}

func (s Suite) TestJailAndSwitchOff() {
	user2 := app.DefaultGenesisUsers["user2"]
	user1key := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	user1ca := sdk.ConsAddress(user1key.Address().Bytes())
	_, user2key, user2ca := app.NewTestConsPubAddress()

	if err := s.k.SwitchOn(s.ctx, user2, user2key, false); err != nil { panic(err) }

	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: false},
	}, nil)
	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: false},
	}, nil)
	{
		d, _ := s.k.Get(s.ctx, user2)
		s.True(d.Jailed)
	}
	s.NoError(s.k.SwitchOff(s.ctx, user2))
	s.checkExportImport()
}

func (s Suite) TestByzantine() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	user1key := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	user1ca := sdk.ConsAddress(user1key.Address().Bytes())
	_, user2key, user2ca := app.NewTestConsPubAddress()
	_, user3key, user3ca := app.NewTestConsPubAddress()

	if err := s.k.SwitchOn(s.ctx, user2, user2key, false); err != nil { panic(err) }
	if err := s.k.SwitchOn(s.ctx, user3, user3key, true); err != nil { panic(err) }

	val1 := abci.Validator{Address: user1ca, Power: 10}
	val2 := abci.Validator{Address: user2ca, Power: 10}
	val3 := abci.Validator{Address: user3ca, Power: 1}

	s.nextBlock(
		user1key,
		[]abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: true},
		},
		[]abci.Evidence{
			{
				Type:      "evil_deed",
				Validator: val2,
				Height:    s.ctx.BlockHeight(),
			},
		},
	)
	s.nextBlock(
		user1key,
		[]abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: false},
			{Validator: val3, SignedLastBlock: false},
		},
		[]abci.Evidence{
			{
				Type:      "evil_deed",
				Validator: val2,
				Height:    s.ctx.BlockHeight(),
			},
			{
				Type:      "evil_deed",
				Validator: val3,
				Height:    s.ctx.BlockHeight(),
			},
		},
	)
	{
		d, _ := s.k.Get(s.ctx, user2)
		s.True(d.BannedForLife)
		d, _ = s.k.Get(s.ctx, user3)
		s.False(d.BannedForLife)
	}
	s.checkExportImport()
}

func (s Suite) TestStaff() {
	s.NoError(s.k.AddToStaff(s.ctx, app.DefaultGenesisUsers["user1"]))
	s.NoError(s.k.AddToStaff(s.ctx, app.DefaultGenesisUsers["user13"]))
	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		[]string{
			noding.StoreKey,
		},
		map[string]app.Decoder{
			noding.StoreKey: app.AccAddressDecoder,
		},
		map[string]app.Decoder{
			noding.StoreKey: func(bz []byte)(string, error){
				var value types.D
				err := s.app.Codec().UnmarshalBinaryLengthPrefixed(bz, &value)
				if err != nil { return "", err }
				return fmt.Sprintf("%+v", value), nil
			},
		},)
}

func (s Suite) nextBlock(proposer crypto.PubKey, votes []abci.VoteInfo, byzantine []abci.Evidence) (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	bbr := s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{
		Header: abci.Header{
			ProposerAddress: proposer.Address().Bytes(),
		},
		LastCommitInfo: abci.LastCommitInfo{
			Votes: votes,
		},
		ByzantineValidators: byzantine,
	})
	return ebr, bbr
}