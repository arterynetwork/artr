package keeper

import (
	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier creates a new querier for profile clients.
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryProfileByAddress:
			return queryProfile(ctx, req, k, legacyQuerierCdc)
		case types.QueryProfileByNickname:
			return queryAccountByNickname(ctx, req, k, legacyQuerierCdc)
		case types.QueryProfileByCardNumber:
			return queryAccountByCardNumber(ctx, req, k, legacyQuerierCdc)
		case types.QueryParams:
			return queryParams(ctx, k, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown profile query endpoint")
		}
	}
}

func queryProfile(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.GetByAddressRequest

	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	profile := k.GetProfile(ctx, params.GetAddress())
	if profile == nil {
		return nil, types.ErrNotFound
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, types.GetByAddressResponse{Profile: *profile})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAccountByNickname(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.GetByNicknameRequest

	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	addr := k.GetProfileAccountByNickname(ctx, params.Nickname)
	if addr == nil {
		return nil, types.ErrNotFound
	}
	profile := k.GetProfile(ctx, addr)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, types.GetByNicknameResponse{
		Address: addr.String(),
		Profile: *profile,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAccountByCardNumber(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.GetByCardNumberRequest

	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	addr := k.GetProfileAccountByCardNumber(ctx, params.CardNumber)
	if addr == nil {
		return nil, types.ErrNotFound
	}
	profile := k.GetProfile(ctx, addr)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, types.GetByCardNumberResponse{
		Address: addr.String(),
		Profile: *profile,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryParams(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, types.ParamsResponse{
		Params: params,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
