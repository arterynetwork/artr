package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group delegating queries under a subcommand
	delegatingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"d"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	delegatingQueryCmd.AddCommand(
		getRevokingCmd(),
		getAccumulationCmd(),
		util.LineBreak(),
		cmdGet(),
		util.LineBreak(),
		getParamsCmd(),
	)

	return delegatingQueryCmd
}

func getRevokingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoking <address>",
		Aliases: []string{"r"},
		Short:   "how many coins are being revoked from delegating",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			accAddress := args[0]

			res, err := queryClient.Revoking(
				context.Background(),
				&types.RevokingRequest{
					AccAddress: accAddress,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Revoking)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getAccumulationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "accum <address>",
		Aliases: []string{"a"},
		Short:   "Info about next payment accumulation progress",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			accAddress := args[0]

			res, err := queryClient.Accumulation(
				context.Background(),
				&types.AccumulationRequest{
					AccAddress: accAddress,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getParamsCmd() *cobra.Command {
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

func cmdGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <address>",
		Short: "get all the module info for a specified account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			accAddress := args[0]

			res, err := queryClient.Get(
				context.Background(),
				&types.GetRequest{
					AccAddress: accAddress,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}
