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
	"github.com/arterynetwork/artr/x/profile/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group profile queries under a subcommand
	profileQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"p"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	profileQueryCmd.AddCommand(
		getProfileCmd(),
		getAccountByNicknameCmd(),
		getAccountByCardNumberCmd(),
		util.LineBreak(),
		getCmdParams(),
	)

	return profileQueryCmd
}

// getProfileCmd returns a query profile that will display the state of the
// profile at a given address.
func getProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <address>",
		Short: "Query profile info for address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			addr := args[0]

			res, err := queryClient.GetByAddress(
				context.Background(),
				&types.GetByAddressRequest{
					Address: addr,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res.Profile)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getAccountByNicknameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-by-nickname <nickname>",
		Aliases: []string{"account_by_nickname"},
		Short:   "Query account address by profile nickname",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			nick := args[0]

			res, err := queryClient.GetByNickname(
				context.Background(),
				&types.GetByNicknameRequest{
					Nickname: nick,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getAccountByCardNumberCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-by-card-number <card number>",
		Aliases: []string{"account_by_card_number"},
		Short:   "Query account address by profile card number",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			cardNo, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return errors.Wrap(err, "cannot parse card number")
			}

			res, err := queryClient.GetByCardNumber(
				context.Background(),
				&types.GetByCardNumberRequest{
					CardNumber: cardNo,
				},
			)
			if err != nil {
				return err
			}

			return util.PrintConsoleOutput(clientCtx, res)
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
