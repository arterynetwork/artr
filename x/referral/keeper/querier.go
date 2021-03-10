package keeper

import (
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/referral/types"
)

// NewQuerier creates a new querier for referral clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryStatus:
			return queryStatus(ctx, path[1:], k)
		case types.QueryReferrer:
			return queryParent(ctx, path[1:], k)
		case types.QueryReferrals:
			return queryChildren(ctx, path[1:], k)
		case types.QueryCoinsInNetwork:
			return queryCoins(ctx, path[1:], k)
		case types.QueryDelegatedInNetwork:
			return queryDelegated(ctx, path[1:], k)
		case types.QueryCheckStatus:
			return queryCheckStatus(ctx, path[1:], k)
		case types.QueryWhenCompression:
			return queryCompressionTime(ctx, path[1:], k)
		case types.QueryPendingTransition:
			return queryPendingTransition(ctx, path[1:], k)
		case types.QueryValidateTransition:
			return queryValidateTransition(ctx, path[1:], k)
		case types.QueryParams:
			return queryParams(ctx, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown referral query endpoint")
		}
	}
}

func queryStatus(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	status, err := k.GetStatus(ctx, addr)
	if err != nil {
		return nil, err
	}
	res, err := codec.MarshalJSONIndent(k.cdc, status)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryParent(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	parent, err := k.GetParent(ctx, addr)
	if err != nil {
		return nil, err
	}
	res, err := codec.MarshalJSONIndent(k.cdc, parent)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryChildren(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	children, err := k.GetChildren(ctx, addr)
	if err != nil {
		return nil, err
	}
	res, err := codec.MarshalJSONIndent(k.cdc, types.QueryResChildren(children))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryCoins(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	var d = 10
	if len(path) > 1 {
		d, err = strconv.Atoi(path[1])
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	coins, err := k.GetCoinsInNetwork(ctx, addr, d)
	if err != nil {
		return nil, err
	}
	res, err := codec.MarshalJSONIndent(k.cdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryDelegated(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	var d = 10
	if len(path) > 1 {
		d, err = strconv.Atoi(path[1])
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	coins, err := k.GetDelegatedInNetwork(ctx, addr, d)
	if err != nil {
		return nil, err
	}
	res, err := codec.MarshalJSONIndent(k.cdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryCheckStatus(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	var (
		addr     sdk.AccAddress
		status   int
		result   types.StatusCheckResult
		resBytes []byte
		err      error
	)
	addr, err = sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	status, err = strconv.Atoi(path[1])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	result, err = k.AreStatusRequirementsFulfilled(ctx, addr, types.Status(status))
	if err != nil {
		return nil, err
	}
	resBytes, err = codec.MarshalJSONIndent(k.cdc, result)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return resBytes, nil
}

func queryCompressionTime(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	h, err := k.GetCompressionBlockHeight(ctx, addr)
	if err != nil {
		return nil, err
	}
	res, err := codec.MarshalJSONIndent(k.cdc, h)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryPendingTransition(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	dest, err := k.GetPendingTransition(ctx, addr)
	return dest, err
}

func queryValidateTransition(ctx sdk.Context, path []string, k Keeper) ([]byte, error) {
	subj, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot parse subject address: "+err.Error())
	}
	dest, err := sdk.AccAddressFromBech32(path[1])
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot parse destination address: "+err.Error())
	}
	var res types.QueryResValidateTransition
	if err := k.ValidateTransition(ctx, subj, dest); err != nil {
		res = types.QueryResValidateTransition{Ok: false, Err: err.Error()}
	} else {
		res = types.QueryResValidateTransition{Ok: true}
	}
	json, err := codec.MarshalJSONIndent(k.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return json, nil
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
