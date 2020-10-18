package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/x/vpn/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a new querier for vpn clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryVpnState:
			return queryVpnState(ctx, req, k)
		case types.QueryVpnLimit:
			return queryVpnLimit(ctx, req, k)
		case types.QueryVpnCurrent:
			return queryVpnCurrent(ctx, req, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown vpn query endpoint")
		}
	}
}

func queryVpnStateBase(ctx sdk.Context, req abci.RequestQuery, k Keeper) (types.VpnInfo, error) {
	var params types.QueryVpnParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return types.VpnInfo{}, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	info, err := k.GetInfo(ctx, params.Address)
	if err != nil {
		info = types.VpnInfo{}
	}

	return info, nil
}

func queryVpnState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	info, err := queryVpnStateBase(ctx, req, k)

	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResState{State: info})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryVpnLimit(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	info, err := queryVpnStateBase(ctx, req, k)

	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResLimit{Limit: info.Limit})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryVpnCurrent(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	info, err := queryVpnStateBase(ctx, req, k)

	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResCurrent{Current: info.Current})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
