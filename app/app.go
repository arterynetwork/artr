package app

import (
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/earning"
	"github.com/arterynetwork/artr/x/noding"
	"github.com/arterynetwork/artr/x/profile"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/schedule"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/subscription"
	"github.com/arterynetwork/artr/x/voting"
	"github.com/arterynetwork/artr/x/vpn"
	"encoding/json"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/arterynetwork/artr/x/bank"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	//distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
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
		supply.AppModuleBasic{},
		referral.AppModuleBasic{},
		profile.AppModuleBasic{},
		schedule.AppModuleBasic{},
		delegating.AppModuleBasic{},
		vpn.AppModuleBasic{},
		storage.AppModuleBasic{},
		subscription.AppModuleBasic{},
		voting.AppModuleBasic{},
		noding.AppModuleBasic{},
		earning.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName: nil,
		vpn.ModuleName:        nil,
		storage.ModuleName:    nil,
		noding.ModuleName:     nil,
		earning.ModuleName:    nil,
	}
)

// MakeCodec creates the application codec. The codec is sealed before it is
// returned.
func MakeCodec() *codec.Codec {
	var cdc = codec.New()

	ModuleBasics.RegisterCodec(cdc)
	vesting.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc.Seal()
}

// ArteryApp extended ABCI application
type ArteryApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tKeys map[string]*sdk.TransientStoreKey

	// subspaces
	subspaces map[string]params.Subspace

	// keepers
	accountKeeper      auth.AccountKeeper
	bankKeeper         bank.Keeper
	supplyKeeper       supply.Keeper
	paramsKeeper       params.Keeper
	upgradeKeeper      upgrade.Keeper
	referralKeeper     referral.Keeper
	profileKeeper      profile.Keeper
	scheduleKeeper     schedule.Keeper
	delegatingKeeper   delegating.Keeper
	vpnKeeper          vpn.Keeper
	storageKeeper      storage.Keeper
	subscriptionKeeper subscription.Keeper
	votingKeeper       voting.Keeper
	nodingKeeper       noding.Keeper
	earningKeeper      earning.Keeper

	// Module Manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
}

// verify app interface at compile time
var _ simapp.App = (*ArteryApp)(nil)

