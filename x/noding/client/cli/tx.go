package cli

import (
	"fmt"
	"bufio"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/arterynetwork/artr/x/noding/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	nodingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nodingTxCmd.AddCommand(flags.PostCommands(
		GetCmdOn(cdc),
		GetCmdOff(cdc),
		GetCmdUnjail(cdc),
	)...)

	return nodingTxCmd
}

func GetCmdOn(cdc *codec.Codec) *cobra.Command {
	result := cobra.Command{
		Use: "on [public key]",
		Short: "Switch noding on",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf  := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			accAddr := cliCtx.GetFromAddress()
			pubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, args[0])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("cannot parse public key: %s", args[0]))
			}

			msg := types.NewMsgSwitchOn(accAddr, pubKey)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return &result
}

func GetCmdOff(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "off [public key]",
		Short: "Switch noding off",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf  := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			accAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgSwitchOff(accAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdUnjail(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "unjail",
		Short: "Unjail (won't work if jail period isn't over)",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf  := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			accAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgUnjail(accAddr)
			if err:= msg.ValidateBasic(); err != nil { return err }

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}