package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/arterynetwork/artr/x/bank/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// query balance path
	QueryBalance = "balances"
	QueryParams  = "params"
)

// NewQuerier returns a new sdk.Keeper instance.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryBalance:
			return queryBalance(ctx, req, k, legacyQuerierCdc)
		case QueryParams:
			return queryParams(ctx, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

// queryBalance fetch an account's balance for the supplied height.
// Height and account address are passed as first and second path components respectively.
func queryBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryBalanceParams

	if err := cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	coins := k.GetBalance(ctx, params.Address)

	bz, err := codec.MarshalJSONIndent(cdc, coins)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryParams(ctx sdk.Context, k Keeper, cdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)
	bz, err := codec.MarshalJSONIndent(cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
