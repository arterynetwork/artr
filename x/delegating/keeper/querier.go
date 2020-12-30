package keeper

import (
	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier creates a new querier for delegating clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryRevoking:
			return queryRevoking(ctx, k, path[1:])
		case types.QueryAccumulation:
			return queryAccumulation(ctx, k, path[1:])
		case types.QueryParams:
			return queryParams(ctx, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown delegating query endpoint")
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

func queryRevoking(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	acc, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot parse account address")
	}
	data, err := k.GetRevoking(ctx, acc)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResRevoking(data))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryAccumulation(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	k.Logger(ctx).Debug("queryAccumulation", "path", path)
	acc, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		k.Logger(ctx).Error("cannot parse account address", "accAddr", path[0])
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot parse account address")
	}
	data, err := k.GetAccumulation(ctx, acc)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Debug("queryAccumulation", "data", data)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResAccumulation(data))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}
