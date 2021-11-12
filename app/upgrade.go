package app

import (
	"encoding/binary"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/arterynetwork/artr/x/bank"
	referralK "github.com/arterynetwork/artr/x/referral/keeper"
	referralT "github.com/arterynetwork/artr/x/referral/types"
	scheduleT "github.com/arterynetwork/artr/x/schedule/types"
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
