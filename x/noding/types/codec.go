package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"strings"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSwitchOn{}, strings.Join([]string{ModuleName, SwitchOnConst}, "/"), nil)
	cdc.RegisterConcrete(MsgSwitchOff{}, strings.Join([]string{ModuleName, SwitchOffConst}, "/"), nil)
	cdc.RegisterConcrete(AllowedQueryRes{}, "noding/AllowedQueryRes", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
