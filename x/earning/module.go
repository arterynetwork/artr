package earning

import (
	"context"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/arterynetwork/artr/x/earning/client/cli"
	"github.com/arterynetwork/artr/x/earning/keeper"
	"github.com/arterynetwork/artr/x/earning/types"
)

// Type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the earning module.
type AppModuleBasic struct{}

// Name returns the earning module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the earning module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers interfaces and implementations of the delegating module.
func (AppModuleBasic) RegisterInterfaces(registry codecTypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the earning
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the earning module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data GenesisState
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal x/earning genesis state")
	}
	return ValidateGenesis(data)
}

// RegisterRESTRoutes registers the REST routes for the earning module.
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {}

// GetTxCmd returns the root tx command for the earning module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns no root query command for the earning module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.NewQueryCmd()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the earning module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

//____________________________________________________________________________

// AppModule implements an application module for the earning module.
type AppModule struct {
	AppModuleBasic

	keeper         Keeper
	supplyKeeper   types.BankKeeper
	scheduleKeeper types.ScheduleKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k Keeper, supplyKeeper types.BankKeeper, scheduleKeeper types.ScheduleKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		supplyKeeper:   supplyKeeper,
		scheduleKeeper: scheduleKeeper,
	}
}

// Name returns the earning module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.MsgServer(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.QueryServer(am.keeper))
}

// RegisterInvariants registers the earning module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the earning module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

// QuerierRoute returns the earning module's querier route name.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns the earning module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return NewQuerier(am.keeper, legacyQuerierCdc)
}

// InitGenesis performs genesis initialization for the earning module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, mrshl codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	mrshl.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the earning
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, mrshl codec.JSONMarshaler) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return mrshl.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the earning module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}

// EndBlock returns the end blocker for the earning module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
