package earning

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/earning/keeper"
	"github.com/arterynetwork/artr/x/earning/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		srv := keeper.MsgServer(k)
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		sdkCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgListEarners:
			res, err := srv.ListEarners(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgReset:
			res, err := srv.Reset(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRun:
			res, err := srv.Run(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
