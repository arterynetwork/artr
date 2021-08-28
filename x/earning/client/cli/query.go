package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

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
		cmdList(),
		cmdState(),
		util.LineBreak(),
		getCmdParams(),
	)

	return earningQueryCmd
}

func cmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a loaded earner list",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.List(
				context.Background(),
				&types.ListRequest{},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.List)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Get the scheduled payment (if any) parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.State(
				context.Background(),
				&types.StateRequest{},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.State)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
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

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
