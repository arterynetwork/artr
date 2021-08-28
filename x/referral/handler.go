package referral

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/referral/keeper"
	"github.com/arterynetwork/artr/x/referral/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		srv := keeper.NewMsgServer(k)
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		sdkCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgRequestTransition:
			res, err := srv.RequestTransition(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgResolveTransition:
			res, err := srv.ResolveTransition(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
