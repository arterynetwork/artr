package keeper

import (
	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier creates a new querier for profile clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryProfile:
			return queryProfile(ctx, req, k)
		case types.QueryCreators:
			return queryCreators(ctx, req, k)
		case types.QueryAccountAddressByNickname:
			return queryAccountByNickname(ctx, req, k)
		case types.QueryAccountAddressByCardNumber:
			return queryAccountByCardNumber(ctx, req, k)
		case types.QueryParams:
			return queryParams(ctx, k)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown profile query endpoint")
		}
	}
}

func queryCreators(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	moduleParams := k.GetParams(ctx)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryCreatorsRes{Creators: moduleParams.Creators})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryProfile(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryProfileParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	profile := k.GetProfile(ctx, params.Address)
	if profile == nil {
		profile = &types.Profile{}
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResProfile{Profile: *profile})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAccountByNickname(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAccountByNicknameParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	addr := k.GetProfileAccountByNickname(ctx, params.Nickname)
	if addr == nil {
		addr = sdk.AccAddress{}
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResAccountBy{Address: addr})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAccountByCardNumber(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAccountByCardNumberParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	addr := k.GetProfileAccountByCardNumber(ctx, params.CardNumber)
	if addr == nil {
		addr = sdk.AccAddress{}
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, types.QueryResAccountBy{Address: addr})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
