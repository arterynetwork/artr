package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/x/voting/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a new querier for voting clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, k)
		case types.QueryGovernment:
			return queryGovernment(ctx, k)
		case types.QueryCurrent:
			return queryCurrent(ctx, k)
		case types.QueryStatus:
			return queryStatus(ctx, k)
		case types.QueryHistory:
			return queryHistory(ctx, k, req)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown voting query endpoint")
		}
	}
}

func queryHistory(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.QueryHistoryParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, k.GetHistory(ctx, params.Limit, params.Page))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryGovernment(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetGovernment(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, types.NewQueryGovernmentRes(params))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryCurrent(ctx sdk.Context, k Keeper) ([]byte, error) {
	proposal := k.GetCurrentProposal(ctx)

	var (
		res []byte
		err error
	)

	if proposal == nil {
		res, err = codec.MarshalJSONIndent(types.ModuleCdc, nil)
	} else {
		res, err = codec.MarshalJSONIndent(types.ModuleCdc, types.NewQueryCurrentRes(*proposal))
	}

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryStatus(ctx sdk.Context, k Keeper) ([]byte, error) {
	proposal := k.GetCurrentProposal(ctx)

	var (
		res []byte
		err error
	)

	if proposal == nil {
		res, err = codec.MarshalJSONIndent(types.ModuleCdc, nil)
	} else {
		res, err = codec.MarshalJSONIndent(types.ModuleCdc, types.NewQueryStatusRes(*proposal,
			k.GetGovernment(ctx),
			k.GetAgreed(ctx),
			k.GetDisagreed(ctx),
		))
	}

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
