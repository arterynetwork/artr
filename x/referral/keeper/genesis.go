package keeper

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/referral/types"
)

func (k Keeper) ExportToGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	var (
		data         types.Info
		err          error
		params       types.Params
		topLevel     []string
		other        []types.Refs
		banished     []types.Banished
		neverPaid    []string
		compressions []types.Compression
		downgrades   []types.Downgrade
		transitions  []types.Transition

		children  []string
		thisLevel []types.Refs
		nextLevel []types.Refs
	)
	params = k.GetParams(ctx)
	topLevel, banished, neverPaid, err = k.GetTopLevelAndBanishedAccounts(ctx)
	if err != nil {
		return nil, err
	}

	for _, addr := range topLevel {
		data, err = k.Get(ctx, addr)
		if err != nil {
			return nil, err
		}
		if data.CompressionAt != nil {
			compressions = append(compressions, *types.NewCompression(addr, *data.CompressionAt))
		}
		if data.StatusDowngradeAt != nil {
			downgrades = append(downgrades, *types.NewDowngrade(addr, data.Status, *data.StatusDowngradeAt))
		}
		if data.Transition != "" {
			transitions = append(transitions, types.Transition{
				Subject:     addr,
				Destination: data.Transition,
			})
		}
		children, err = k.GetChildren(ctx, addr)
		if err != nil {
			return nil, err
		}
		if len(children) == 0 {
			continue
		}
		nextLevel = append(nextLevel, *types.NewRefs(addr, children))
	}
	for len(nextLevel) != 0 {
		other = append(other, nextLevel...)
		thisLevel = nextLevel
		nextLevel = nil
		for _, r := range thisLevel {
			for _, addr := range r.Referrals {
				data, err = k.Get(ctx, addr)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot obtain %s data", addr)
				}
				if data.CompressionAt != nil {
					compressions = append(compressions, *types.NewCompression(addr, *data.CompressionAt))
				}
				if data.StatusDowngradeAt != nil {
					downgrades = append(downgrades, *types.NewDowngrade(addr, data.Status, *data.StatusDowngradeAt))
				}
				if data.Transition != "" {
					transitions = append(transitions, types.Transition{
						Subject:     addr,
						Destination: data.Transition,
					})
				}
				children, err = k.GetChildren(ctx, addr)
				if err != nil {
					return nil, err
				}
				if len(children) == 0 {
					continue
				}
				nextLevel = append(nextLevel, *types.NewRefs(addr, children))
			}
		}
	}

	return types.NewGenesisState(params, topLevel, other, banished, neverPaid, compressions, downgrades, transitions), nil
}

func (k Keeper) ImportFromGenesis(
	ctx sdk.Context,
	topLevel []string,
	otherAccounts []types.Refs,
	banished []types.Banished,
	neverPaid []string,
	compressions []types.Compression,
	downgrades []types.Downgrade,
	transitions []types.Transition,
) error {
	store := ctx.KVStore(k.storeKey)

	neverPaidSet := make(map[string]bool, len(neverPaid))
	for _, acc := range neverPaid {
		neverPaidSet[acc] = true
	}

	k.Logger(ctx).Info("... top level accounts")
	for _, acc := range topLevel {
		if err := k.AddTopLevelAccount(ctx, acc); err != nil {
			panic(errors.Wrapf(err, "cannot add %s", acc))
		}
		if !neverPaidSet[acc] {
			key := []byte(acc)
			var info types.Info
			k.cdc.MustUnmarshalBinaryBare(store.Get(key), &info)
			info.NeverPaid = false
			store.Set(key, k.cdc.MustMarshalBinaryBare(&info))

		}
		k.Logger(ctx).Debug("account added", "acc", acc, "parent", nil)
	}
	k.Logger(ctx).Info("... other accounts")
	for _, r := range otherAccounts {
		for _, acc := range r.Referrals {
			if err := k.appendChild(ctx, r.Referrer, acc, true); err != nil {
				panic(errors.Wrapf(err, "cannot add %s", acc))
			}
			if !neverPaidSet[acc] {
				key := []byte(acc)
				var info types.Info
				k.cdc.MustUnmarshalBinaryBare(store.Get(key), &info)
				info.NeverPaid = false
				store.Set(key, k.cdc.MustMarshalBinaryBare(&info))

			}
			k.Logger(ctx).Debug("account added", "acc", acc, "parent", r.Referrer)
		}
	}
	{
		k.Logger(ctx).Info("... banished accounts")

		for _, ba := range banished {
			store.Set(
				[]byte(ba.Account),
				k.cdc.MustMarshalBinaryBare(&types.Info{
					Referrer:     ba.FormerReferrer,
					Coins:        []sdk.Int{k.getBalance(ctx, ba.Account)},
					Delegated:    []sdk.Int{k.getDelegated(ctx, ba.Account)},
					Banished:     true,
					NeverPaid:    neverPaidSet[ba.Account],
				}),
			)
		}
	}

	bu := newBunchUpdater(k, ctx)
	k.Logger(ctx).Info("... compressions")
	for _, x := range compressions {
		if err := bu.update(x.Account, false, func(value *types.Info) error {
			value.CompressionAt = &x.Time
			return nil
		}); err != nil {
			return err
		}
	}
	k.Logger(ctx).Info("... status downgrades")
	for _, x := range downgrades {
		if err := bu.update(x.Account, false, func(value *types.Info) error {
			k.Logger(ctx).Debug("status downgrade", "acc", x.Account, "from", x.Current, "to", value.Status)
			k.setStatus(ctx, value, x.Current, x.Account)
			value.StatusDowngradeAt = &x.Time
			return nil
		}); err != nil {
			return err
		}
	}
	k.Logger(ctx).Info("... transitions")
	for _, trans := range transitions {
		if err := bu.update(trans.Subject, false, func(value *types.Info) error {
			value.Transition = trans.Destination
			return nil
		}); err != nil {
			return err
		}
	}
	k.Logger(ctx).Info("... persisting")
	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}
