package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	serverCmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/arterynetwork/artr/app"
	"github.com/arterynetwork/artr/x/bank"
	bankcmd "github.com/arterynetwork/artr/x/bank/client/cli"
)

const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func main() {
	app.InitConfig()
	ec := app.NewEncodingConfig()

	app.ModuleBasics.RegisterInterfaces(ec.InterfaceRegistry)
	ec.InterfaceRegistry.RegisterInterface("tendermint.crypto.PubKey", (*cryptoTypes.PubKey)(nil), &secp256k1.PubKey{})
	ec.InterfaceRegistry.RegisterInterface("cosmos.tx.v1beta1.Tx", (*sdk.Tx)(nil), &tx.Tx{})

	clientCtx := ec.BuildClientContext().
		WithInput(os.Stdin).
		WithViper("")

	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "artrd",
		Short:             "Artery Blockchain node (server + client)",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx = client.ReadHomeFlag(clientCtx, cmd)

			clientCtx, err := config.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd)
		},
	}

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		debug.Cmd(),
	)
	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp(ec), exportAppState(ec), addModuleInitFlags)
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCmd(),
		txCmd(),
		keys.Commands(app.DefaultCLIHome),
		config.Cmd(),
	)

	if err := serverCmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}

func newApp(ec app.EncodingConfig) serverTypes.AppCreator {
	return func	(logger	log.Logger, db	dbm.DB, traceStore	io.Writer, appOpts	serverTypes.AppOptions) serverTypes.Application{
		var cache sdk.MultiStorePersistentCache

		if viper.GetBool(server.FlagInterBlockCache){
			cache = store.NewCommitKVStoreCacheManager()
		}

		return app.NewArteryApp(
			logger, db, traceStore, true, invCheckPeriod, ec,
			baseapp.SetPruning(types.NewPruningOptionsFromString(viper.GetString("pruning"))),
			baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
			baseapp.SetHaltHeight(viper.GetUint64(server.FlagHaltHeight)),
			baseapp.SetHaltTime(viper.GetUint64(server.FlagHaltTime)),
			baseapp.SetInterBlockCache(cache),
		)
	}
}

func exportAppState(ec app.EncodingConfig) serverTypes.AppExporter{
	return func (
		logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
		_ serverTypes.AppOptions,
	) (serverTypes.ExportedApp, error) {

		if height != -1 {
			aApp := app.NewArteryApp(logger, db, traceStore, false, uint(1), ec)
			err := aApp.LoadHeight(height)
			if err != nil {
				return serverTypes.ExportedApp{}, err
			}
			return aApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
		}

		aApp := app.NewArteryApp(logger, db, traceStore, true, uint(1), ec)

		return aApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
}

func addModuleInitFlags(_ *cobra.Command) { }

func queryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(),
		flags.LineBreak,
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
		flags.LineBreak,
	)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd)

	return queryCmd
}

func txCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.NewSendTxCmd(),
		flags.LineBreak,
		authcmd.GetSignCommand(),
		authcmd.GetMultiSignCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd)

	// remove auth and bank commands as they're mounted under the root tx command
	var cmdsToRemove []*cobra.Command

	for _, cmd := range txCmd.Commands() {
		if cmd.Use == authtypes.ModuleName || cmd.Use == bank.ModuleName {
			cmdsToRemove = append(cmdsToRemove, cmd)
		}
	}

	txCmd.RemoveCommand(cmdsToRemove...)

	return txCmd
}
