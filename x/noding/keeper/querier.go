package keeper

import (
	"github.com/arterynetwork/artr/x/noding/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"
)

// NewQuerier creates a new querier for noding clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		//case types.QueryParams:
		//	return queryParams(ctx, k)
		case types.QueryStatus:
			return queryStatus(ctx, k, path[1:])
		case types.QueryInfo:
			return queryInfo(ctx, k, path[1:])
		case types.QueryProposer:
			return queryProposer(ctx, k, path[1:])
		case types.QueryAllowed:
			return queryAllowed(ctx, k, path[1:])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown noding query endpoint")
		}
	}
}

//func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
//	params := k.GetParams(ctx)
//
//	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
//	if err != nil {
//		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
//	}
//
//	return res, nil
//}

func queryStatus(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "not enough arguments")
	}

	accAddress, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("cannot parse address: %s", path[0]))
	}

	data, err := k.IsValidator(ctx, accAddress)
	if err != nil {	return nil, err }

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, data)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryInfo(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "not enough arguments")
	}

	accAddress, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("cannot parse address: %s", path[0]))
	}

	data, err := k.Get(ctx, accAddress)
	if err != nil {	return nil, err }

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, data)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryProposer(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	var (
		height int64
		err    error
	)
	if len(path) == 0 {
		height = 0
	} else {
		height, err = strconv.ParseInt(path[0], 0, 64)
		if err != nil { return nil, err }
	}
	return k.GetBlockProposer(ctx, height)
}

func queryAllowed(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "not enough arguments")
	}

	accAddress, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("cannot parse address: %s", path[0]))
	}

	verdict, _, reason, err := k.IsQualified(ctx, accAddress)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, types.NewAllowedQueryRes(verdict, reason))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}