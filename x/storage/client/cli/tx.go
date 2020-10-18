package cli

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/arterynetwork/artr/x/storage/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	storageTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	storageTxCmd.AddCommand(flags.PostCommands(
		GetSetStorageDataCmd(cdc),
	)...)

	return storageTxCmd
}

// GetSetProfileCmd will create a send tx and sign it with the given key.
func GetSetStorageDataCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data [size] [base64_data]",
		Short: "Create and sign a set data tx",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			size, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			bz, err := b64.StdEncoding.DecodeString(args[1])

			if len(bz) > 10*1024 {
				return types.ErrDataToLong
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgSetStorageData(cliCtx.FromAddress, size, args[1])
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
