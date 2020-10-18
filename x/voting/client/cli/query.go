package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/arterynetwork/artr/x/voting/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group voting queries under a subcommand
	votingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	votingQueryCmd.AddCommand(
		flags.GetCommands(
			GetQueryGovernmentCmd(queryRoute, cdc),
			GetQueryCurrentCmd(queryRoute, cdc),
			GetQueryStatusCmd(queryRoute, cdc),
			GetQueryHistoryCmd(queryRoute, cdc),
		)...,
	)

	return votingQueryCmd
}

func GetQueryHistoryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Query history of voting",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			data := cdc.MustMarshalJSON(types.QueryHistoryParams{
				Limit: int32(viper.GetInt(FlagLimit)),
				Page:  int32(viper.GetInt(FlagPage)),
			})

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryHistory), data)
			if err != nil {
				return err
			}

			var out []types.ProposalHistoryRecord
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	cmd.Flags().Int(FlagLimit, FlagLimitDefault, "Query number of history records per page returned")
	cmd.Flags().Int(FlagPage, FlagPageDefault, "Query a specific page of paginated results")

	return cmd
}

func GetQueryGovernmentCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "government",
		Short: "Query government accounts list",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGovernment))
			if err != nil {
				return err
			}

			var out types.QueryGovernmentRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}

func GetQueryCurrentCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Query current proposal",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCurrent))
			if err != nil {
				return err
			}

			var out types.QueryCurrentRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}

func GetQueryStatusCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query current proposal status - proposal info, voters list, votes",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryStatus))
			if err != nil {
				return err
			}

			var out types.QueryStatusRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}
