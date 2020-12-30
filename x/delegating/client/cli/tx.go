package cli

import (
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	delegatingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"d"},
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	delegatingTxCmd.AddCommand(flags.PostCommands(
		GetCmdDelegate(cdc),
		GetCmdRevoke(cdc),
	)...)

	return delegatingTxCmd
}

func GetCmdDelegate(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "delegate <microARTRs>",
		Aliases: []string{"d"},
		Short:   "delegate funds",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				inBuf  = bufio.NewReader(cmd.InOrStdin())
				cliCtx = context.NewCLIContext().WithCodec(cdc)
				txBldr = auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

				err    error
				amount uint64
				msg    sdk.Msg
			)

			_, err = fmt.Sscan(args[0], &amount)
			if err != nil {
				return err
			}

			msg = types.NewMsgDelegate(cliCtx.FromAddress, sdk.NewIntFromUint64(amount))
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdRevoke(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "revoke <microARTRs>",
		Aliases: []string{"r", "u"},
		Short:   "revoke funds from delegating",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				inBuf  = bufio.NewReader(cmd.InOrStdin())
				cliCtx = context.NewCLIContext().WithCodec(cdc)
				txBldr = auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

				err    error
				amount uint64
				msg    sdk.Msg
			)

			_, err = fmt.Sscan(args[0], &amount)
			if err != nil {
				return err
			}

			msg = types.NewMsgRevoke(cliCtx.FromAddress, sdk.NewIntFromUint64(amount))
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
