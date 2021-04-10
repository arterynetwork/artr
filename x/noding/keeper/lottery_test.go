// +build testing

package keeper_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNodingKeeper_Lottery(t *testing.T) {
	suite.Run(t, new(LotterySuite))
}

type LotterySuite struct {
	BaseSuite

	pubKeys []crypto.PubKey
}

func (s *LotterySuite) SetupTest() {
	data, err := ioutil.ReadFile("test-genesis-lottery.json")
	if err != nil {
		panic(err)
	}
	s.setupTest(data)

	for _, key := range []string{
		"artrvalconspub1zcjduepqpme87trszw7awc62ra2de9edwr40v7xy7yfhvpvds96fncagm04qxu308e",
		"artrvalconspub1zcjduepq6ju0rje9444gqkt63k5n2l9ua72545p8c5eqy0d7uhvtxf53c3xq52ydjy",
		"artrvalconspub1zcjduepqh4yvd86v0ej8zu890zlxxypgqjulf6ca3a9szyfpkpjxxw74kz3s7yf9qt",
		"artrvalconspub1zcjduepqka83z5c8huh88w9d2llf3asrth6gt8x5cjqk4gz7xfpk0nshfzeqlcpg6p",
		"artrvalconspub1zcjduepqtczsyayrexuaxg294al04qrvqzg738s9e5jfx82sm87w87w3hq6sc3xq82",
		"artrvalconspub1zcjduepq753pcpuhu2kyugz9z4lyvye222rtjxraazxffqw9yz0rv7m270jqurvy6q",
		"artrvalconspub1zcjduepqucxw7h4cz59c3hdnqucu702fcw556l9c5dyewkjzkjjxgvklxnzqfufx5s",
	} {
		s.pubKeys = append(s.pubKeys, sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, key))
	}
}

func (s *LotterySuite) TestChoose() {
	// Choose the lucky two
	resp, _ := s.nextBlock(
		s.pubKeys[0],
		s.votes(map[int]bool{0: true, 1: true, 2: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{5: 10, 6: 10}, resp.ValidatorUpdates)

	// Keep them until something happens
	resp, _ = s.nextBlock(
		s.pubKeys[1],
		s.votes(map[int]bool{0: true, 1: true, 2: true, 5: true, 6: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{}, resp.ValidatorUpdates)
}

func (s *LotterySuite) TestMissedBlock() {
	resp, _ := s.nextBlock(
		s.pubKeys[0],
		s.votes(map[int]bool{0: true, 1: true, 2: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{5: 10, 6: 10}, resp.ValidatorUpdates)

	// Expel a lucky if they misses a block
	resp, _ = s.nextBlock(
		s.pubKeys[1],
		s.votes(map[int]bool{0: true, 1: true, 2: true, 5: false, 6: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{}, resp.ValidatorUpdates)
	resp, _ = s.nextBlock(
		s.pubKeys[2],
		s.votes(map[int]bool{0: true, 1: true, 2: true, 5: true, 6: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{5: 0, 4: 10}, resp.ValidatorUpdates)
}

func (s *LotterySuite) TestProposedBlock() {
	resp, _ := s.nextBlock(
		s.pubKeys[0],
		s.votes(map[int]bool{0: true, 1: true, 2: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{5: 10, 6: 10}, resp.ValidatorUpdates)

	// Expel a lucky if they successfully proposes a block
	resp, _ = s.nextBlock(
		s.pubKeys[5],
		s.votes(map[int]bool{0: true, 1: true, 2: true, 5: true, 6: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{}, resp.ValidatorUpdates)
	resp, _ = s.nextBlock(
		s.pubKeys[1],
		s.votes(map[int]bool{0: false, 1: true, 2: true, 5: true, 6: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{5: 0, 4: 10}, resp.ValidatorUpdates)
}

func (s *LotterySuite) TestBecomingTop() {
	resp, _ := s.nextBlock(
		s.pubKeys[0],
		s.votes(map[int]bool{0: false, 1: true, 2: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{5: 10, 6: 10}, resp.ValidatorUpdates)

	// Choose a lucky if a chosen one owns its position
	// (The 0'th goes to a jail, the 6'th becomes top-ish and so frees a position, the 4'th takes the position.)
	resp, _ = s.nextBlock(
		s.pubKeys[1],
		s.votes(map[int]bool{0: false, 1: true, 2: true, 5: true, 6: true}),
		nil,
	)
	s.checkUpdates(map[int]int64{0: 0, 4: 10}, resp.ValidatorUpdates)
}

func (s *LotterySuite) votes(data map[int]bool) []abci.VoteInfo {
	var result []abci.VoteInfo
	for n, signed := range data {
		result = append(result, abci.VoteInfo{
			Validator: abci.Validator{
				Address: s.pubKeys[n].Address().Bytes(),
				Power:   10,
			},
			SignedLastBlock: signed,
		})
	}
	return result
}

func (s *LotterySuite) checkUpdates(expected map[int]int64, actual []abci.ValidatorUpdate) {
	s.Equal(len(expected), len(actual), "length")

	extra := len(expected) != len(actual)
	for n, power := range expected {
		ok := false
		key := tmtypes.TM2PB.PubKey(s.pubKeys[n])
		for _, upd := range actual {
			if key.Equal(upd.PubKey) {
				s.Equal(power, upd.Power, "wrong power for validator #%d", n)
				ok = true
				break
			}
		}
		if !ok {
			s.Failf("Not equal:", "power %d for validator #%d missing", power, n)
			extra = true
		}
	}
	if extra {
		for n := 0; n <= 6; n++ {
			if _, ok := expected[n]; ok {
				continue
			}
			key := tmtypes.TM2PB.PubKey(s.pubKeys[n])
			for _, upd := range actual {
				if key.Equal(upd.PubKey) {
					s.Failf("Not equal:", "unexpected power %d for validator #%d", upd.Power, n)
					break
				}
			}
		}
	}
}
