package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"

	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	config2 "github.com/cosmos/cosmos-sdk/server/config"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeKeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	_ "github.com/arterynetwork/artr/client/docs/statik"
	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/earning"
	earningKeeper "github.com/arterynetwork/artr/x/earning/keeper"
	earningTypes "github.com/arterynetwork/artr/x/earning/types"
	"github.com/arterynetwork/artr/x/noding"
	nodingKeeper "github.com/arterynetwork/artr/x/noding/keeper"
	nodingTypes "github.com/arterynetwork/artr/x/noding/types"
	"github.com/arterynetwork/artr/x/profile"
	profileKeeper "github.com/arterynetwork/artr/x/profile/keeper"
	profileTypes "github.com/arterynetwork/artr/x/profile/types"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/schedule"
	scheduleKeeper "github.com/arterynetwork/artr/x/schedule/keeper"
	scheduleTypes "github.com/arterynetwork/artr/x/schedule/types"
	"github.com/arterynetwork/artr/x/voting"
	votingKeeper "github.com/arterynetwork/artr/x/voting/keeper"
	votingTypes "github.com/arterynetwork/artr/x/voting/types"
)

const appName = "artery"

var (
	// DefaultCLIHome default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/.artrcli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored
	DefaultNodeHome = os.ExpandEnv("$HOME/.artrd")

	// ModuleBasics The module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		params.AppModuleBasic{},
		referral.AppModuleBasic{},
		profile.AppModuleBasic{},
		schedule.AppModuleBasic{},
		delegating.AppModuleBasic{},
		voting.AppModuleBasic{},
		noding.AppModuleBasic{},
		earning.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authTypes.FeeCollectorName:   nil,
		noding.ModuleName:            nil,
		earning.ModuleName:           nil,
		earning.VpnCollectorName:     nil,
		earning.StorageCollectorName: nil,
	}
)

// ArteryApp extended ABCI application
type ArteryApp struct {
	*bam.BaseApp
	ec             EncodingConfig
	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tKeys map[string]*sdk.TransientStoreKey

	// subspaces
	subspaces map[string]paramTypes.Subspace

	// keepers
	accountKeeper    authKeeper.AccountKeeper
	bankKeeper       bank.Keeper
	paramsKeeper     paramKeeper.Keeper
	upgradeKeeper    upgradeKeeper.Keeper
	referralKeeper   referral.Keeper
	profileKeeper    profileKeeper.Keeper
	scheduleKeeper   scheduleKeeper.Keeper
	delegatingKeeper *delegating.Keeper
	votingKeeper     votingKeeper.Keeper
	nodingKeeper     noding.Keeper
	earningKeeper    earning.Keeper

	// Module Manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
}

// verify app interface at compile time
//var _ simapp.App = (*ArteryApp)(nil)
var _ serverTypes.Application = (*ArteryApp)(nil)

