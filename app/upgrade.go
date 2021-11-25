package app

import (
	"encoding/binary"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank"
	bankT "github.com/arterynetwork/artr/x/bank/types"
	delegatingK "github.com/arterynetwork/artr/x/delegating/keeper"
	delegatingT "github.com/arterynetwork/artr/x/delegating/types"
	"github.com/arterynetwork/artr/x/noding"
	referralK "github.com/arterynetwork/artr/x/referral/keeper"
	referralT "github.com/arterynetwork/artr/x/referral/types"
	scheduleK "github.com/arterynetwork/artr/x/schedule/keeper"
	scheduleT "github.com/arterynetwork/artr/x/schedule/types"
	votingKeeper "github.com/arterynetwork/artr/x/voting/keeper"
	votingTypes "github.com/arterynetwork/artr/x/voting/types"
)

func Chain(handlers ...upgrade.UpgradeHandler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		for _, handler := range handlers {
			handler(ctx, plan)
		}
	}
}

func NopUpgradeHandler(_ sdk.Context, _ upgrade.Plan) {}

func RecalculateActiveReferrals(k referralK.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RecalculateActiveReferrals ...")
		k.Iterate(ctx, func(acc string, info *referralT.Info) (changed, checkForStatusUpdate bool) {
			arl := make([]string, 0, len(info.Referrals))
			ars := make(map[string]bool, len(info.Referrals))
			for _, rAddr := range info.Referrals {
				rInfo, err := k.Get(ctx, rAddr)
				if err != nil {
					logger.Error("Account %s not found", rAddr)
					panic(err)
				}
				if rInfo.Active {
					arl = append(arl, rAddr)
					ars[rAddr] = true
				}
			}
			oars := make(map[string]bool, len(info.ActiveReferrals))
			for _, addr := range info.ActiveReferrals {
				if !ars[addr] {
					changed = true
					break
				}
				oars[addr] = true
			}
			if !changed {
				for _, addr := range arl {
					if !oars[addr] {
						changed = true
						break
					}
				}
			}
			if changed {
				logger.Debug("... %s: %v → %v", acc, info.ActiveReferrals, arl)
				info.ActiveReferrals = arl
				return true, true
			}
			return false, false
		})
		logger.Info("... done")
	}
}

