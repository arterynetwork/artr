package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank/types"
	"github.com/cosmos/cosmos-sdk/client"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group bank queries under a subcommand
	bankQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"b"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	bankQueryCmd.AddCommand(
		getParamsCmd(),
		cmdSupply(),
		cmdBalance(),
	)

	return bankQueryCmd
}

func getParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Aliases: []string{"p"},
		Short:   "Get module parameters",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.ParamsRequest{})
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Params)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "supply",
		Aliases: []string{"s"},
		Short:   "Get supply total",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Supply(context.Background(), &types.SupplyRequest{})
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Supply.Total)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balance <acc_address>",
		Aliases: []string{"b"},
		Short:   "Get an account balance",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			accAddress := args[0]

			res, err := queryClient.Balance(context.Background(), &types.BalanceRequest{AccAddress: accAddress})
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Balance)
		},
	}

	util.AddQueryFlagsToCmd(cmd)
	return cmd
}
