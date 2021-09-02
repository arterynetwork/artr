package keeper

import (
	"github.com/pkg/errors"

	"github.com/arterynetwork/artr/x/referral/types"
)

func checkStatusRequirements(status types.Status, value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
	if value.Banished {
		return types.StatusCheckResult{
			Overall: false,
			Criteria: []types.StatusCheckResult_Criterion{
				{
					Met:         false,
					Rule:        types.RULE_PARTICIPATE_IN_REFERRAL_PROGRAM,
					TargetValue: 1,
					ActualValue: 0,
				},
			},
		}, nil
	}
	return statusRequirements[status](value, bu)
}

var statusRequirements = map[types.Status]func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error){
	types.STATUS_LUCKY: func(_ types.Info, _ *bunchUpdater) (types.StatusCheckResult, error) {
		return types.StatusCheckResult{Overall: true}, nil
	},
	types.STATUS_LEADER: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsXByX(value, bu, 2, 2)
	},
	types.STATUS_MASTER: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsXByX(value, bu, 3, 3)
	},
	types.STATUS_CHAMPION: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.STATUS_MASTER.LinesOpened(), 0, 15)
	},
	types.STATUS_BUSINESSMAN: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.STATUS_CHAMPION.LinesOpened(), 150_000, 60)
	},
	types.STATUS_PROFESSIONAL: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.STATUS_BUSINESSMAN.LinesOpened(), 300_000, 200)
	},
	types.STATUS_TOP_LEADER: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.STATUS_PROFESSIONAL.LinesOpened(), 1_000_000, 500)
	},
	types.STATUS_HERO: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.STATUS_TOP_LEADER.LinesOpened(), 2_000_000, 1_000)
	},
	types.STATUS_ABSOLUTE_CHAMPION: func(value types.Info, bu *bunchUpdater) (types.StatusCheckResult, error) {
		return statusRequirementsCore(value, bu, types.STATUS_HERO.LinesOpened(), 5_000_000, 2_000)
	},
}

func statusRequirementsXByX(value types.Info, bu *bunchUpdater, count int, size int) (types.StatusCheckResult, error) {
	var (
		result    = types.StatusCheckResult{}
		criterion types.StatusCheckResult_Criterion
	)

	criterion = types.StatusCheckResult_Criterion{
		Rule:        types.RULE_N_REFERRALS_WITH_X_REFERRALS_EACH,
		TargetValue: uint64(count),
		ParameterX:  uint64(size),
	}

	found := 0
	for _, childAcc := range value.ActiveReferrals {
		child, err := bu.get(childAcc)
		if err != nil {
			return result, err
		}
		if child.ActiveRefCounts[1] >= uint64(size) {
			found++
			if found >= count {
				criterion.Met = true
				break
			}
		}
	}
	criterion.ActualValue = uint64(found)

	result.Criteria = []types.StatusCheckResult_Criterion{criterion}
	result.Overall = criterion.Met
	return result, nil
}

func statusRequirementsCore(value types.Info, bu *bunchUpdater, linesOpen int, coins int64, leg uint64) (types.StatusCheckResult, error) {
	var (
		result    = types.StatusCheckResult{Overall: true}
		criterion types.StatusCheckResult_Criterion
	)

	if coins > 0 {
		criterion = types.StatusCheckResult_Criterion{
			Rule:        types.RULE_N_COINS_IN_STRUCTURE,
			TargetValue: uint64(coins),
			ActualValue: value.CoinsAtLevelsUpTo(linesOpen).Uint64() / 1_000_000,
		}
		if criterion.ActualValue >= criterion.TargetValue {
			criterion.Met = true
			criterion.ActualValue = criterion.TargetValue
		}
		result.Criteria = append(result.Criteria, criterion)
		result.Overall = result.Overall && criterion.Met
	}

	criterion = types.StatusCheckResult_Criterion{
		Rule:        types.RULE_N_TEAMS_OF_X_PEOPLE_EACH,
		TargetValue: 3,
		ParameterX:  leg,
	}
	xByX := types.StatusCheckResult_Criterion{
		Rule:        types.RULE_N_REFERRALS_WITH_X_REFERRALS_EACH,
		TargetValue: 3,
		ParameterX:  3,
	}
	for _, childAcc := range value.ActiveReferrals {
		child, err := bu.get(childAcc)
		if err != nil {
			result.Overall = false
			return result, errors.Wrapf(err, `cannot obtain data for referral "%s"`, childAcc)
		}
		var s uint64
		if !criterion.Met {
			for _, x := range child.ActiveRefCounts {
				s += x
				if s >= leg {
					criterion.ActualValue++
					if criterion.ActualValue >= criterion.TargetValue {
						criterion.Met = true
					}
					break
				}
			}
		}
		if !xByX.Met {
			if child.ActiveRefCounts[1] >= xByX.ParameterX {
				xByX.ActualValue++
				if xByX.ActualValue >= xByX.TargetValue {
					xByX.Met = true
				}
			}
		}
		if criterion.Met && xByX.Met {
			break
		}
	}
	result.Criteria = append(result.Criteria, criterion, xByX)
	result.Overall = result.Overall && criterion.Met && xByX.Met

	return result, nil
}