func ScheduleBanishment(rk referralK.Keeper, bk bank.Keeper, rKey, sKey sdk.StoreKey, cdc codec.BinaryMarshaler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ScheduleBanishment ...")

		{
			var (
				ms  = ctx.MultiStore().CacheMultiStore()
				ctx = ctx.WithMultiStore(ms)

				sStore = ctx.KVStore(sKey)
				sIt    = sStore.Iterator(nil, nil)

				rStore = ctx.KVStore(rKey)

				dd = bk.GetParams(ctx).DustDelegation
			)
			logger.Info("... fixing schedule ...")
			for ; sIt.Valid(); sIt.Next() {
				var sch scheduleT.Schedule
				cdc.MustUnmarshalBinaryBare(sIt.Value(), &sch)

				changed := false
				tasks := make([]scheduleT.Task, 0, len(sch.Tasks))

				TasksForAMomentInTime:
				for _, task := range sch.Tasks {
					if task.HandlerName == referralK.BanishHookName {
						var addr sdk.AccAddress

						switch len(task.Data) {
						case AddrLen:
							addr = task.Data

							before := string(task.Data)
							after := sdk.AccAddress(task.Data).String()

							task.Data = []byte(after)
							changed = true

							logger.Debug("... ... restoring task data",
								"before", before,
								"after", after,
								"t", task.Time.String(),
							)
						case 43: // Bech32
							var err error
							if addr, err = sdk.AccAddressFromBech32(string(task.Data)); err != nil {
								logger.Error("Length is OK, but still cannot parse",
									"acc", string(task.Data),
									"err", err,
								)
							}
						default:
							logger.Error("Unexpected length", "acc", string(task.Data), "len", len(task.Data))
						}

						var r referralT.Info
						cdc.MustUnmarshalBinaryBare(rStore.Get(task.Data), &r)
						var wrong = r.Active || r.Banished
						if r.CompressionAt != nil && r.CompressionAt.After(ctx.BlockTime()) {
							var found bool

							var sch scheduleT.Schedule
							key := make([]byte, 8)
							binary.BigEndian.PutUint64(key, uint64(r.CompressionAt.UnixNano()))
							if schBz := sStore.Get(key); schBz != nil {
								cdc.MustUnmarshalBinaryBare(schBz, &sch)
								for _, t := range sch.Tasks {
									if t.HandlerName == referralK.CompressionHookName && string(t.Data) == addr.String() {
										found = true
										break
									}
								}
							}

							if found {
								wrong = true
							} else {
								logger.Debug("... ... deleting   CompressionAt",
									"acc", addr.String(),
									"was", r.CompressionAt.String(),
								)
								r.CompressionAt = nil
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							}
						}
						wrong = wrong || len(r.Delegated) > 0 && r.Delegated[0].Int64() > dd

						if wrong {
							logger.Debug("... ... unscheduling",
								"acc", addr.String(),
								"t", task.Time.String(),
							)
							if r.BanishmentAt != nil {
								logger.Debug("... ... erasing    BanishmentAt ",
									"acc", addr.String(),
									"was", r.BanishmentAt.String(),
								)
								r.BanishmentAt = nil
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							}
							changed = true
							continue TasksForAMomentInTime
						} else {
							if r.BanishmentAt == nil {
								logger.Debug("... ... recovering BanishmentAt ",
									"acc", addr.String(),
									"val", task.Time.String(),
								)
								r.BanishmentAt = &task.Time
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							} else if !r.BanishmentAt.Equal(task.Time) {
								logger.Debug("... ... fixing     BanishmentAt ",
									"acc", addr.String(),
									"from", r.BanishmentAt.String(),
									"to", task.Time.String(),
								)
								r.BanishmentAt = &task.Time
								rStore.Set(task.Data, cdc.MustMarshalBinaryBare(&r))
							}
						}
					}
					tasks = append(tasks, task)
				}

				if changed {
					if len(tasks) == 0 {
						sStore.Delete(sIt.Key())
					} else {
						sch.Tasks = tasks
						sStore.Set(sIt.Key(), cdc.MustMarshalBinaryBare(&sch))
					}
				}
			}
			sIt.Close()

			logger.Info("... fixing referral ...")
			var (
				rIt = rStore.Iterator(nil, nil)
			)
			for ; rIt.Valid(); rIt.Next() {
				acc := string(rIt.Key())
				if !strings.HasPrefix(acc, "artr1") {
					logger.Debug("... ... deleting from referral", "acc", acc)
					rStore.Delete(rIt.Key())
					continue
				}

				var r referralT.Info
				cdc.MustUnmarshalBinaryBare(rIt.Value(), &r)
				if r.Banished && r.Active {
					logger.Info("... ... banished yet active, restoring",
						"acc", acc,
					)
					if err := rk.ComeBack(ctx, acc); err != nil {
						logger.Error("Cannot restore unfairly banished account", "acc", acc, "err", err)
					}
				}
			}
			rIt.Close()

			logger.Debug("... persisting multistore")
			ms.Write()
		}
		logger.Info("... ScheduleBanishment done!")
	}
}

func InitPollPeriodParam(k votingKeeper.Keeper, paramspace params.Subspace) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting InitPollPeriodParam ...")

		var pz votingTypes.Params
		paramspace.Get(ctx, votingTypes.KeyParamVotingPeriod, &pz.VotingPeriod)
		pz.PollPeriod = pz.VotingPeriod

		k.SetParams(ctx, pz)
		logger.Info("... InitPollPeriodParam done!", "params", pz)
	}
}

func ForceOnStatusChangedCallback(k noding.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ForceOnStatusChangedCallback ...")

		if validators, err := k.GetActiveValidatorList(ctx); err != nil {
			logger.Error("Cannot get active validator list", "err", err)
		} else {
			for _, acc := range validators {
				if ok, _, _, err := k.IsQualified(ctx, acc); err != nil {
					logger.Error("Cannot check qualification", "err", err, "acc", acc)
				} else if !ok {
					logger.Info("... switching off validator", "acc", acc)
					if err = k.SwitchOff(ctx, acc); err != nil {
						logger.Error("Cannot switch off validator", "err", err, "acc", acc)
					}
				}
			}
		}

		logger.Info("... ForceOnStatusChangedCallback done!")
	}
}

