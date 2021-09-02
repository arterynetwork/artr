package cli

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/earning/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	earningTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"e", "earn"},
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	earningTxCmd.AddCommand(
		getCmdListEarners(),
		getCmdRun(),
		getCmdReset(),
	)

	return earningTxCmd
}

func getCmdListEarners() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <signer_key_or_address> [[address] [vpn point] [storage points]...]",
		Short: "Add earners to the pending list",
		Args:  cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			earners := make([]types.Earner, len(args)/3)
			for i := 0; i < len(args)/3; i++ {
				var (
					address      sdk.AccAddress
					vpn, storage int64
					err          error
				)
				if address, err = sdk.AccAddressFromBech32(args[3*i+1]); err != nil {
					return errors.Wrapf(err, "invalid address #%d", i)
				}
				if vpn, err = strconv.ParseInt(args[3*i+2], 0, 64); err != nil {
					return errors.Wrapf(err, "invalid vpn points #%d", i)
				}
				if storage, err = strconv.ParseInt(args[3*i+3], 0, 64); err != nil {
					return errors.Wrapf(err, "invalid storage points #%d", i)
				}
				earners[i] = types.NewEarner(address, vpn, storage)
			}

			msg := types.NewMsgListEarners(clientCtx.GetFromAddress(), earners)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <signer_key_or_address> <fund_part> <accounts_per_block> <total_vpn_points> <total_storage_points> <time>",
		Short: "Lock earner list and schedule distribution for a specified block height",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				clientCtx    client.Context
				err          error
				fundPart     util.Fraction
				perBlock     uint32
				totalVpn     int64
				totalStorage int64
				time         *timestamp.Timestamp
			)

			if err = cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}
			if clientCtx, err = client.GetClientTxContext(cmd); err != nil {
				return err
			}

			if fundPart, err = util.ParseFraction(args[1]); err != nil {
				return err
			}
			if x, err := strconv.ParseInt(args[2], 0, 16); err != nil {
				return err
			} else {
				perBlock = uint32(x)
			}
			if totalVpn, err = strconv.ParseInt(args[3], 0, 64); err != nil {
				return err
			}
			if totalStorage, err = strconv.ParseInt(args[4], 0, 64); err != nil {
				return err
			}
			if time, err = runtime.Timestamp(fmt.Sprintf(`"%s"`, args[5])); err != nil {
				return err
			}

			msg := types.NewMsgRun(
				clientCtx.GetFromAddress(),
				fundPart,
				perBlock,
				totalVpn,
				totalStorage,
				time.AsTime(),
			)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func getCmdReset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset <signer_key_or_address>",
		Short: "Reset all data, unlock earner list",
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

			msg := types.NewMsgReset(clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
