package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/x/subscription/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a new querier for subscription clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryActivityInfo:
			return queryActivityInfo(ctx, k, req)
		case types.QueryPrices:
			return queryPrices(ctx, k, req)
		case types.QueryParams:
			return queryParams(ctx, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown subscription query endpoint "+path[0])
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// Query is account active or not
func queryActivityInfo(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.QueryActivityParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	info := k.GetActivityInfo(ctx, params.Address)

	activityRes := types.NewQueryActivityInfoRes(info, ctx.BlockHeight())

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, activityRes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// Query is account active or not
func queryPrices(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	course, subscription, vpn, storage, baseStorage, baseVpn := k.GetPrices(ctx)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc,
		types.QueryPricesRes{
			Subscription: int64(subscription) * int64(course),
			VPN:          int64(vpn) * int64(course),
			Storage:      int64(storage) * int64(course),
			Course:       int64(course),
			StorageGb:    int32(baseStorage),
			VPNGb:        int32(baseVpn),
		})

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