func ForceGlobalDelegation(rk referralK.Keeper, bk bank.Keeper, dk delegatingK.Keeper, sk scheduleK.Keeper, bKey, dKey sdk.StoreKey, cdc codec.BinaryMarshaler) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, plan upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting ForceGlobalDelegation ...")
		defer logger.Info("... ForceGlobalDelegation done!")

		dMain   := sdk.NewInt(0)
		dRevoke := sdk.NewInt(0)

		oneDay := sk.OneDay(ctx)
		t      := time.Unix(0, plan.Time.UnixNano()).Add(oneDay)

		q      := util.FractionZero()
		deltaT := -1 * time.Second
		deltaQ := util.NewFraction(-deltaT.Nanoseconds(), oneDay.Nanoseconds()).Reduce()

		bStore := ctx.KVStore(bKey)
		key := make([]byte, len(bankT.BalancesPrefix) + AddrLen)
		copy(key, bankT.BalancesPrefix)

		rk.Iterate(ctx, func(bech32 string, r *referralT.Info) (changed, _ bool) {
			if r.Banished { return }

			empty := r.Coins[0].Equal(r.Delegated[0])
			for i := 0; i <= 10; i++ {
				if !r.Coins[i].Equal(r.Delegated[i]) {
					r.Delegated[i] = r.Coins[i]
					changed = true
				}
			}
			if empty { return }

			acc, err := sdk.AccAddressFromBech32(bech32)
			if err != nil {
				logger.Error("Cannot parse account address", "acc", bech32, "err", err)
				return false, false
			}

			balance := bk.GetBalance(ctx, acc)

			mainBal := balance.AmountOf(util.ConfigMainDenom)
			if !mainBal.IsZero() {
				balance = balance.
					Sub(sdk.NewCoins(sdk.NewCoin(util.ConfigMainDenom, mainBal))).
					Add(sdk.NewCoin(util.ConfigDelegatedDenom, mainBal))
			}

			revokeBal := balance.AmountOf(util.ConfigRevokingDenom)
			if !revokeBal.IsZero() {
				balance = balance.
					Sub(sdk.NewCoins(sdk.NewCoin(util.ConfigRevokingDenom, revokeBal))).
					Add(sdk.NewCoin(util.ConfigDelegatedDenom, revokeBal))
			}

			di := dk.Get(ctx, acc)
			diChanged := false

			if di == nil { di = &delegatingT.Record{} }

			if len(di.Requests) != 0 {
				for _, r := range di.Requests {
					sk.Delete(ctx, r.Time, delegatingT.RevokeHookName, acc)
				}
				di.Requests = nil
				diChanged = true
			}
			if di.NextAccrue != nil {
				missedPart := util.NewFraction(mainBal.Int64() + revokeBal.Int64(), balance.AmountOf(util.ConfigDelegatedDenom).Int64()).Mul(util.NewFraction(plan.Time.Sub(di.NextAccrue.Add(-oneDay)).Nanoseconds(), oneDay.Nanoseconds()))
				di.MissedPart = &missedPart
				diChanged = true
			} else if balance.AmountOf(util.ConfigDelegatedDenom).Int64() > bk.GetParams(ctx).DustDelegation {
				t = t.Add(deltaT)
				q = q.Add(deltaQ)

				di.NextAccrue = &t
				di.MissedPart = &q
				diChanged = true
				sk.ScheduleTask(ctx, t, delegatingT.AccrueHookName, acc)

				if r.BanishmentAt != nil {
					sk.Delete(ctx, *r.BanishmentAt, referralK.BanishHookName, []byte(bech32))
					r.BanishmentAt = nil
					changed = true
				}
			}
			if diChanged {
				if di.IsEmpty() {
					ctx.KVStore(dKey).Delete(acc)
				} else {
					ctx.KVStore(dKey).Set(acc, cdc.MustMarshalBinaryBare(di))
				}
			}

			copy(key[len(bankT.BalancesPrefix):], acc.Bytes())
			bStore.Set(key, cdc.MustMarshalBinaryBare(&bankT.Balance{Coins: balance}))

			dMain = dMain.Add(mainBal)
			dRevoke = dRevoke.Add(revokeBal)
			return
		})

		supply := bk.GetSupply(ctx)
		supply.Deflate(sdk.NewCoins(
			sdk.NewCoin(util.ConfigMainDenom, dMain),
			sdk.NewCoin(util.ConfigRevokingDenom, dRevoke),
		))
		supply.Inflate(sdk.NewCoins(
			sdk.NewCoin(util.ConfigDelegatedDenom, dMain.Add(dRevoke)),
		))
		bk.SetSupply(ctx, supply)

		// Referral statuses must be refreshed after this
	}
}

func RefreshReferralStatuses(rk referralK.Keeper) upgrade.UpgradeHandler {
	return func(ctx sdk.Context, _ upgrade.Plan) {
		logger := ctx.Logger().With("module", "x/upgrade")
		logger.Info("Starting RefreshReferralStatuses ...")
		defer logger.Info("... RefreshReferralStatuses done!")

		rk.Iterate(ctx, func(_ string, _ *referralT.Info) (changed, checkForStatusUpdate bool) {
			return false, true
		})
	}
}
