package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/noding/types"
)

// GetQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group noding queries under a subcommand
	nodingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"n"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nodingQueryCmd.AddCommand(
		cmdInfo(),
		cmdState(),
		cmdProposer(),
		cmdIsAllowed(),
		cmdOperator(),
		util.LineBreak(),
		cmdSwitchedOn(),
		cmdQueue(),
		util.LineBreak(),
		cmdParams(),
	)

	return nodingQueryCmd
}

func cmdInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info <address>",
		Aliases: []string{"i"},
		Short:   "query validator info",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.GetRequest{
				Account: args[0],
			}

			res, err := queryClient.Get(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Info)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdState() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "state <address>",
		Aliases: []string{"s"},
		Short:   `query validator state (is it in the "main" or "lucky" set or in reserve, is it jailed or so on)`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.StateRequest{
				Account: args[0],
			}

			res, err := queryClient.State(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.State)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdProposer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "proposer [height]",
		Aliases: []string{"p"},
		Short:   "Get a block proposer account address. Height is optional, default is the last block.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.ProposerRequest{}

			if len(args) > 0 {
				h, err := strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					return errors.Wrap(err, "cannot parse height")
				}
				req.Height = h
			}

			res, err := queryClient.Proposer(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Account)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdIsAllowed() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "is-allowed <address>",
		Aliases: []string{"ia", "a"},
		Short:   "check if noding is allowed for an account",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.IsAllowedRequest{
				Account: args[0],
			}

			res, err := queryClient.IsAllowed(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdOperator() *cobra.Command {
	var hex bool

	cmd := &cobra.Command{
		Use:     "whois <consensus address>",
		Aliases: []string{"w"},
		Short:   "find account address by attached node consensus address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			var format types.OperatorRequest_Format
			if hex {
				format = types.OperatorRequest_FORMAT_HEX
			} else {
				format = types.OperatorRequest_FORMAT_BECH32
			}
			req := &types.OperatorRequest{
				ConsAddress: args[0],
				Format:      format,
			}

			res, err := queryClient.Operator(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Account)
		},
	}

	cmd.Flags().BoolVarP(&hex, "hex", "x", false, "consensus address in hex format instead of bech32")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get the module params",
		Args:  cobra.NoArgs,
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
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdSwitchedOn() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "switched-on",
		Aliases: []string{"so", "on"},
		Short:   "Get the list of validators that are switched on and not jailed",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.SwitchedOnRequest{}

			res, err := queryClient.SwitchedOn(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Accounts)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func cmdQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "queue",
		Aliases: []string{"q"},
		Short:   `Get the list of "lucky" and "spare" validators`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueueRequest{}

			res, err := queryClient.Queue(context.Background(), req)
			if err != nil {
				return err
			}
			return util.PrintConsoleOutput(clientCtx, res.Queue)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
