package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/vpn/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group vpn queries under a subcommand
	vpnQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	vpnQueryCmd.AddCommand(
		flags.GetCommands(
			GetVPNStatusCmd(queryRoute, cdc),
			GetVPNLimitCmd(queryRoute, cdc),
			GetVPNCurrentCmd(queryRoute, cdc),
		)...,
	)

	return vpnQueryCmd
}

// GetVPNStatusCmd returns a query profile that will display the state of the
// profile at a given address.
func GetVPNStatusCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [address]",
		Short: "Query vpn status for address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryVpnParams(addr)
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/query_state", queryRoute), bz)
			if err != nil {
				fmt.Printf("could not find data for address- %s \n", addr)
				return err
			}

			var out types.QueryResState
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}

func GetVPNLimitCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "limit [address]",
		Short: "Query vpn limit for address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryVpnParams(addr)
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/query_limit", queryRoute), bz)
			if err != nil {
				fmt.Printf("could not find data for address- %s \n", addr)
				return err
			}

			var out types.QueryResLimit
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}

func GetVPNCurrentCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current [address]",
		Short: "Query vpn current traffic for address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryVpnParams(addr)
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/query_current", queryRoute), bz)
			if err != nil {
				fmt.Printf("could not find data for address- %s \n", addr)
				return err
			}

			var out types.QueryResCurrent
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}
