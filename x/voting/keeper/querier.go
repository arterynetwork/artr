package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/x/voting/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a new querier for voting clients.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, k, legacyQuerierCdc)
		case types.QueryGovernment:
			return queryGovernment(ctx, k, legacyQuerierCdc)
		case types.QueryCurrent:
			return queryCurrent(ctx, k, legacyQuerierCdc)
		case types.QueryHistory:
			return queryHistory(ctx, k, req, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown voting query endpoint")
		}
	}
}

func queryHistory(ctx sdk.Context, k Keeper, req abci.RequestQuery, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.HistoryRequest

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, k.GetHistory(ctx, params.Limit, params.Page))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryParams(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryGovernment(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	gov := k.GetGovernment(ctx)

	res, err := codec.MarshalJSONIndent(
		legacyQuerierCdc,
		types.GovernmentResponse{
			Members: gov.Members,
		},
	)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryCurrent(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	proposal := k.GetCurrentProposal(ctx)

	var (
		res []byte
		err error
	)

	if proposal == nil {
		res, err = codec.MarshalJSONIndent(legacyQuerierCdc, nil)
	} else {
		gov := k.GetGovernment(ctx)
		agreed := k.GetAgreed(ctx)
		disagreed := k.GetDisagreed(ctx)
		res, err = codec.MarshalJSONIndent(
			legacyQuerierCdc,
			types.CurrentResponse{
				Proposal:   *proposal,
				Government: gov.Members,
				Agreed:     agreed.Members,
				Disagreed:  disagreed.Members,
			},
		)
	}

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
