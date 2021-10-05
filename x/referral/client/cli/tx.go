package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	referralTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"ref", "r"},
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	referralTxCmd.AddCommand(
		getCmdRequestTransition(),
		getCmdResolveTransition(),
	)

	return referralTxCmd
}

func getCmdRequestTransition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "request-transition <subject_key_or_address> <new_referrer_address>",
		Aliases: []string{"rt", "transit", "trans"},
		Short:   "Request transition to another referrer",
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
			msg := types.NewMsgRequestTransition(clientCtx.GetFromAddress().String(), args[1])
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdResolveTransition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resolve-transition <signer_key_or_address> <subject_address> [yes|no]",
		Aliases: []string{"resolve"},
		Short:   "Approve/decline a transition request (approve is default)",
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddr := clientCtx.GetFromAddress().String()
			subjAddr := args[1]
			approved := true
			if len(args) > 2 {
				switch strings.ToLower(args[2]) {
				case "yes", "y":
					approved = true
				case "no", "n":
					approved = false
				default:
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot parse the 3rd argument")
				}
			}

			msg := types.NewMsgResolveTransition(senderAddr, subjAddr, approved)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}
