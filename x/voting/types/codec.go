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
	// Messages
	cdc.RegisterConcrete(MsgPropose{}, ModuleName+"/CreateProposal", nil)
	cdc.RegisterConcrete(MsgVote{}, ModuleName+"/ProposalVote", nil)
	// Proposal Params
	cdc.RegisterConcrete(PriceArgs{}, ModuleName+"/PriceArgs", nil)
	cdc.RegisterConcrete(Proposal_Price{}, ModuleName+"/PriceArgsWrap", nil)
	cdc.RegisterConcrete(DelegationAwardArgs{}, ModuleName+"/DelegationAwardArgs", nil)
	cdc.RegisterConcrete(Proposal_DelegationAward{}, ModuleName+"/DelegationAwardArgsWrap", nil)
	cdc.RegisterConcrete(NetworkAwardArgs{}, ModuleName+"/NetworkAwardArgs", nil)
	cdc.RegisterConcrete(Proposal_NetworkAward{}, ModuleName+"/NetworkAwardArgsWrap", nil)
	cdc.RegisterConcrete(AddressArgs{}, ModuleName+"/AddressArgs", nil)
	cdc.RegisterConcrete(Proposal_Address{}, ModuleName+"/AddressArgsWrap", nil)
	cdc.RegisterConcrete(SoftwareUpgradeArgs{}, ModuleName+"/SoftwareUpgradeArgs", nil)
	cdc.RegisterConcrete(Proposal_SoftwareUpgrade{}, ModuleName+"/SoftwareUpgradeArgsWrap", nil)
	cdc.RegisterConcrete(MinAmountArgs{}, ModuleName+"/MinAmountArgs", nil)
	cdc.RegisterConcrete(Proposal_MinAmount{}, ModuleName+"/MinAmountArgsWrap", nil)
	cdc.RegisterConcrete(CountArgs{}, ModuleName+"/CountArgs", nil)
	cdc.RegisterConcrete(Proposal_Count{}, ModuleName+"/CountArgsWrap", nil)
	cdc.RegisterConcrete(StatusArgs{}, ModuleName+"/StatusArgs", nil)
	cdc.RegisterConcrete(Proposal_Status{}, ModuleName+"/StatusArgsWrap", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPropose{},
		&MsgVote{},
	)
	//TODO: Do we need to register isProposal_Args implementations here?

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