// NewArteryApp is a constructor function for ArteryApp
func NewArteryApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) *ArteryApp {
	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(bam.MainStoreKey, auth.StoreKey,
		supply.StoreKey, params.StoreKey, upgrade.StoreKey,
		profile.StoreKey, profile.AliasStoreKey, profile.CardStoreKey,
		schedule.StoreKey, referral.StoreKey, referral.IndexStoreKey, delegating.MainStoreKey,
		delegating.ClusterStoreKey, vpn.StoreKey, storage.StoreKey,
		subscription.StoreKey, voting.StoreKey, noding.StoreKey, noding.IdxStoreKey, earning.StoreKey)

	tKeys := sdk.NewTransientStoreKeys(params.TStoreKey)

	// Here you initialize your application with the store keys it requires
	var app = &ArteryApp{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tKeys:          tKeys,
		subspaces:      make(map[string]params.Subspace),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, keys[params.StoreKey], tKeys[params.TStoreKey])
	// Set specific subspaces
	app.subspaces[auth.ModuleName] = app.paramsKeeper.Subspace(auth.DefaultParamspace)
	app.subspaces[bank.ModuleName] = app.paramsKeeper.Subspace(bank.DefaultParamspace)
	app.subspaces[referral.ModuleName] = app.paramsKeeper.Subspace(referral.DefaultParamspace)
	app.subspaces[profile.ModuleName] = app.paramsKeeper.Subspace(profile.ModuleName)
	app.subspaces[schedule.ModuleName] = app.paramsKeeper.Subspace(schedule.ModuleName)
	app.subspaces[vpn.ModuleName] = app.paramsKeeper.Subspace(vpn.ModuleName)
	app.subspaces[storage.ModuleName] = app.paramsKeeper.Subspace(storage.DefaultParamspace)
	app.subspaces[delegating.ModuleName] = app.paramsKeeper.Subspace(delegating.DefaultParamspace)
	app.subspaces[subscription.ModuleName] = app.paramsKeeper.Subspace(subscription.DefaultParamspace)
	app.subspaces[voting.ModuleName] = app.paramsKeeper.Subspace(voting.DefaultParamspace)
	app.subspaces[noding.ModuleName] = app.paramsKeeper.Subspace(noding.DefaultParamspace)
	app.subspaces[earning.DefaultParamspace] = app.paramsKeeper.Subspace(earning.DefaultParamspace)

	// Scheduler handles block height based tasks
	app.scheduleKeeper = schedule.NewKeeper(
		cdc,
		keys[schedule.StoreKey],
		app.subspaces[schedule.ModuleName],
	)

	//app.scheduleKeeper.AddHook("event-test", func(ctx sdk.Context, data []byte) {
	//	ctx.Logger().Error("test event called")
	//	addr, _ := sdk.AccAddressFromBech32("cosmos1ey3aa0uxndvdrvgyvsd0afyt69uet9avw7cseq")
	//	coins := sdk.NewCoins(sdk.NewCoin("artr", sdk.NewInt(1000)))
	//	app.bankKeeper.AddCoins(ctx, addr, coins)
	//})

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keys[auth.StoreKey],
		app.subspaces[auth.ModuleName],
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.subspaces[bank.ModuleName],
		app.ModuleAccountAddrs(),
	)

	//app.bankKeeper.AddHook("SetCoins", "test-event", func(ctx sdk.Context, acc authexported.Account) {
	//	logger.Error("Set coins hook", acc)
	//})

	// The SupplyKeeper collects transaction fees and renders them to the fee distribution module
	app.supplyKeeper = supply.NewKeeper(
		app.cdc,
		keys[supply.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		maccPerms,
	)

	app.referralKeeper = referral.NewKeeper(
		app.cdc,
		keys[referral.StoreKey],
		keys[referral.IndexStoreKey],
		app.subspaces[referral.ModuleName],
		app.accountKeeper,
		app.scheduleKeeper,
		app.bankKeeper,
		app.supplyKeeper,
	)

	app.profileKeeper = profile.NewKeeper(
		app.cdc,
		keys[profile.StoreKey],
		keys[profile.AliasStoreKey],
		keys[profile.CardStoreKey],
		app.subspaces[profile.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		app.referralKeeper,
		app.supplyKeeper,
	)

	app.delegatingKeeper = delegating.NewKeeper(
		app.cdc,
		keys[delegating.MainStoreKey],
		keys[delegating.ClusterStoreKey],
		app.subspaces[delegating.DefaultParamspace],
		app.accountKeeper,
		app.scheduleKeeper,
		app.profileKeeper,
		app.bankKeeper,
		app.supplyKeeper,
		app.referralKeeper,
	)

	app.vpnKeeper = vpn.NewKeeper(
		app.cdc,
		keys[vpn.StoreKey],
		app.subspaces[vpn.ModuleName],
	)

	app.storageKeeper = storage.NewKeeper(
		app.cdc,
		keys[storage.StoreKey],
		app.subspaces[storage.ModuleName],
	)

	app.subscriptionKeeper = subscription.NewKeeper(
		app.cdc,
		keys[subscription.StoreKey],
		app.subspaces[subscription.DefaultParamspace],
		app.bankKeeper,
		app.referralKeeper,
		app.scheduleKeeper,
		app.vpnKeeper,
		app.storageKeeper,
		app.supplyKeeper,
		app.profileKeeper,
	)

	app.upgradeKeeper = upgrade.NewKeeper(
		map[int64]bool{},
		keys[upgrade.StoreKey],
		cdc,
	)

	app.nodingKeeper = noding.NewKeeper(
		app.cdc,
		keys[noding.StoreKey],
		keys[noding.IdxStoreKey],
		app.referralKeeper,
		app.scheduleKeeper,
		app.supplyKeeper,
		app.subspaces[noding.DefaultParamspace],
		auth.FeeCollectorName,
	)

	app.earningKeeper = earning.NewKeeper(
		app.cdc,
		keys[earning.StoreKey],
		app.subspaces[earning.DefaultParamspace],
		app.supplyKeeper,
		app.scheduleKeeper,
	)

	app.votingKeeper = voting.NewKeeper(
		app.cdc,
		keys[voting.StoreKey],
		app.subspaces[voting.DefaultParamspace],
		app.scheduleKeeper,
		app.upgradeKeeper,
		app.nodingKeeper,
		app.delegatingKeeper,
		app.referralKeeper,
		app.subscriptionKeeper,
		app.profileKeeper,
		app.earningKeeper,
		app.vpnKeeper,
		app.bankKeeper,
	)

	app.bankKeeper.AddHook("SetCoins", "update-referral",
		func(ctx sdk.Context, acc authexported.Account) error {
			err := app.referralKeeper.OnBalanceChanged(ctx, acc.GetAddress())

			if err != nil {
				return sdkerrors.Wrap(err, "update-referral hook error")
			}

			return nil
		})

	app.scheduleKeeper.AddHook(referral.StatusDowngradeHookName, app.referralKeeper.PerformDowngrade)
	app.scheduleKeeper.AddHook(referral.CompressionHookName, app.referralKeeper.PerformCompression)
	app.scheduleKeeper.AddHook(referral.TransitionTimeoutHookName, app.referralKeeper.PerformTransitionTimeout)
	app.scheduleKeeper.AddHook(subscription.HookName, app.subscriptionKeeper.ProcessSchedule)
	app.scheduleKeeper.AddHook(voting.HookName, app.votingKeeper.ProcessSchedule)
	app.scheduleKeeper.AddHook(earning.StartHookName, app.earningKeeper.MustPerformStart)
	app.scheduleKeeper.AddHook(earning.ContinueHookName, app.earningKeeper.MustPerformContinue)
	app.scheduleKeeper.AddHook(delegating.RevokeHookName, app.delegatingKeeper.MustPerformRevoking)

	app.referralKeeper.AddHook(referral.StatusUpdatedCallback, app.nodingKeeper.OnStatusUpdate)
	app.referralKeeper.AddHook(referral.StakeChangedCallback, app.nodingKeeper.OnStakeChanged)

	app.upgradeKeeper.SetUpgradeHandler("1.1.1", NopUpgradeHandler)
	//Cancelled: app.upgradeKeeper.SetUpgradeHandler("1.1.2", CliWarningUpgradeHandler)
	app.upgradeKeeper.SetUpgradeHandler("1.1.3", Chain(
		CliWarningUpgradeHandler,
		RefreshStatus(app.referralKeeper, referral.StatusLeader),
	))
	app.upgradeKeeper.SetUpgradeHandler("1.2.0", Chain(
		InitializeTransitionCost(app.referralKeeper, app.subspaces[referral.ModuleName]),
		RestoreTrafficLimit(app.storageKeeper),
		ScheduleCompression(app.referralKeeper),
		CountRevoking(app.accountKeeper, app.referralKeeper),
	))
	app.upgradeKeeper.SetUpgradeHandler("1.2.1", Chain(
		ClearInvalidNicknames(app.accountKeeper, app.profileKeeper),
		InitializeMinDelegate(app.delegatingKeeper, app.subspaces[delegating.ModuleName]),
	))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		schedule.NewAppModule(app.scheduleKeeper),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper, app.supplyKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		profile.NewAppModule(app.profileKeeper, app.accountKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		referral.NewAppModule(app.referralKeeper, app.accountKeeper, app.scheduleKeeper, app.bankKeeper, app.supplyKeeper),
		delegating.NewAppModule(app.delegatingKeeper, app.accountKeeper, app.scheduleKeeper, app.bankKeeper, app.supplyKeeper, app.profileKeeper, app.referralKeeper),
		vpn.NewAppModule(app.vpnKeeper),
		storage.NewAppModule(app.storageKeeper),
		subscription.NewAppModule(app.subscriptionKeeper),
		noding.NewAppModule(app.nodingKeeper, app.referralKeeper, app.scheduleKeeper, app.supplyKeeper),
		earning.NewAppModule(app.earningKeeper, app.supplyKeeper, app.scheduleKeeper),
		voting.NewAppModule(app.votingKeeper, app.scheduleKeeper, app.upgradeKeeper, app.nodingKeeper, app.delegatingKeeper, app.referralKeeper, app.subscriptionKeeper, app.profileKeeper, app.earningKeeper, app.vpnKeeper),
	)
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.

	app.mm.SetOrderBeginBlockers(upgrade.ModuleName, noding.ModuleName, referral.ModuleName, delegating.ModuleName, schedule.ModuleName)
	app.mm.SetOrderEndBlockers(noding.ModuleName)

	// Sets the order of Genesis - Order matters, genutil is to always come last
	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		schedule.ModuleName,
		auth.ModuleName,
		bank.ModuleName,
		profile.ModuleName,
		referral.ModuleName,
		delegating.ModuleName,
		vpn.ModuleName,
		storage.ModuleName,
		subscription.ModuleName,
		voting.ModuleName,
		supply.ModuleName,
		noding.ModuleName,
		earning.ModuleName,
	)

	// register all module routes and module queriers
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(
		auth.NewAnteHandler(
			app.accountKeeper,
			app.supplyKeeper,
			auth.DefaultSigVerificationGasConsumer,
		),
	)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tKeys)

	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	return ModuleBasics.DefaultGenesis()
}

// InitChainer application update at chain initialization
func (app *ArteryApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState simapp.GenesisState

	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)

	return app.mm.InitGenesis(ctx, genesisState)
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
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *ArteryApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// Codec returns the application's sealed codec.
func (app *ArteryApp) Codec() *codec.Codec {
	return app.cdc
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
