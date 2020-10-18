package keeper

import (
	//"fmt"

	"github.com/arterynetwork/artr/x/referral/types"
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"

	//"github.com/cosmos/cosmos-sdk/client"
	//"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	//"github.com/arterynetwork/artr/x/referral/types"
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
	coins, err := k.GetCoinsInNetwork(ctx, addr)
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
	coins, err := k.GetDelegatedInNetwork(ctx, addr)
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
	if err != nil { return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error()) }
	status, err = strconv.Atoi(path[1])
	if err != nil {	return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error()) }
	result, err = k.AreStatusRequirementsFulfilled(ctx, addr, types.Status(status))
	if err != nil { return nil, err }
	resBytes, err = codec.MarshalJSONIndent(k.cdc, result)
	if err != nil { return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error()) }
	return resBytes, nil
}
