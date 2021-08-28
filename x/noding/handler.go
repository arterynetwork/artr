package noding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/noding/keeper"
	"github.com/arterynetwork/artr/x/noding/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		srv := keeper.MsgServer(k)
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		sdkCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgOn:
			res, err := srv.On(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgOff:
			res, err := srv.Off(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUnjail:
			res, err := srv.Unjail(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
