package cli

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/bank/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		CmdSend(),
		CmdBurn(),
	)
	return txCmd
}

// SendTxCmd will create a send tx and sign it with the given key.
func CmdSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <from_key_or_address> <to_address> <amount>",
		Short: "Create and sign a send tx",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			toAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "burn <from_key_or_address> <uartr_amount>",
		Short:   "Burn some uARTRs",
		Example: "  artrd tx burn ivan 1000000",
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

			amount, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return errors.Wrap(err, "cannot parse amount")
			}

			msg := types.MsgBurn{
				Account: clientCtx.GetFromAddress().String(),
				Amount:  amount,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}
