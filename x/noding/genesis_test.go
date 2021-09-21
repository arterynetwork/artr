// +build testing

package noding_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
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

	app     *app.ArteryApp
	cleanup func()
	ctx     sdk.Context
	k       noding.Keeper
}

func (s *Suite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.app, s.cleanup, s.ctx = app.NewAppFromGenesis(nil)
	s.k = s.app.GetNodingKeeper()
}

func (s *Suite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
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

	if err := s.k.SwitchOn(s.ctx, user2, user2key); err != nil {
		panic(err)
	}
	if err := s.k.SwitchOn(s.ctx, user3, user3key); err != nil {
		panic(err)
	}

	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user3ca, Power: 10}, SignedLastBlock: false},
	}, nil)
	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: false},
		{Validator: abci.Validator{Address: user3ca, Power: 10}, SignedLastBlock: false},
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

	if err := s.k.SwitchOn(s.ctx, user2, user2key); err != nil {
		panic(err)
	}

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

func (s Suite) TestUnjail() {
	user2 := app.DefaultGenesisUsers["user2"]
	user1key := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	user1ca := sdk.ConsAddress(user1key.Address().Bytes())
	_, user2key, user2ca := app.NewTestConsPubAddress()

	if err := s.k.SwitchOn(s.ctx, user2, user2key); err != nil {
		panic(err)
	}

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
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + types.DefaultUnjailAfter)

	s.NoError(s.k.Unjail(s.ctx, user2))
	s.nextBlock(user1key, []abci.VoteInfo{
		{Validator: abci.Validator{Address: user1ca, Power: 10}, SignedLastBlock: true},
		{Validator: abci.Validator{Address: user2ca, Power: 10}, SignedLastBlock: true},
	}, nil)
	s.checkExportImport()
}

func (s Suite) TestByzantine() {
	user2 := app.DefaultGenesisUsers["user2"]
	user3 := app.DefaultGenesisUsers["user3"]
	user1key := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	user1ca := sdk.ConsAddress(user1key.Address().Bytes())
	_, user2key, user2ca := app.NewTestConsPubAddress()
	_, user3key, user3ca := app.NewTestConsPubAddress()

	if err := s.k.SwitchOn(s.ctx, user2, user2key); err != nil {
		panic(err)
	}
	if err := s.k.SwitchOn(s.ctx, user3, user3key); err != nil {
		panic(err)
	}

	val1 := abci.Validator{Address: user1ca, Power: 10}
	val2 := abci.Validator{Address: user2ca, Power: 10}
	val3 := abci.Validator{Address: user3ca, Power: 10}

	s.nextBlock(
		user1key,
		[]abci.VoteInfo{
			{Validator: val1, SignedLastBlock: true},
			{Validator: val2, SignedLastBlock: true},
			{Validator: val3, SignedLastBlock: true},
		},
		[]abci.Evidence{
			{
				Type:      abci.EvidenceType_DUPLICATE_VOTE,
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
				Type:      abci.EvidenceType_DUPLICATE_VOTE,
				Validator: val2,
				Height:    s.ctx.BlockHeight(),
			},
			{
				Type:      abci.EvidenceType_DUPLICATE_VOTE,
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

func (s Suite) TestProposers() {
	user1 := app.DefaultGenesisUsers["user1"]
	user2 := app.DefaultGenesisUsers["user2"]
	user1key := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, app.DefaultUser1ConsPubKey)
	_, user2key, _ := app.NewTestConsPubAddress()
	s.NoError(s.k.SwitchOn(s.ctx, user2, user2key))

	s.nextBlock(user1key, nil, nil)
	s.nextBlock(user2key, nil, nil)
	s.Equal([]uint64{1}, s.k.GetBlocksProposedBy(s.ctx, user1))
	s.Equal([]uint64{2}, s.k.GetBlocksProposedBy(s.ctx, user2))

	s.checkExportImport()
}

func (s Suite) checkExportImport() {
	s.app.CheckExportImport(s.T(),
		s.ctx.BlockTime(),
		[]string{
			noding.StoreKey,
			noding.IdxStoreKey,
		},
		map[string]app.Decoder{
			noding.StoreKey: app.AccAddressDecoder,
			noding.IdxStoreKey: func(bz []byte) (string, error) {
				switch bz[0] {
				case 0x01:
					if len(bz) != 21 {
						return "", fmt.Errorf("wrong address length")
					}
					consAddr := sdk.ConsAddress(bz[1:])
					return consAddr.String(), nil
				case 0x02:
					if len(bz) != 9 {
						return "", fmt.Errorf("wrongth height length")
					}
					height := binary.BigEndian.Uint64(bz[1:])
					return fmt.Sprintf("H %d", height), nil
				default:
					return "", fmt.Errorf("unknown prefix")
				}
			},
		},
		map[string]app.Decoder{
			noding.StoreKey: func(bz []byte) (string, error) {
				var value types.Info
				err := s.app.Codec().UnmarshalBinaryBare(bz, &value)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%+v", value), nil
			},
			noding.IdxStoreKey: app.AccAddressDecoder,
		},
		map[string][][]byte{
			noding.IdxStoreKey: {{0x01}},
		},
	)
}

func (s *Suite) nextBlock(proposer crypto.PubKey, votes []abci.VoteInfo, byzantine []abci.Evidence) (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	ebr := s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
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
