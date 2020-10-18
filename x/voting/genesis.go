package voting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math"

	"github.com/arterynetwork/artr/x/voting/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.SetParams(ctx, data.Params)
	k.SetGovernment(ctx, data.Government)
	if data.CurrentProposal.TypeCode != types.ProposalTypeNone {
		k.SetCurrentProposal(ctx, data.CurrentProposal)
		k.SetStartBlock(ctx.WithBlockHeight(data.StartBlock))
	}
	if len(data.Agreed) > 0 {
		k.SetAgreed(ctx, data.Agreed)
	}
	if len(data.Disagreed) > 0 {
		k.SetDisagreed(ctx, data.Disagreed)
	}
	for _, record := range data.History {
		k.AddProposalHistoryRecord(ctx, record)
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	var currentProposal types.Proposal
	var startBlock int64
	if pp := k.GetCurrentProposal(ctx); pp != nil {
		currentProposal = *pp
		startBlock = k.GetStartBlock(ctx)
	}
	return NewGenesisState(
		k.GetParams(ctx),
		k.GetGovernment(ctx),
		currentProposal,
		startBlock,
		k.GetAgreed(ctx),
		k.GetDisagreed(ctx),
		k.GetHistory(ctx, math.MaxInt32, 1),
	)
}
