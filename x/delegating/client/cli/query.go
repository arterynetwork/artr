package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/spf13/cobra"
	"strings"

	"github.com/arterynetwork/artr/x/delegating/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group delegating queries under a subcommand
	delegatingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"d"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	delegatingQueryCmd.AddCommand(
		flags.GetCommands(
			GetCmdRevoking(queryRoute, cdc),
			GetCmdAccumulation(queryRoute, cdc),
		)...,
	)

	return delegatingQueryCmd
}

func GetCmdRevoking(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoking [address]",
		Short: "how many coins are being revoked from delegating",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					types.QueryRevoking,
					accAddress,
				}, "/",
			))
			if err != nil {
				fmt.Printf("could not get revoke requests for address %s\n:%s\n", accAddress, err.Error())
				return nil
			}

			var out types.QueryResRevoking
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdAccumulation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "accum [address]",
		Short: "Info about next payment accumulation progress",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					types.QueryAccumulation,
					accAddress,
				}, "/",
			))
			if err != nil {
				fmt.Printf("could not get accumulation progress for address %s:\n%s\n", accAddress, err.Error())
				return nil
			}

			var out types.QueryResAccumulation
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}