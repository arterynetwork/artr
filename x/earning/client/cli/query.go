package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/earning/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group earning queries under a subcommand
	earningQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"e"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	earningQueryCmd.AddCommand(
		getCmdGet(),
		getCmdList(),
		util.LineBreak(),
		getCmdParams(),
	)

	return earningQueryCmd
}

func getCmdGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <address>...",
		Short: "Get earners info by account addresses",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetMultiple(
				context.Background(),
				&types.GetMultipleRequest{Addresses: args},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Earners)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [limit [page]]",
		Short: "Get a loaded earner list",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.ListRequest{
				Limit: int32(viper.GetInt(FlagLimit)),
				Page:  int32(viper.GetInt(FlagPage)),
			}

			res, err := queryClient.List(context.Background(), req)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.List)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(FlagLimit, FlagLimitDefault, "Query number of earners per page returned")
	cmd.Flags().Int(FlagPage, FlagPageDefault, "Query a specific page of paginated results")

	return cmd
}

func getCmdParams() *cobra.Command {
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

			res, err := queryClient.Params(
				context.Background(),
				&types.ParamsRequest{},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Params)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}
