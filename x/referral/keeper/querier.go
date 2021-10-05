package keeper

import (
	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/referral/types"
)

// NewQuerier creates a new querier for referral clients.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryCoinsInNetwork:
			return queryCoins(ctx, req, k, legacyQuerierCdc)
		case types.QueryCheckStatus:
			return queryCheckStatus(ctx, req, k, legacyQuerierCdc)
		case types.QueryValidateTransition:
			return queryValidateTransition(ctx, req, k, legacyQuerierCdc)
		case types.QueryParams:
			return queryParams(ctx, k, legacyQuerierCdc)
		case types.QueryInfo:
			return queryInfo(ctx, req, k, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown referral query endpoint")
		}
	}
}

func queryCoins(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.CoinsRequest
	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	maxDepth := int(params.MaxDepth)

	//TODO: Refactor keeper to merge these calls in one
	n, err := k.GetCoinsInNetwork(ctx, params.AccAddress, maxDepth)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain total coins")
	}
	d, err := k.GetDelegatedInNetwork(ctx, params.AccAddress, maxDepth)
	if err != nil {
		return nil, errors.Wrap(err, "cannot obtain delegated coins")
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, types.CoinsResponse{
		Total:     n,
		Delegated: d,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryCheckStatus(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.CheckStatusRequest
	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	result, err := k.AreStatusRequirementsFulfilled(ctx, params.AccAddress, params.Status)
	if err != nil {
		return nil, err
	}
	resBytes, err := codec.MarshalJSONIndent(legacyQuerierCdc, result)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return resBytes, nil
}

func queryValidateTransition(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.ValidateTransitionRequest
	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	var response types.ValidateTransitionResponse
	if err := k.validateTransition(ctx, params.Subject, params.Target, true); err != nil {
		response = types.ValidateTransitionResponse{Error: err.Error()}
	} else {
		response = types.ValidateTransitionResponse{Ok: true}
	}
	json, err := codec.MarshalJSONIndent(legacyQuerierCdc, response)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return json, nil
}

func queryParams(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryInfo(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.GetRequest
	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	data, err := k.Get(ctx, params.AccAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	json, err := codec.MarshalJSONIndent(legacyQuerierCdc, data)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return json, nil
}
