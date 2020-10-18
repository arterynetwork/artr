package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/referral/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group referral queries under a subcommand
	referralQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"ref", "r"},
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	referralQueryCmd.AddCommand(
		flags.GetCommands(
			GetStatusCmd(queryRoute, cdc),
			GetReferrerCmd(queryRoute, cdc),
			GetReferralsCmd(queryRoute, cdc),
			GetCoinsCmd(queryRoute, cdc),
			GetDelegatedCoinsCmd(queryRoute, cdc),
			GetCheckStatusCmd(queryRoute, cdc),
		)...,
	)

	return referralQueryCmd
}

const customRoute = "custom"

func GetStatusCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "status [address]",
		Short: "Query for account status",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			data, _, err := clientCtx.Query(
				strings.Join([]string{
					customRoute,
					queryRoute,
					types.QueryStatus,
					accAddress,
				}, "/"),
			)

			if err != nil {
				fmt.Printf("could not get status of %s\n", accAddress)
			}
			var res types.Status
			cdc.MustUnmarshalJSON(data, &res)
			return clientCtx.PrintOutput(res)
		},
	}
}

func GetReferrerCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "referrer [address]",
		Short: "Get referrer's account address",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			data, _, err := clientCtx.Query(
				strings.Join([]string{
					customRoute,
					queryRoute,
					types.QueryReferrer,
					accAddress,
				}, "/"),
			)

			if err != nil {
				fmt.Printf("could not get referrer for %s\n", accAddress)
			}
			var res sdk.AccAddress
			cdc.MustUnmarshalJSON(data, &res)
			return clientCtx.PrintOutput(res)
		},
	}
}

func GetReferralsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "referrals [address]",
		Short: "Get list of referrals' account addresses",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			data, _, err := clientCtx.Query(
				strings.Join([]string{
					customRoute,
					queryRoute,
					types.QueryReferrals,
					accAddress,
				}, "/"),
			)

			if err != nil {
				fmt.Printf("could not get referrals for %s\n", accAddress)
			}
			var res []sdk.AccAddress
			cdc.MustUnmarshalJSON(data, &res)
			return clientCtx.PrintOutput(res)
		},
	}
}

func GetCoinsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "coins [address]",
		Short: "Get coins in one's network total",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			data, _, err := clientCtx.Query(
				strings.Join([]string{
					customRoute,
					queryRoute,
					types.QueryCoinsInNetwork,
					accAddress,
				}, "/"),
			)

			if err != nil {
				fmt.Printf("could not get coins total for %s\n", accAddress)
			}
			var res sdk.Int
			cdc.MustUnmarshalJSON(data, &res)
			return clientCtx.PrintOutput(res)
		},
	}
}

func GetDelegatedCoinsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "delegated [address]",
		Short: "Get delegated coins in one's network total",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]

			data, _, err := clientCtx.Query(
				strings.Join([]string{
					customRoute,
					queryRoute,
					types.QueryDelegatedInNetwork,
					accAddress,
				}, "/"),
			)

			if err != nil {
				fmt.Printf("could not get delegated coins total for %s\n", accAddress)
			}
			var res sdk.Int
			cdc.MustUnmarshalJSON(data, &res)
			return clientCtx.PrintOutput(res)
		},
	}
}

func GetCheckStatusCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: "check-status [address] [n]",
		Short: "Check if status #n requirements are fulfilled",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx  := context.NewCLIContext().WithCodec(cdc)
			accAddress := args[0]
			status     := args[1]

			data, _, err := clientCtx.Query(
				strings.Join([]string{
					customRoute,
					queryRoute,
					types.QueryCheckStatus,
					accAddress,
					status,
				}, "/"),
			)

			if err != nil {
				fmt.Printf("could not check %s for status %s", accAddress, status)
			}
			var res types.StatusCheckResult
			cdc.MustUnmarshalJSON(data, &res)
			return clientCtx.PrintOutput(res)
		},
	}
}