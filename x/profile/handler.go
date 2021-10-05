package profile

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/profile/keeper"
	"github.com/arterynetwork/artr/x/profile/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		srv := keeper.MsgServer(k)
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		sdkCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgCreateAccount:
			res, err := srv.CreateAccount(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateProfile:
			res, err := srv.UpdateProfile(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSetStorageCurrent:
			res, err := srv.SetStorageCurrent(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSetVpnCurrent:
			res, err := srv.SetVpnCurrent(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgPayTariff:
			res, err := srv.PayTariff(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgBuyStorage:
			res, err := srv.BuyStorage(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgGiveStorageUp:
			res, err := srv.GiveStorageUp(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgBuyVpn:
			res, err := srv.BuyVpn(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSetRate:
			res, err := srv.SetRate(sdkCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
