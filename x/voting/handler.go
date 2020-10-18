package voting

import (
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/voting/types"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the voting type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgCreateProposal:
			return handleMsgCreateProposal(ctx, k, msg)
		case types.MsgProposalVote:
			return handleMsgProposalVote(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// handleMsgCreateProposal handle proposal creation messages
func handleMsgCreateProposal(ctx sdk.Context, k Keeper, msg types.MsgCreateProposal) (*sdk.Result, error) {
	if k.GetCurrentProposal(ctx) != nil {
		return nil, types.ErrOtherActive
	}

	gov := k.GetGovernment(ctx)

	if !gov.Contains(msg.Author) {
		return nil, sdkerrors.Wrap(types.ErrSignerNotAllowed, msg.Author.String())
	}

	params := k.GetParams(ctx)
	endBLock := ctx.BlockHeight() + int64(params.VotingPeriod)

	switch msg.TypeCode {
	case types.ProposalTypeDelegationAward:
		p := msg.Params.(types.DelegationAwardProposalParams)
		err := delegating.NewPercentage(int(p.Minimal), int(p.ThousandPlus), int(p.TenKPlus), int(p.HundredKPlus)).Validate()
		if err != nil {
			return nil, err
		}
	case types.ProposalTypeDelegationNetworkAward, types.ProposalTypeProductNetworkAward:
		if err := msg.Params.(types.NetworkAwardProposalParams).Award.Validate(); err != nil {
			return nil, err
		}
	case types.ProposalTypeGovernmentAdd:
		if gov.Contains(msg.Params.(types.AddressProposalParams).Address) {
			return nil, types.ErrProposalGovernorExists
		}
	case types.ProposalTypeGovernmentRemove:
		if !gov.Contains(msg.Params.(types.AddressProposalParams).Address) {
			return nil, types.ErrProposalGovernorNotExists
		}
	}

	proposal := types.Proposal{
		Name:     msg.Name,
		TypeCode: msg.TypeCode,
		Params:   msg.Params,
		Author:   msg.Author,
		EndBlock: endBLock,
	}

	// Set proposal
	k.SetCurrentProposal(ctx, proposal)

	// Set empty lists of voters
	k.SetAgreed(ctx, types.Government{msg.Author})
	k.SetDisagreed(ctx, types.NewEmptyGovernment())
	k.ScheduleEnding(ctx, endBLock)
	k.SetStartBlock(ctx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateProposal,
			sdk.NewAttribute(types.AttributeKeyAuthor, msg.Author.String()),
			sdk.NewAttribute(types.AttributeKeyTypeCode, fmt.Sprint(msg.TypeCode)),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgCreateProposal handle proposal creation messages
func handleMsgProposalVote(ctx sdk.Context, k Keeper, msg types.MsgProposalVote) (*sdk.Result, error) {
	proposal := k.GetCurrentProposal(ctx)
	if proposal == nil {
		return nil, types.ErrNoActiveProposals
	}

	gov := k.GetGovernment(ctx)

	if !gov.Contains(msg.Voter) {
		return nil, sdkerrors.Wrap(types.ErrSignerNotAllowed, msg.Voter.String())
	}

	agreed := k.GetAgreed(ctx)

	if agreed.Contains(msg.Voter) {
		return nil, sdkerrors.Wrap(types.ErrAlreadyVoted, msg.Voter.String())
	}

	disagreed := k.GetDisagreed(ctx)

	if disagreed.Contains(msg.Voter) {
		return nil, sdkerrors.Wrap(types.ErrAlreadyVoted, msg.Voter.String())
	}

	if msg.Agree {
		agreed = agreed.Append(msg.Voter)
		k.SetAgreed(ctx, agreed)
	} else {
		disagreed = disagreed.Append(msg.Voter)
		k.SetDisagreed(ctx, disagreed)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyAuthor, msg.Voter.String()),
			sdk.NewAttribute(types.AttributeKeyAgree, fmt.Sprint(msg.Agree)),
		),
	)

	complete, agree := k.Validate(gov, agreed, disagreed)

	if complete {
		k.EndProposal(ctx, *proposal, agree)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
