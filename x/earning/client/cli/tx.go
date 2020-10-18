package cli

import (
	"github.com/arterynetwork/artr/util"
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"strconv"

	//"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	//"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	//sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/cosmos/cosmos-sdk/x/auth"
	//"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/arterynetwork/artr/x/earning/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	earningTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"e", "earn"},
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	earningTxCmd.AddCommand(flags.PostCommands(
		GetCmdListEarners(cdc),
		GetCmdRun(cdc),
		GetCmdReset(cdc),
	)...)

	return earningTxCmd
}

func GetCmdListEarners(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "list [[address] [vpn point] [storage points]...]",
		Short: "Add earners to the pending list",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			earners := make([]types.Earner, 0, len(args)/3)
			for i := 0; i < len(args)/3; i++ {
				var (
					address sdk.AccAddress
					vpn, storage int64
					err error
				)
				if address, err = sdk.AccAddressFromBech32(args[3*i]); err != nil { return err }
				if vpn, err     = strconv.ParseInt(args[3*i+1], 0, 64); err != nil { return err }
				if storage, err = strconv.ParseInt(args[3*i+2], 0, 64); err != nil { return err }
				earners = append(earners, types.Earner{
					Points:  types.Points{
						Vpn:     vpn,
						Storage: storage,
					},
					Account: address,
				})
			}

			msg := types.NewMsgListEarners(cliCtx.GetFromAddress(), earners)
			if err := msg.ValidateBasic(); err != nil { return err }

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdRun(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "run [fund_part] [accounts_per_block] [total_vpn_points] [total_storage_points] [height]",
		Short: "Lock earner list and schedule distribution for a specified block height",
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			var (
				fundPart     util.Fraction
				perBlock     uint16
				totalVpn     int64
				totalStorage int64
				height       int64
				err          error
			)
			if fundPart, err = util.ParseFraction(args[0]); err != nil { return err }
			if x, err := strconv.ParseInt(args[1], 0, 16); err != nil { return err } else { perBlock = uint16(x) }
			if totalVpn, err = strconv.ParseInt(args[2], 0, 64); err != nil { return err }
			if totalStorage, err = strconv.ParseInt(args[3], 0, 64); err != nil { return err }
			if height, err = strconv.ParseInt(args[4], 0, 64); err != nil { return err }

			msg := types.NewMsgRun(cliCtx.GetFromAddress(), fundPart, perBlock, totalVpn, totalStorage, height)
			if err := msg.ValidateBasic(); err != nil { return err }

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdReset(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "reset",
		Short: "Reset all data, unlock earner list",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgReset(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil { return err }

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
