package keeper

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/noding/types"
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
		case types.QueryOperator:
			return queryOperator(ctx, k, path[1:])
		case types.QueryParams:
			return queryParams(ctx, k)
		case types.QuerySwitchedOn:
			return querySwitchedOn(ctx, k)
		case types.QueryState:
			return queryState(ctx, k, path[1:])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown noding query endpoint")
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

func queryStatus(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "not enough arguments")
	}

	accAddress, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("cannot parse address: %s", path[0]))
	}

	data, err := k.IsValidator(ctx, accAddress)
	if err != nil {
		return nil, err
	}

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
	if err != nil {
		return nil, err
	}

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
		if err != nil {
			return nil, err
		}
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

func queryOperator(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	if len(path) < 2 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "not enough arguments")
	}

	var (
		consAddress sdk.ConsAddress
		err         error
	)
	if path[0] == types.QueryOperatorFormatHex {
		consAddress, err = sdk.ConsAddressFromHex(path[1])
	} else {
		consAddress, err = sdk.ConsAddressFromBech32(path[1])
	}
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("cannot parse address (%s): %s", path[0], path[1]))
	}

	data, found := k.getNodeOperatorFromIndex(ctx, consAddress)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNotFound, "cannot find data by consensus address: %s", consAddress.String())
	}

	return data, nil
}

func querySwitchedOn(ctx sdk.Context, k Keeper) ([]byte, error) {
	list, err := k.GetActiveValidatorList(ctx)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, list)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryState(ctx sdk.Context, k Keeper, path []string) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "not enough arguments")
	}

	accAddress, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("cannot parse address: %s", path[0]))
	}

	data := k.GetValidatorState(ctx, accAddress)
	return []byte{byte(data)}, nil
}
