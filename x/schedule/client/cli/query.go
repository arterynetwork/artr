package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/schedule/types"
)

func CmdQuery() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"s"},
		Short:                      "Querying commands for the x/schedule module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		cmdAtHeight(),
		cmdAll(),
	)
	return queryCmd
}

func cmdAtHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "in <since> <to>",
		Short: "Get all tasks scheduled to a specified time range",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			since := args[0]
			to := args[1]

			res, err := queryClient.Get(
				context.Background(),
				&types.GetRequest{Since: since, To: to},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Tasks)
		},
	}
	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdAll() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all",
		Aliases: []string{"a"},
		Short:   "Get all the scheduled tasks",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.All(
				context.Background(),
				&types.AllRequest{},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Tasks)
		},
	}
	util.AddQueryFlagsToCmd(cmd)
	return cmd
}
