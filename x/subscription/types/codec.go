package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPaySubscription{}, "profile/PaySubscription", nil)
	cdc.RegisterConcrete(MsgPayVPN{}, "profile/PayVPN", nil)
	cdc.RegisterConcrete(MsgPayStorage{}, "profile/PayStorage", nil)
	cdc.RegisterConcrete(MsgSetTokenRate{}, "profile/SetTokenCourse", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
