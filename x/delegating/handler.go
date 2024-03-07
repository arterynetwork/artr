package delegating

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/delegating/keeper"
	"github.com/arterynetwork/artr/x/delegating/types"
)

func NewHandler(k Keeper) sdk.Handler {
	srv := keeper.NewMsgServer(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		sdkCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgDelegate:
			res, err := srv.Delegate(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRevoke:
			res, err := srv.Revoke(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgExpressRevoke:
			res, err := srv.ExpressRevoke(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
