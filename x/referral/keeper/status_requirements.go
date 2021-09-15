package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/referral/types"
)

func checkStatusRequirements(status types.Status, value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
	if status == 0 {
		return types.StatusCheckResult{Overall: true}, nil
	}
	if value.Banished {
		return types.StatusCheckResult{
			Overall: false,
			Criteria: map[string]bool{
				"participate in referral program": false,
			},
		}, nil
	}
	return statusRequirements[status](value, bu)
}

var statusRequirements = map[types.Status]func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error){
	types.Lucky: func(_ types.R, _ *bunchUpdater) (types.StatusCheckResult, error) {
		return types.StatusCheckResult{Overall: true}, nil
	},
	types.Leader: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsXByX(value, bu, 2, 2)
	},
	types.Master: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsXByX(value, bu, 3, 3)
	},
	types.Champion: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.Master.LinesOpened(), 0, 15)
	},
	types.Businessman: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.Champion.LinesOpened(), 150_000_000000, 60)
	},
	types.Professional: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.Businessman.LinesOpened(), 300_000_000000, 200)
	},
	types.TopLeader: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.Professional.LinesOpened(), 1_000_000_000000, 500)
	},
	types.Hero: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.TopLeader.LinesOpened(), 2_000_000_000000, 1_000)
	},
	types.AbsoluteChampion: func(value types.R, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.Hero.LinesOpened(), 5_000_000_000000, 2_000)
	},
}

func statusRequirementsXByX(value types.R, bu *bunchUpdater, count int, size int) (types.StatusCheckResult, error) {
	var (
		result    = types.NewStatusCheckResult()
		criterion string
	)

	criterion = fmt.Sprintf("%d active accounts with %d active referrals each in the 1st line", count, size)
	if value.ActiveReferralsCount[1] < count || value.ActiveReferralsCount[2] < count*size {
		result.Criteria[criterion] = false
		result.Overall = false
		return result, nil
	}
	found := 0
	for _, childAcc := range value.Referrals {
		child, err := bu.get(childAcc)
		if err != nil {
			return result, err
		}
		if !child.Active {
			continue
		}
		if child.ActiveReferralsCount[1] >= size {
			found++
			if found >= count {
				result.Criteria[criterion] = true
				return result, nil
			}
		}
	}

	result.Criteria[criterion] = false
	result.Overall = false
	return result, nil
}

func statusRequirementsCore(value types.R, bu *bunchUpdater, linesOpen int, coins int64, leg int) (types.StatusCheckResult, error) {
	var (
		result    = types.NewStatusCheckResult()
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
	xByX := fmt.Sprintf("%d active accounts with %d active referrals each in the 1st line", 3, 3)
	result.Criteria[criterion] = false
	result.Criteria[xByX] = false

	if value.ActiveReferralsCount[1] < 3 {
		result.Overall = false
		return result, nil
	}

	legs := 0
	foundByX := 0
	for _, childAcc := range value.Referrals {
		child, err := bu.get(childAcc)
		if err != nil {
			return result, err
		}
		if !child.Active {
			continue
		}
		if !result.Criteria[criterion] {
			s := 0
			for _, x := range child.ActiveReferralsCount {
				s += x
			}
			if s >= leg {
				legs++
				if legs >= 3 {
					result.Criteria[criterion] = true
				}
			}
		}
		if !result.Criteria[xByX] {
			if child.ActiveReferralsCount[1] >= 3 {
				foundByX++
				if foundByX >= 3 {
					result.Criteria[xByX] = true
				}
			}
		}
		if result.Criteria[criterion] && result.Criteria[xByX] {
			return result, nil
		}
	}
	result.Overall = false
	return result, nil
}
