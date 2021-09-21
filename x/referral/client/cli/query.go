package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group referral queries under a subcommand
	referralQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"ref", "r"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	referralQueryCmd.AddCommand(
		getCmdInfo(),
		getCoinsCmd(),
		getCheckStatusCmd(),
		getValidateTransitionCmd(),

		util.LineBreak(),
		cmdAllWithStatus(),
		util.LineBreak(),
		getCmdParams(),
	)

	return referralQueryCmd
}

func getCmdInfo() *cobra.Command {
	var light bool

	cmd := &cobra.Command{
		Use:     "info <address>",
		Aliases: []string{"i"},
		Short:   "Get all info for the account",
		Args:    cobra.ExactArgs(1),
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
					Light:      light,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Info)
		},
	}
	cmd.Flags().BoolVarP(&light, "light", "l", false, "omit Referrals and ActiveReferrals fields")
	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getCoinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "coins <address> [max_depth]",
		Aliases: []string{"c"},
		Short:   "Get coins total in one's referral structure",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			accAddress := args[0]

			var maxDepth uint32
			if len(args) > 1 {
				if n, err := strconv.ParseUint(args[1], 0, 32); err == nil {
					maxDepth = uint32(n)
				} else {
					return errors.Wrap(err, "cannot parse max_depth")
				}
			}

			res, err := queryClient.Coins(
				context.Background(),
				&types.CoinsRequest{
					AccAddress: accAddress,
					MaxDepth:   maxDepth,
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

func getCheckStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check-status <address> <n>",
		Aliases: []string{"check_status", "cs"},
		Short:   "Check if status #n requirements are fulfilled",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			accAddress := args[0]

			var status types.Status
			if len(args) > 1 {
				if n, err := strconv.ParseUint(args[1], 0, 32); err == nil {
					status = types.Status(n)
				} else {
					return errors.Wrap(err, "cannot parse status (uint32 expected)")
				}
			}

			res, err := queryClient.CheckStatus(
				context.Background(),
				&types.CheckStatusRequest{
					AccAddress: accAddress,
					Status:     status,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Result)
		},
	}
	util.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getValidateTransitionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate-transition <subject address> <destination address>",
		Aliases: []string{"validate_transition", "vt"},
		Args:    cobra.ExactArgs(2),
		Short:   "Check if the subject can be transferred under the new referrer",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			subject := args[0]
			target := args[1]

			res, err := queryClient.ValidateTransition(
				context.Background(),
				&types.ValidateTransitionRequest{
					Subject: subject,
					Target:  target,
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

func cmdAllWithStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all-with-status <n>",
		Aliases: []string{"all_with_status", "aws"},
		Short:   "Get all accounts with status #n (only for n ≥ 5)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var status types.Status
			if n, err := strconv.ParseUint(args[0], 0, 32); err == nil {
				status = types.Status(n)
			} else {
				return errors.Wrap(err, "cannot parse status (uint32 expected)")
			}

			res, err := queryClient.AllWithStatus(
				context.Background(),
				&types.AllWithStatusRequest{
					Status: status,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Accounts)
		},
	}
	util.AddQueryFlagsToCmd(cmd)
	return cmd
}
