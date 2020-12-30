package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/arterynetwork/artr/x/referral/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	referralTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"ref", "r"},
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	referralTxCmd.AddCommand(flags.PostCommands(
		getCmdRequestTransition(cdc),
		getCmdResolveTransition(cdc),
	)...)

	return referralTxCmd
}

func getCmdRequestTransition(cdc *codec.Codec) *cobra.Command {
	result := cobra.Command{
		Use:     "request-transition <address>",
		Aliases: []string{"rt", "transit", "trans"},
		Short:   "Request transition to another referrer",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			subjAddr := cliCtx.GetFromAddress()
			destAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}

			msg := types.NewMsgRequestTransition(subjAddr, destAddr)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return &result
}

func getCmdResolveTransition(cdc *codec.Codec) *cobra.Command {
	result := cobra.Command{
		Use:     "resolve-transition <address> [yes|no]",
		Aliases: []string{"resolve"},
		Short:   "Approve/decline a transition request (approve is default)",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			senderAddr := cliCtx.GetFromAddress()
			subjAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
			}

			approved := true
			if len(args) > 1 {
				switch strings.ToLower(args[1]) {
				case "yes", "y":
					approved = true
				case "no", "n":
					approved = false
				default:
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "cannot parse the second argument")
				}
			}

			msg := types.NewMsgResolveTransition(senderAddr, subjAddr, approved)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return &result
}
