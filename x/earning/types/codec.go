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
	cdc.RegisterConcrete(MsgSet{}, "earning/set", nil)
	cdc.RegisterConcrete(MsgSetMultiple{}, "earning/setMultiple", nil)
	cdc.RegisterConcrete(MsgListEarners{}, "earning/listEarners", nil)
	cdc.RegisterConcrete(MsgRun{}, "earning/run", nil)
	cdc.RegisterConcrete(MsgReset{}, "earning/reset", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSet{},
		&MsgSetMultiple{},
		&MsgListEarners{},
		&MsgRun{},
		&MsgReset{},
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
