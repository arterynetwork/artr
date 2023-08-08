package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type EncodingConfig struct {
	Marshaler         codec.Marshaler
	InterfaceRegistry codecTypes.InterfaceRegistry
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func (conf EncodingConfig) BuildClientContext() client.Context {
	return client.Context{}.
		WithJSONMarshaler(conf.Marshaler).
		WithInterfaceRegistry(conf.InterfaceRegistry).
		WithTxConfig(conf.TxConfig).
		WithLegacyAmino(conf.Amino).
		WithAccountRetriever(authTypes.AccountRetriever{})
}

func NewEncodingConfig() EncodingConfig {
	var (
		ir       = codecTypes.NewInterfaceRegistry()
		cdc      = codec.NewProtoCodec(ir)
		txConfig = tx.NewTxConfig(cdc, tx.DefaultSignModes)
		amino    = codec.NewLegacyAmino()
	)
	return EncodingConfig{
		InterfaceRegistry: ir,
		Marshaler:         cdc,
		TxConfig:          txConfig,
		Amino:             amino,
	}
}
