package cli

import (
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"strconv"

	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group profile queries under a subcommand
	profileQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	profileQueryCmd.AddCommand(
		//flags.GetCommands(
		GetProfileCmd(queryRoute, cdc),
		GetAccountByNicknameCmd(queryRoute, cdc),
		GetAccountByCardNumberCmd(queryRoute, cdc),
		GetCreatorsCmd(queryRoute, cdc),
		//)...,
	)

	return profileQueryCmd
}

// GetCreatorsCmd returns accounts who can create another accounts for free
func GetCreatorsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creators",
		Short: "Query for free accounts creators",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			params := types.QueryCreatorsParams{}
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCreators), bz)
			if err != nil {
				return err
			}

			var out types.QueryCreatorsRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return flags.GetCommands(cmd)[0]
}

// GetProfileCmd returns a query profile that will display the state of the
// profile at a given address.
func GetProfileCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [address]",
		Short: "Query profile info for address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryProfileParams(addr)
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/profile", queryRoute), bz)
			if err != nil {
				fmt.Printf("could not find profile for address- %s \n", addr)
				return nil
			}

			var out types.QueryResProfile
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return flags.GetCommands(cmd)[0]
}

func GetAccountByNicknameCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-by-nickname [nickname]",
		Aliases: []string{"account_by_nickname"},
		Short:   "Query account address by profile nickname",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			params := types.NewQueryAccountByNicknameParams(args[0])
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAccountAddressByNickname), bz)

			if err != nil {
				return errors.New("could not find address for nickname: " + err.Error())
			}

			var out types.QueryResAccountBy
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return flags.GetCommands(cmd)[0]
}

func GetAccountByCardNumberCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-by-card-number [card number]",
		Aliases: []string{"account_by_card_number"},
		Short:   "Query account address by profile card number",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			cardNumber, err := strconv.ParseUint(args[0], 10, 64)

			if err != nil {
				return errors.New("invalid card number format: only decimal digits accepted")
			}

			params := types.NewQueryAccountByCardNumberParams(cardNumber)
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAccountAddressByCardNumber), bz)
			if err != nil {
				return err
				//fmt.Printf("could not find address for card number- %s \n", args[1])
				//return nil
			}

			var out types.QueryResAccountBy
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return flags.GetCommands(cmd)[0]
}
