// +build testing

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/arterynetwork/artr/x/noding/keeper"
)

func TestVPGenerator(t *testing.T) {
	suite.Run(t, new(VpgSuite))
}

type VpgSuite struct {
	BaseSuite
}

func (s *VpgSuite) SetupTest() {
	defer func() {
		if e := recover(); e != nil {
			s.FailNow("panic on setup", e)
		}
	}()
	s.setupTest(nil)
}

func (s *VpgSuite) TestAllDifferent() {
	vpg := keeper.NewVotingPowerGenerator(s.k, s.ctx)

	for i := 0; i < 15; i++ {
		s.EqualValues(15, vpg.GetVotingPower(int64(100-i)), "i = %d", i)
	}
	for i := 15; i < 100; i++ {
		s.EqualValues(10, vpg.GetVotingPower(int64(100-i)), "i = %d", i)
	}
}

func (s *VpgSuite) TestAllEqual() {
	vpg := keeper.NewVotingPowerGenerator(s.k, s.ctx)

	for i := 0; i < 100; i++ {
		s.EqualValues(15, vpg.GetVotingPower(100), "i = %d", i)
	}
}
