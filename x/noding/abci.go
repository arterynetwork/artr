package noding

import (
	"github.com/arterynetwork/artr/x/noding/types"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	if err := payProposerReward(ctx, req.Header.ProposerAddress, k); err != nil {
		k.Logger(ctx).Error(
			"Couldn't pay proposer reward",
			"address", req.Header.ProposerAddress,
			"error", err,
		)
	}
	if err := markStrokesAndTicks(ctx, req.LastCommitInfo.Votes, k); err != nil {
		k.Logger(ctx).Error(
			"Couldn't update statistics",
			"votes", req.LastCommitInfo.Votes,
			"error", err,
		)
	}
	if err := punishWrongdoers(ctx, req.ByzantineValidators, k); err != nil {
		k.Logger(ctx).Error(
			"Byzantine behavior detected",
			"evidences", req.ByzantineValidators,
			"error", err,
		)
	}
}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	updz, err := k.GatherValidatorUpdates(ctx)
	if err != nil { panic(err) }
	return updz
}

func findValidatorAccAddress(ctx sdk.Context, k Keeper, validator abci.Validator) (sdk.AccAddress, error) {
	consAddr := sdk.ConsAddress(validator.Address)
	accAddr, found, _, err := k.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil { return accAddr, sdkerrors.Wrap(err, "couldn't find validator") }
	if !found { return nil, errors.New("validator not found for consensus address " + consAddr.String())}
	return accAddr, nil
}

// punishWrongdoers - records infractions to the store and ban validators if needed
func punishWrongdoers(ctx sdk.Context, evz []abci.Evidence, k Keeper) error {
	for _, ev := range evz {
		accAddr, err := findValidatorAccAddress(ctx, k, ev.Validator)
		if err != nil { return err }
		err = k.MarkByzantine(ctx, accAddr, ev)
		if err != nil { return err }
	}
	return nil
}

// markStrokesAndTicks - increments signed/missed block counter and jail validators if needed
func markStrokesAndTicks(ctx sdk.Context, votes []abci.VoteInfo, k Keeper) error {
	for _, vote := range votes {
		accAddr, err := findValidatorAccAddress(ctx, k, vote.Validator)
		if err != nil { return err }
		if vote.SignedLastBlock {
			err = k.MarkTick(ctx, accAddr)
		} else {
			err = k.MarkStroke(ctx, accAddr)
		}
		if err != nil { return sdkerrors.Wrap(err, "cannot count a block for account " + accAddr.String()) }
	}
	return nil
}

func payProposerReward(ctx sdk.Context, consAddr sdk.ConsAddress, k Keeper) error {
	accAddr, found, _, err := k.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil { return err }
	if !found { return types.ErrNotFound }
	if err = k.PayProposerReward(ctx, accAddr); err != nil { return err }
	return nil
}