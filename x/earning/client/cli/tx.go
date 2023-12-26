package cli

import (
	"fmt"
	"strings"
	"time"

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
		getCmdSet(),
	)

	return earningTxCmd
}

func getCmdSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <signer_key_or_address> <address;vpn_timestamp;storage_timestamp>...",
		Short: "Set or delete earners info by account addresses",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			earners := make([]types.Earner, len(args)-1)
			for i := 0; i < len(args)-1; i++ {
				var (
					address      sdk.AccAddress
					vpn, storage *time.Time
					err          error
				)
				parts := strings.Split(args[i+1], ";")
				if len(parts) != 3 {
					return errors.Errorf("cannot parse the earners #%d: exactly two semicolon expected", i)
				}
				if address, err = sdk.AccAddressFromBech32(parts[0]); err != nil {
					return errors.Wrapf(err, "invalid address #%d", i)
				}
				if parts[1] != "" {
					var tStamp *timestamp.Timestamp
					if tStamp, err = runtime.Timestamp(fmt.Sprintf(`"%s"`, parts[1])); err != nil {
						return errors.Wrapf(err, "invalid vpn timestamp #%d", i)
					} else {
						t := tStamp.AsTime()
						vpn = &t
					}
				}
				if parts[2] != "" {
					var tStamp *timestamp.Timestamp
					if tStamp, err = runtime.Timestamp(fmt.Sprintf(`"%s"`, parts[2])); err != nil {
						return errors.Wrapf(err, "invalid storage timestamp #%d", i)
					} else {
						t := tStamp.AsTime()
						storage = &t
					}
				}
				earners[i] = types.NewEarner(address, vpn, storage)
			}

			msg := types.NewMsgSetMultiple(clientCtx.GetFromAddress(), earners)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}
