package voting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/voting/keeper"
	"github.com/arterynetwork/artr/x/voting/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		srv := keeper.MsgServer(k)
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		sdkCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgPropose:
			res, err := srv.Propose(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgVote:
			res, err := srv.Vote(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
