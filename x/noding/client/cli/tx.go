package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/noding/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	nodingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nodingTxCmd.AddCommand(
		cmdOn(),
		cmdOff(),
		cmdUnjail(),
	)

	return nodingTxCmd
}

func cmdOn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "on <from key or address> <node public key>",
		Short: "Switch noding on",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgOn{
				Account: args[0],
				PubKey:  args[1],
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdOff() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "off <from key or address>",
		Short: "Switch noding off",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgOff{
				Account: args[0],
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdUnjail() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail <from key or address>",
		Short: "Unjail (won't work if jail period isn't over)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUnjail{
				Account: args[0],
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}
