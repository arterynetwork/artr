package keeper

import (
	"github.com/arterynetwork/artr/x/storage/types"
	b64 "encoding/base64"
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a new querier for storage clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryStorageData:
			return queryData(ctx, k, req)
		case types.QueryStorageInfo:
			return queryInfo(ctx, k, req)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown storage query endpoint")
		}
	}
}

// Query is account active or not
func queryData(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.QueryStorageParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	data := k.GetData(ctx, params.Address)

	res := types.NewQueryStorageDataRes(b64.StdEncoding.EncodeToString(data))

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// Query is account active or not
func queryInfo(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.QueryStorageParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	limit := k.GetLimit(ctx, params.Address)
	current := k.GetCurrent(ctx, params.Address)

	res := types.NewQueryStorageInfoRes(limit, current)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
