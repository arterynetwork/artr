package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	//"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	//sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/noding/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group noding queries under a subcommand
	nodingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	nodingQueryCmd.AddCommand(
		flags.GetCommands(
			GetCmdStatus(queryRoute, cdc),
			GetCmdInfo(queryRoute, cdc),
			GetCmdProposer(queryRoute, cdc),
			GetCmdIsAllowed(queryRoute, cdc),
			GetCmdOperator(queryRoute, cdc),
		)...,
	)

	return nodingQueryCmd
}

func GetCmdStatus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "status [address]",
		Short: "query if noding's on",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					types.QueryStatus,
					accAddress,
				}, "/",
			))
			if err != nil {
				fmt.Printf("could not get noding status for address %s\n", accAddress)
				return nil
			}

			var out bool
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdInfo(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "info [address]",
		Short: "query validator info",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					types.QueryInfo,
					accAddress,
				}, "/",
			))
			if err != nil {
				if err == types.ErrNotFound {
					fmt.Println("no data")
				} else {
					fmt.Printf("could not get noding status for address %s\n", accAddress)
				}
				return nil
			}

			var out types.D
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdProposer(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "proposer [height]",
		Aliases: []string{"p"},
		Short:   "Get a block proposer account address. Height is optional, default is the last block.",
		Args:    cobra.MaximumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			path := []string{
				"custom",
				queryRoute,
				types.QueryProposer,
			}
			if len(args) > 0 { path = append(path, args[0]) }
			res, _, err := cliCtx.Query(strings.Join(path, "/"))
			if err != nil { return err }

			var out sdk.AccAddress = res
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdIsAllowed(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "is-allowed [address]",
		Short: "check is noding is allowed for an account",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					types.QueryAllowed,
					accAddress,
				}, "/",
			))
			if err != nil {
				fmt.Println("no data for account")
				return err
			}

			var out types.AllowedQueryRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdOperator(queryRoute string, cdc *codec.Codec) *cobra.Command {
	var hex bool

	result := &cobra.Command{
		Use: "whois [consensus address]",
		Short: "find account address by attached node consensus address",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]
			var format string
			if hex {
				format = types.QueryOperatorFormatHex
			} else {
				format = types.QueryOperatorFormatBech32
			}

			res, _, err := cliCtx.Query(strings.Join(
				[]string{
					"custom",
					queryRoute,
					types.QueryOperator,
					format,
					accAddress,
				}, "/",
			))
			if err != nil {
				if err == types.ErrNotFound {
					fmt.Println("no data")
				} else {
					fmt.Printf("could not get operator account for node %s:\n%s\n", accAddress, err.Error())
				}
				return nil
			}

			var out = sdk.AccAddress(res)
			return cliCtx.PrintOutput(out)
		},
	}

	result.Flags().BoolVarP(&hex, "hex", "x", false, "consensus address in hex format instead of bech32")

	return result
}
