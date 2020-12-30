package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/arterynetwork/artr/x/bank/internal/keeper"
	"github.com/arterynetwork/artr/x/bank/internal/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group bank queries under a subcommand
	bankQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"d"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	bankQueryCmd.AddCommand(
		flags.GetCommands(
			getParamsCmd(queryRoute, cdc),
		)...,
	)

	return bankQueryCmd
}

func getParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "params",
		Aliases: []string{"p"},
		Short:   "Get module parameters",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					keeper.QueryParams,
				}, "/",
			))
			if err != nil {
				fmt.Println("could not get module parameters:", err)
				return err
			}

			var out types.QueryResParams
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
