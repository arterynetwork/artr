package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(MsgUpdateProfile{}, "profile/UpdateProfile", nil)
	cdc.RegisterConcrete(MsgCreateAccount{}, "profile/CreateAccount", nil)
	cdc.RegisterConcrete(MsgSetStorageCurrent{}, "profile/SetStorageCurrent", nil)
	cdc.RegisterConcrete(MsgSetVpnCurrent{}, "profile/SetVpnCurrent", nil)
	cdc.RegisterConcrete(MsgPayTariff{}, "profile/PayTariff", nil)
	cdc.RegisterConcrete(MsgBuyStorage{}, "profile/BuyStorage", nil)
	cdc.RegisterConcrete(MsgGiveStorageUp{}, "profile/GiveStorageUp", nil)
	cdc.RegisterConcrete(MsgBuyVpn{}, "profile/BuyVpn", nil)
	cdc.RegisterConcrete(MsgSetRate{}, "profile/SetRate", nil)
	cdc.RegisterConcrete(MsgBuyImExtraStorage{}, "profile/BuyImExtraStorage", nil)
	cdc.RegisterConcrete(MsgGiveUpImExtra{}, "profile/GiveUpImExtra", nil)
	cdc.RegisterConcrete(MsgProlongImExtra{}, "profile/ProlongImExtra", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateProfile{},
		&MsgCreateAccount{},
		&MsgSetStorageCurrent{},
		&MsgSetVpnCurrent{},
		&MsgPayTariff{},
		&MsgBuyStorage{},
		&MsgGiveStorageUp{},
		&MsgBuyVpn{},
		&MsgSetRate{},
		&MsgBuyImExtraStorage{},
		&MsgGiveUpImExtra{},
		&MsgProlongImExtra{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// ModuleCdc defines the module codec
var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
