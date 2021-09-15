package keeper

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/referral/types"
)

func (k Keeper) ExportToGenesis(ctx sdk.Context) (types.GenesisState, error) {
	var (
		data         types.R
		err          error
		params       types.Params
		topLevel     []sdk.AccAddress
		other        []types.Refs
		banished     []types.GenesisBanishedOne
		compressions []types.GenesisCompression
		downgrades   []types.GenesisStatusDowngrade
		transitions  []types.Transition

		children  []sdk.AccAddress
		thisLevel []types.Refs
		nextLevel []types.Refs
	)
	params = k.GetParams(ctx)
	topLevel, banished, err = k.GetTopLevelAndBanishedAccounts(ctx)
	if err != nil {
		return types.GenesisState{}, err
	}

	for _, addr := range topLevel {
		data, err = k.Get(ctx, addr)
		if err != nil {
			return types.GenesisState{}, err
		}
		if data.CompressionAt != -1 {
			compressions = append(compressions, types.NewGenesisCompression(addr, data.CompressionAt))
		}
		if data.StatusDowngradeAt != -1 {
			downgrades = append(downgrades, types.NewGenesisStatusDowngrade(addr, data.Status, data.StatusDowngradeAt))
		}
		if !data.Transition.Empty() {
			transitions = append(transitions, types.NewTransition(addr, data.Transition))
		}
		children, err = k.GetChildren(ctx, addr)
		if err != nil {
			return types.GenesisState{}, err
		}
		if len(children) == 0 {
			continue
		}
		nextLevel = append(nextLevel, types.Refs{addr, children})
	}
	for len(nextLevel) != 0 {
		other = append(other, nextLevel...)
		thisLevel = nextLevel
		nextLevel = nil
		for _, r := range thisLevel {
			for _, addr := range r.Referrals {
				data, err = k.Get(ctx, addr)
				if err != nil {
					return types.GenesisState{}, err
				}
				if data.CompressionAt != -1 {
					compressions = append(compressions, types.NewGenesisCompression(addr, data.CompressionAt))
				}
				if data.StatusDowngradeAt != -1 {
					downgrades = append(downgrades, types.NewGenesisStatusDowngrade(addr, data.Status, data.StatusDowngradeAt))
				}
				if !data.Transition.Empty() {
					transitions = append(transitions, types.NewTransition(addr, data.Transition))
				}
				children, err = k.GetChildren(ctx, addr)
				if err != nil {
					return types.GenesisState{}, err
				}
				if len(children) == 0 {
					continue
				}
				nextLevel = append(nextLevel, types.Refs{addr, children})
			}
		}
	}

	return types.NewGenesisState(params, topLevel, other, banished, compressions, downgrades, transitions), nil
}

func (k Keeper) ImportFromGenesis(
	ctx sdk.Context,
	topLevel []sdk.AccAddress,
	otherAccounts []types.Refs,
	banishedAccounts []types.GenesisBanishedOne,
	compressions []types.GenesisCompression,
	downgrades []types.GenesisStatusDowngrade,
	transitions []types.Transition,
) error {
	for _, acc := range topLevel {
		err := k.AddTopLevelAccount(ctx, acc)
		if err != nil {
			panic(errors.Wrapf(err, "cannot add %s", acc))
		}
		k.Logger(ctx).Debug("account added", "acc", acc, "parent", nil)
	}
	for _, r := range otherAccounts {
		for _, acc := range r.Referrals {
			err := k.appendChild(ctx, r.Referrer, acc, true)
			if err != nil {
				panic(errors.Wrapf(err, "cannot add %s", acc))
			}
			k.Logger(ctx).Debug("account added", "acc", acc, "parent", r.Referrer)
		}
	}

	{
		store := ctx.KVStore(k.storeKey)
		for _, ba := range banishedAccounts {
			store.Set(
				ba.Account,
				k.cdc.MustMarshalBinaryLengthPrefixed(types.R{
					Banished:          true,
					BanishmentAt:      ba.Height,
					Referrer:          ba.FormerReferrer,
					StatusDowngradeAt: -1,
					CompressionAt:     -1,
					Coins: [11]sdk.Int{
						k.getBalance(ctx, ba.Account),
						sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
						sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
					},
					Delegated: [11]sdk.Int{
						k.getDelegated(ctx, ba.Account),
						sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
						sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(),
					},
				}),
			)
			k.Logger(ctx).Debug("dead account added", "acc", ba.Account, "former_parent", ba.FormerReferrer)
		}
	}

	bu := newBunchUpdater(k, ctx)
	for _, x := range compressions {
		if err := bu.update(x.Account, false, func(value *types.R) {
			value.CompressionAt = x.Height
		}); err != nil {
			return err
		}
	}
	for _, x := range downgrades {
		if err := bu.update(x.Account, false, func(value *types.R) {
			k.setStatus(ctx, value, types.Status(x.Current), x.Account)
			value.StatusDowngradeAt = x.Height
		}); err != nil {
			return err
		}
	}
	for _, x := range transitions {
		if err := bu.update(x.Subject, false, func(value *types.R) {
			value.Transition = x.Destination
		}); err != nil {
			return err
		}
	}
	if err := bu.commit(); err != nil {
		return err
	}
	return nil
}
