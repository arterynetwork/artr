package keeper

import (
	"github.com/arterynetwork/artr/x/referral/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var statusRequirements = map[types.Status]func(value types.R, bu bunchUpdater) (types.StatusCheckResult, error) {
	types.Lucky: func (_ types.R, _ bunchUpdater) (types.StatusCheckResult, error) {
		return types.NewStatusCheckResult(), nil
	},
	types.Leader: func (value types.R, _ bunchUpdater) (types.StatusCheckResult, error) {
		var (
			result = types.NewStatusCheckResult()
			criterion string
		)

		criterion = "2 active accounts in the 1st line"
		if value.ActiveReferralsCount[1] >= 2 {
			result.Criteria[criterion] = true
		} else {
			result.Criteria[criterion] = false
			result.Overall             = false
		}

		criterion = "4 active accounts in the 2nd line"
		if value.ActiveReferralsCount[2] >= 4 {
			result.Criteria[criterion] = true
		} else {
			result.Criteria[criterion] = false
			result.Overall             = false
		}

		return result, nil
	},
	types.Master: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error) {
		var (
			result = types.NewStatusCheckResult()
			criterion string
		)

		criterion = "3 active account with 3 active referrals each in the 1st line"
		if value.ActiveReferralsCount[1] < 3 {
			result.Criteria[criterion] = false
			result.Overall = false
			return result, nil
		}
		triples := 0
		for _, childAcc := range value.Referrals {
			child, err := bu.get(childAcc)
			if err != nil {
				return result, err
			}
			if !child.Active {
				continue
			}
			if child.ActiveReferralsCount[1] >= 3 {
				triples++
				if triples >= 3 {
					result.Criteria[criterion] = true
					return result, nil
				}
			}
		}

		result.Criteria[criterion] = false
		result.Overall = false
		return result, nil
	},
	types.Champion: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error)  {
		return statusRequirementsCore(value, bu, types.Master.LinesOpened(), 0, 15)
	},
	types.Businessman: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error)  {
		return statusRequirementsCore(value, bu, types.Champion.LinesOpened(), 150_000_000000, 60)
	},
	types.Professional: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error)  {
		return statusRequirementsCore(value, bu, types.Businessman.LinesOpened(), 300_000_000000, 200)
	},
	types.TopLeader: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error)  {
		return statusRequirementsCore(value, bu, types.Professional.LinesOpened(), 1_000_000_000000, 500)
	},
	types.Hero: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error)  {
		return statusRequirementsCore(value, bu, types.TopLeader.LinesOpened(), 2_000_000_000000, 1_000)
	},
	types.AbsoluteChampion: func (value types.R, bu bunchUpdater) (types.StatusCheckResult, error)  {
		return statusRequirementsCore(value, bu, types.Hero.LinesOpened(), 5_000_000_000000, 2_000)
	},
}

func statusRequirementsCore(value types.R, bu bunchUpdater, linesOpen int, coins int64, leg int) (types.StatusCheckResult, error) {
	var (
		result = types.NewStatusCheckResult()
		criterion string
	)

	if coins > 0 {
		criterion = fmt.Sprintf("%d+ ARTR in the structure", coins/1_000000)
		if value.CoinsAtLevelsUpTo(linesOpen).GTE(sdk.NewInt(coins)) {
			result.Criteria[criterion] = true
		} else {
			result.Criteria[criterion] = false
			result.Overall = false
		}
	}

	criterion = fmt.Sprintf("3 teams of %d each", leg)
	if value.ActiveReferralsCount[1] < 3 {
		result.Criteria[criterion] = false
		result.Overall = false
		return result, nil
	}
	legs := 0
	for _, childAcc := range value.Referrals {
		child, err := bu.get(childAcc)
		if err != nil {
			return result, err
		}
		if !child.Active {
			continue
		}
		s := 0
		for _, x := range child.ActiveReferralsCount[1:] {
			s += x
		}
		if s >= leg {
			legs++
			if legs >= 3 {
				result.Criteria[criterion] = true
				return result, nil
			}
		}
	}
	result.Criteria[criterion] = false
	result.Overall = false
	return result, nil
}
