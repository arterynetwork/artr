package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSetProfile{}, "profile/SetProfile", nil)
	cdc.RegisterConcrete(MsgSetNickname{}, "profile/SetNickname", nil)
	cdc.RegisterConcrete(MsgSetCardNumber{}, "profile/SetCardNumber", nil)
	cdc.RegisterConcrete(MsgCreateAccount{}, "profile/CreateAccount", nil)
	cdc.RegisterConcrete(MsgCreateAccountWithProfile{}, "profile/CreateAccountWithProfile", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
