package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a new querier for schedule clients.
func NewQuerier(k Keeper, mrshl *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		//case types.QueryTasks:
		//	return queryTasks(ctx, req, k, mrshl)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown schedule query endpoint")
		}
	}
}

//func queryTasks(ctx sdk.Context, req abci.RequestQuery, k Keeper, mrshl *codec.LegacyAmino) ([]byte, error) {
//	var params types.QueryTasksParams
//
//	if err := mrshl.UnmarshalJSON(req.Data, &params); err != nil {
//		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
//	}
//
//	tasks := k.GetTasks(ctx, uint64(params.BlockHeight))
//
//	bz, err := codec.MarshalJSONIndent(mrshl, tasks)
//	if err != nil {
//		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
//	}
//
//	return bz, nil
//}
