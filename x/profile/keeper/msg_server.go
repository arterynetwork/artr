package keeper

import (
	"context"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile/types"
)

type MsgServer Keeper

var _ types.MsgServer = MsgServer{}

func (s MsgServer) CreateAccount(ctx context.Context, msg *types.MsgCreateAccount) (*types.MsgCreateAccountResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)

	if acc := k.accountKeeper.GetAccount(sdkCtx, msg.GetAddress()); acc != nil {
		return nil, types.ErrAccountAlreadyExists
	}

	p := k.GetParams(sdkCtx)
	if p.RenamePrice > 0 {
		freeCreation := false
		for _, acc := range p.Creators {
			if msg.Creator == acc {
				freeCreation = true
				break
			}
		}

		if !freeCreation {
			if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, msg.GetCreator(), auth.FeeCollectorName, util.Uartrs(p.RenamePrice)); err != nil {
				return nil, errors.Wrap(err, "cannot charge a fee")
			}
		}
	}

	if msg.WithProfile() {
		if err := Keeper(s).CreateAccountWithProfile(sdkCtx, msg.GetAddress(), msg.GetReferrer(), *msg.Profile); err != nil {
			return nil, errors.Wrap(err, "cannot create an account with profile")
		}
	} else {
		if err := Keeper(s).CreateAccount(sdkCtx, msg.GetAddress(), msg.GetReferrer()); err != nil {
			return nil, errors.Wrap(err, "cannot create an account")
		}
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgCreateAccountResponse{}, nil
}

func (s MsgServer) UpdateProfile(ctx context.Context, msg *types.MsgUpdateProfile) (*types.MsgUpdateProfileResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	addr := msg.GetAddress()
	profile := k.GetProfile(sdkCtx, addr)

	for _, upd := range msg.Updates {
		switch upd.Field {
		case types.MsgUpdateProfile_Update_FIELD_AUTO_PAY:
			profile.AutoPay = upd.GetBool()
		case types.MsgUpdateProfile_Update_FIELD_NODING:
			profile.Noding = upd.GetBool()
		case types.MsgUpdateProfile_Update_FIELD_STORAGE:
			profile.Storage = upd.GetBool()
		case types.MsgUpdateProfile_Update_FIELD_VALIDATOR:
			profile.Validator = upd.GetBool()
		case types.MsgUpdateProfile_Update_FIELD_VPN:
			profile.Vpn = upd.GetBool()
		case types.MsgUpdateProfile_Update_FIELD_NICKNAME:
			profile.Nickname = upd.GetString_()
		case types.MsgUpdateProfile_Update_FIELD_IM_AUTO_PAY:
			profile.AutoPayImExtra = upd.GetBool()
		}
	}

	if err := k.SetProfile(sdkCtx, msg.GetAddress(), *profile); err != nil {
		return nil, errors.Wrap(err, "cannot set profile")
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgUpdateProfileResponse{}, nil
}

func (s MsgServer) SetStorageCurrent(ctx context.Context, msg *types.MsgSetStorageCurrent) (*types.MsgSetStorageCurrentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	p := k.GetParams(sdkCtx)
	authorized := false
	for _, acc := range p.StorageSigners {
		if msg.Sender == acc {
			authorized = true
			break
		}
	}
	if !authorized {
		return nil, types.ErrUnauthorized
	}

	if err := k.SetStorageCurrent(sdkCtx, msg.GetAddress(), msg.Value); err != nil {
		return nil, errors.Wrap(err, "cannot set value")
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgSetStorageCurrentResponse{}, nil
}

func (s MsgServer) SetVpnCurrent(ctx context.Context, msg *types.MsgSetVpnCurrent) (*types.MsgSetVpnCurrentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	p := k.GetParams(sdkCtx)

	authorized := false
	for _, acc := range p.VpnSigners {
		if msg.Sender == acc {
			authorized = true
			break
		}
	}
	if !authorized {
		return nil, types.ErrUnauthorized
	}

	if err := k.SetVpnCurrent(sdkCtx, msg.GetAddress(), msg.Value); err != nil {
		return nil, errors.Wrap(err, "cannot set value")
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgSetVpnCurrentResponse{}, nil
}

func (s MsgServer) PayTariff(ctx context.Context, msg *types.MsgPayTariff) (*types.MsgPayTariffResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.PayTariff(sdkCtx, msg.GetAddress(), msg.StorageAmount, false); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgPayTariffResponse{}, nil
}

func (s MsgServer) BuyStorage(ctx context.Context, msg *types.MsgBuyStorage) (*types.MsgBuyStorageResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.BuyStorage(sdkCtx, msg.GetAddress(), msg.ExtraStorage); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgBuyStorageResponse{}, nil
}

func (s MsgServer) GiveStorageUp(ctx context.Context, msg *types.MsgGiveStorageUp) (*types.MsgGiveStorageUpResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.GiveStorageUp(sdkCtx, msg.GetAccAddress(), msg.Amount); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgGiveStorageUpResponse{}, nil
}

func (s MsgServer) BuyVpn(ctx context.Context, msg *types.MsgBuyVpn) (*types.MsgBuyVpnResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.BuyVpn(sdkCtx, msg.GetAddress(), msg.ExtraTraffic); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgBuyVpnResponse{}, nil
}

func (s MsgServer) SetRate(ctx context.Context, msg *types.MsgSetRate) (*types.MsgSetRateResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	p := k.GetParams(sdkCtx)

	authorized := false
	for _, acc := range p.TokenRateSigners {
		if msg.Sender == acc {
			authorized = true
			break
		}
	}
	if !authorized {
		return nil, types.ErrUnauthorized
	}

	p.TokenRate = msg.Value
	k.SetParams(sdkCtx, p)
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgSetRateResponse{}, nil
}

func (s MsgServer) BuyImExtraStorage(ctx context.Context, msg *types.MsgBuyImExtraStorage) (*types.MsgBuyImExtraStorageResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.BuyImStorage(sdkCtx, msg.GetAddress(), msg.ExtraStorage); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgBuyImExtraStorageResponse{}, nil
}

func (s MsgServer) GiveUpImExtra(ctx context.Context, msg *types.MsgGiveUpImExtra) (*types.MsgGiveUpImExtraResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.GiveImStorageUp(sdkCtx, msg.GetAddress(), msg.Amount); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgGiveUpImExtraResponse{}, nil
}

func (s MsgServer) ProlongImExtra(ctx context.Context, msg *types.MsgProlongImExtra) (*types.MsgProlongImExtraResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := Keeper(s)
	if err := k.ProlongImExtra(sdkCtx, msg.GetAddress()); err != nil {
		return nil, err
	}
	util.TagTx(sdkCtx, types.ModuleName, msg)
	return &types.MsgProlongImExtraResponse{}, nil
}