// NewArteryApp is a constructor function for ArteryApp
func NewArteryApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, ec EncodingConfig, baseAppOptions ...func(*bam.BaseApp),
) *ArteryApp {
	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, ec.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)
	bApp.SetInterfaceRegistry(ec.InterfaceRegistry)

	keys := sdk.NewKVStoreKeys(authTypes.StoreKey, bank.StoreKey,
		paramTypes.StoreKey, upgradeTypes.StoreKey,
		profileTypes.StoreKey, profileTypes.AliasStoreKey, profileTypes.CardStoreKey,
		scheduleTypes.StoreKey, referral.StoreKey, referral.IndexStoreKey, delegating.MainStoreKey,
		votingTypes.StoreKey, noding.StoreKey, noding.IdxStoreKey,
		earning.StoreKey)

	tKeys := sdk.NewTransientStoreKeys(paramTypes.TStoreKey)

	//TODO: pass `ec.Marshaller` to all modules properly and use it properly in

	// Here you initialize your application with the store keys it requires
	var app = &ArteryApp{
		BaseApp:        bApp,
		ec:             ec,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tKeys:          tKeys,
		subspaces:      make(map[string]paramTypes.Subspace),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = paramKeeper.NewKeeper(
		ec.Marshaler,
		ec.Amino,
		keys[paramTypes.StoreKey],
		tKeys[paramTypes.TStoreKey],
	)
	bApp.SetParamStore(app.paramsKeeper.Subspace(bam.Paramspace).WithKeyTable(paramKeeper.ConsensusParamsKeyTable()))
	// Set specific subspaces
	app.subspaces[authTypes.ModuleName] = app.paramsKeeper.Subspace(authTypes.ModuleName)
	app.subspaces[bank.ModuleName] = app.paramsKeeper.Subspace(bank.DefaultParamspace)
	app.subspaces[referral.ModuleName] = app.paramsKeeper.Subspace(referral.DefaultParamspace)
	app.subspaces[profileTypes.ModuleName] = app.paramsKeeper.Subspace(profileTypes.ModuleName)
	app.subspaces[scheduleTypes.ModuleName] = app.paramsKeeper.Subspace(scheduleTypes.ModuleName)
	app.subspaces[delegating.ModuleName] = app.paramsKeeper.Subspace(delegating.DefaultParamspace)
	app.subspaces[votingTypes.ModuleName] = app.paramsKeeper.Subspace(votingTypes.DefaultParamspace)
	app.subspaces[noding.ModuleName] = app.paramsKeeper.Subspace(noding.DefaultParamspace)
	app.subspaces[earning.DefaultParamspace] = app.paramsKeeper.Subspace(earning.DefaultParamspace)

	// Scheduler handles block height based tasks
	app.scheduleKeeper = scheduleKeeper.NewKeeper(
		ec.Marshaler,
		keys[scheduleTypes.StoreKey],
		app.subspaces[scheduleTypes.ModuleName],
	)

	//app.scheduleKeeper.AddHook("event-test", func(ctx sdk.Context, data []byte) {
	//	ctx.Logger().Error("test event called")
	//	addr, _ := sdk.AccAddressFromBech32("cosmos1ey3aa0uxndvdrvgyvsd0afyt69uet9avw7cseq")
	//	coins := sdk.NewCoins(sdk.NewCoin("artr", sdk.NewInt(1000)))
	//	app.bankKeeper.AddCoins(ctx, addr, coins)
	//})

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = authKeeper.NewAccountKeeper(
		ec.Marshaler,
		keys[authTypes.StoreKey],
		app.subspaces[authTypes.ModuleName],
		authTypes.ProtoBaseAccount,
		map[string][]string{
			authTypes.FeeCollectorName:        {},
			earningTypes.VpnCollectorName:     {},
			earningTypes.StorageCollectorName: {},
			earningTypes.ModuleName:           {},
		},
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		ec.Marshaler,
		keys[bank.StoreKey],
		app.accountKeeper,
		app.subspaces[bank.ModuleName],
		make(map[string]bool, 0),
	)

	//app.bankKeeper.AddHook("SetCoins", "test-event", func(ctx sdk.Context, acc authexported.Account) {
	//	logger.Error("Set coins hook", acc)
	//})

	app.referralKeeper = referral.NewKeeper(
		ec.Marshaler,
		keys[referral.StoreKey],
		keys[referral.IndexStoreKey],
		app.subspaces[referral.ModuleName],
		app.accountKeeper,
		app.scheduleKeeper,
		app.bankKeeper,
		app.bankKeeper,
	)

	app.profileKeeper = profileKeeper.NewKeeper(
		ec.Marshaler,
		keys[profileTypes.StoreKey],
		keys[profileTypes.AliasStoreKey],
		keys[profileTypes.CardStoreKey],
		app.subspaces[profileTypes.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		app.referralKeeper,
		app.scheduleKeeper,
	)

	app.delegatingKeeper = delegating.NewKeeper(
		ec.Marshaler,
		keys[delegating.MainStoreKey],
		app.subspaces[delegating.DefaultParamspace],
		app.accountKeeper,
		app.scheduleKeeper,
		app.profileKeeper,
		app.bankKeeper,
		app.referralKeeper,
	)

	app.upgradeKeeper = upgradeKeeper.NewKeeper(
		map[int64]bool{},
		keys[upgradeTypes.StoreKey],
		ec.Marshaler,
		"",
	)

	app.nodingKeeper = nodingKeeper.NewKeeper(
		ec.Marshaler,
		keys[nodingTypes.StoreKey],
		keys[nodingTypes.IdxStoreKey],
		app.referralKeeper,
		app.accountKeeper,
		app.bankKeeper,
		app.subspaces[noding.DefaultParamspace],
		authTypes.FeeCollectorName,
	)

	app.earningKeeper = earningKeeper.NewKeeper(
		ec.Marshaler,
		keys[earningTypes.StoreKey],
		app.subspaces[earningTypes.DefaultParamspace],
		app.accountKeeper,
		app.bankKeeper,
		app.scheduleKeeper,
	)

	app.votingKeeper = votingKeeper.NewKeeper(
		ec.Marshaler,
		keys[votingTypes.StoreKey],
		app.subspaces[votingTypes.DefaultParamspace],
		app.scheduleKeeper,
		app.upgradeKeeper,
		app.nodingKeeper,
		app.delegatingKeeper,
		app.referralKeeper,
		app.profileKeeper,
		app.earningKeeper,
		app.bankKeeper,
	)

	app.delegatingKeeper.SetKeepers(app.nodingKeeper)

	app.bankKeeper.AddHook("SetCoins", "update-referral",
		func(ctx sdk.Context, addr sdk.AccAddress) error {
			if err := app.referralKeeper.OnBalanceChanged(ctx, addr.String()); err != nil {
				return errors.Wrap(err, "update-referral hook error")
			}

			return nil
		})

	app.scheduleKeeper.AddHook(referral.StatusDowngradeHookName, app.referralKeeper.PerformDowngrade)
	app.scheduleKeeper.AddHook(referral.CompressionHookName, app.referralKeeper.PerformCompression)
	app.scheduleKeeper.AddHook(referral.TransitionTimeoutHookName, app.referralKeeper.PerformTransitionTimeout)
	app.scheduleKeeper.AddHook(profileTypes.RefreshHookName, app.profileKeeper.HandleRenewHook)
	app.scheduleKeeper.AddHook(profileTypes.RefreshImHookName, app.profileKeeper.HandleRenewImHook)
	app.scheduleKeeper.AddHook(votingTypes.VoteHookName, app.votingKeeper.ProcessSchedule)
	app.scheduleKeeper.AddHook(votingTypes.PollHookName, app.votingKeeper.EndPollHandler)
	app.scheduleKeeper.AddHook(earning.StartHookName, app.earningKeeper.MustPerformStart)
	app.scheduleKeeper.AddHook(earning.ContinueHookName, app.earningKeeper.MustPerformContinue)
	app.scheduleKeeper.AddHook(delegating.RevokeHookName, app.delegatingKeeper.MustPerformRevoking)
	app.scheduleKeeper.AddHook(delegating.AccrueHookName, app.delegatingKeeper.MustPerformAccrue)
	app.scheduleKeeper.AddHook(referral.BanishHookName, app.referralKeeper.PerformBanish)
	app.scheduleKeeper.AddHook(referral.StatusBonusHookName, app.referralKeeper.PerformStatusBonus)

	app.referralKeeper.AddHook(referral.StatusUpdatedCallback, app.nodingKeeper.OnStatusUpdate)
	app.referralKeeper.AddHook(referral.StakeChangedCallback, app.nodingKeeper.OnStakeChanged)
	app.referralKeeper.AddHook(referral.BanishedCallback, app.delegatingKeeper.OnBanished)

	app.upgradeKeeper.SetUpgradeHandler("2.0.1", RecalculateActiveReferrals(app.referralKeeper))
	app.upgradeKeeper.SetUpgradeHandler("2.1.0",
		ScheduleBanishment(
			app.referralKeeper,
			app.bankKeeper,
			keys[referral.StoreKey],
			keys[scheduleTypes.StoreKey],
			ec.Marshaler,
		),
	)
	app.upgradeKeeper.SetUpgradeHandler("2.2.0", Chain(
		InitPollPeriodParam(app.votingKeeper, app.subspaces[votingTypes.DefaultParamspace]),
		ForceOnStatusChangedCallback(app.nodingKeeper),
	))
	app.upgradeKeeper.SetUpgradeHandler("2.2.1", Chain(
		ForceGlobalDelegation(
			app.referralKeeper,
			app.bankKeeper,
			*app.delegatingKeeper,
			app.scheduleKeeper,
			keys[bank.StoreKey],
			keys[delegating.MainStoreKey],
			ec.Marshaler,
		),
		RefreshReferralStatuses(app.referralKeeper),
	))
	app.upgradeKeeper.SetUpgradeHandler("2.3.0", NopUpgradeHandler)
	app.upgradeKeeper.SetUpgradeHandler("2.3.1", RefreshReferralStatuses(app.referralKeeper))
	app.upgradeKeeper.SetUpgradeHandler("2.3.2", Chain(
		TransferFromTheBanished(app.scheduleKeeper, ec.Marshaler, keys[referral.StoreKey]),
		UnbanishAccountsWithDelegation(app.bankKeeper, app.scheduleKeeper, ec.Marshaler, keys[referral.StoreKey]),
		RefreshReferralStatuses(app.referralKeeper),
	))

	app.upgradeKeeper.SetUpgradeHandler("2.4.0", InitValidatorBonusParam())
	app.upgradeKeeper.SetUpgradeHandler("2.4.1",
		InitValidatorParam(*app.delegatingKeeper, app.subspaces[delegating.DefaultParamspace]),
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		schedule.NewAppModule(app.scheduleKeeper),
		auth.NewAppModule(ec.Marshaler, app.accountKeeper, nil),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		profile.NewAppModule(app.profileKeeper, app.accountKeeper),
		referral.NewAppModule(
			app.referralKeeper, app.accountKeeper, app.scheduleKeeper, app.bankKeeper, app.bankKeeper,
		),
		delegating.NewAppModule(
			*app.delegatingKeeper, app.accountKeeper, app.scheduleKeeper, app.bankKeeper, app.profileKeeper,
			app.referralKeeper,
		),
		noding.NewAppModule(
			app.nodingKeeper, app.referralKeeper, app.accountKeeper, app.bankKeeper,
		),
		earning.NewAppModule(app.earningKeeper, app.bankKeeper, app.scheduleKeeper),
		voting.NewAppModule(
			app.votingKeeper, app.scheduleKeeper, app.upgradeKeeper, app.nodingKeeper, app.delegatingKeeper,
			app.referralKeeper, app.profileKeeper, app.earningKeeper,
		),
	)

	app.RegisterInterfaces(ec.InterfaceRegistry)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.

	app.mm.SetOrderBeginBlockers(
		upgradeTypes.ModuleName,
		noding.ModuleName,
		referral.ModuleName,
		delegating.ModuleName,
		scheduleTypes.ModuleName,
	)
	app.mm.SetOrderEndBlockers(noding.ModuleName)

	// Sets the order of Genesis - Order matters, genutil is to always come last
	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		scheduleTypes.ModuleName,
		authTypes.ModuleName,
		referral.ModuleName,
		bank.ModuleName,
		profileTypes.ModuleName,
		delegating.ModuleName,
		noding.ModuleName,
		votingTypes.ModuleName,
		earning.ModuleName,
	)

	// register all module routes and module queriers
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), ec.Amino)
	app.mm.RegisterServices(module.NewConfigurator(app.MsgServiceRouter(), app.GRPCQueryRouter()))

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(
		ante.NewAnteHandler(
			app.accountKeeper,
			app.bankKeeper,
			ante.DefaultSigVerificationGasConsumer,
			ec.TxConfig.SignModeHandler(),
		),
	)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tKeys)

	if loadLatest {
		err := app.LoadLatestVersion()
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

func (app *ArteryApp) RegisterInterfaces(registry codecTypes.InterfaceRegistry) {
	for _, am := range app.mm.Modules {
		am.RegisterInterfaces(registry)
	}
	registry.RegisterInterface("tendermint.crypto.PubKey", (*cryptoTypes.PubKey)(nil), &secp256k1.PubKey{})
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(mrshl codec.JSONMarshaler) GenesisState {
	return ModuleBasics.DefaultGenesis(mrshl)
}

// InitChainer application update at chain initialization
func (app *ArteryApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState simapp.GenesisState

	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	return app.mm.InitGenesis(ctx, app.ec.Marshaler, genesisState)
}

// BeginBlocker application updates every begin block
func (app *ArteryApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *ArteryApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// LoadHeight loads a particular height
func (app *ArteryApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// Codec returns the application's sealed codec.
func (app *ArteryApp) Codec() codec.BinaryMarshaler {
	return app.ec.Marshaler
}

// SimulationManager implements the SimulationApp interface
func (app *ArteryApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// GetMaccPerms returns a mapping of the application's module account permissions.
func GetMaccPerms() map[string][]string {
	modAccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		modAccPerms[k] = v
	}
	return modAccPerms
}

func (app *ArteryApp) RegisterAPIRoutes(server *api.Server, apiConfig config2.APIConfig) {
	clientCtx := server.ClientCtx
	rpc.RegisterRoutes(clientCtx, server.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, server.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, server.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, server.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterRESTRoutes(clientCtx, server.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, server.GRPCGatewayRouter)

	if apiConfig.Swagger {
		RegisterSwaggerAPI(server.Router)
	}
}

func (app *ArteryApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.ec.InterfaceRegistry)
}

func (app *ArteryApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.ec.InterfaceRegistry)
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.NewWithNamespace("swagger")
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}
