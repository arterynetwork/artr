package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/delegating/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	delegatingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"d"},
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	delegatingTxCmd.AddCommand(
		GetCmdDelegate(),
		GetCmdRevoke(),
	)

	return delegatingTxCmd
}

func GetCmdDelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delegate <key_or_address> <microARTRs>",
		Aliases: []string{"d"},
		Short:   "delegate funds",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var amount uint64
			_, err = fmt.Sscan(args[1], &amount)
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegate(clientCtx.GetFromAddress(), sdk.NewIntFromUint64(amount))
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetCmdRevoke() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke <key_or_address> <microARTRs>",
		Aliases: []string{"r", "u"},
		Short:   "revoke funds from delegating",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var amount uint64
			_, err = fmt.Sscan(args[1], &amount)
			if err != nil {
				return err
			}

			msg := types.NewMsgRevoke(clientCtx.GetFromAddress(), sdk.NewIntFromUint64(amount))
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
