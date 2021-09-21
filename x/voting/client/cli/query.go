package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/voting/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group voting queries under a subcommand
	votingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	votingQueryCmd.AddCommand(
		cmdGovernment(),
		cmdCurrent(),
		cmdHistory(),
		util.LineBreak(),
		cmdParams(),
	)

	return votingQueryCmd
}

func cmdHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Query history of voting",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.HistoryRequest{
				Limit: int32(viper.GetInt(FlagLimit)),
				Page:  int32(viper.GetInt(FlagPage)),
			}

			res, err := queryClient.History(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.History)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(FlagLimit, FlagLimitDefault, "Query number of history records per page returned")
	cmd.Flags().Int(FlagPage, FlagPageDefault, "Query a specific page of paginated results")

	return cmd
}

func cmdGovernment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "government",
		Short: "Query government accounts list",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.GovernmentRequest{}

			res, err := queryClient.Government(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Members)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdCurrent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Query current proposal and its status: votes given and voters list",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.CurrentRequest{}

			res, err := queryClient.Current(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Aliases: []string{"p"},
		Short:   "Get the module params",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.ParamsRequest{}

			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Params)
		},
	}
	util.AddQueryFlagsToCmd(cmd)
	return cmd
}
