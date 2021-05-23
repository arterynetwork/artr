package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	// Interfaces
	cdc.RegisterInterface((*ProposalParams)(nil), nil)
	// Messages
	cdc.RegisterConcrete(MsgCreateProposal{}, ModuleName+"/CreateProposal", nil)
	cdc.RegisterConcrete(MsgProposalVote{}, ModuleName+"/ProposalVote", nil)
	// Proposal Params
	cdc.RegisterConcrete(EmptyProposalParams{}, ModuleName+"/EmptyProposalParams", nil)
	cdc.RegisterConcrete(PriceProposalParams{}, ModuleName+"/PriceProposalParams", nil)
	cdc.RegisterConcrete(DelegationAwardProposalParams{}, ModuleName+"/DelegationAwardProposalParams", nil)
	cdc.RegisterConcrete(NetworkAwardProposalParams{}, ModuleName+"/NetworkAwardProposalParams", nil)
	cdc.RegisterConcrete(AddressProposalParams{}, ModuleName+"/AddressProposalParams", nil)
	cdc.RegisterConcrete(SoftwareUpgradeProposalParams{}, ModuleName+"/SoftwareUpgradeProposalParams", nil)
	cdc.RegisterConcrete(MinAmountProposalParams{}, ModuleName+"/MinAmountProposalParams", nil)
	cdc.RegisterConcrete(ShortCountProposalParams{}, ModuleName+"/ShortCountProposalParams", nil)
	cdc.RegisterConcrete(StatusProposalParams{}, ModuleName+"/StatusProposalParams", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
