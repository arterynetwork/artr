package cli

import (
	"github.com/arterynetwork/artr/util"
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"strconv"

	"github.com/arterynetwork/artr/x/subscription/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	subscriptionTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	subscriptionTxCmd.AddCommand(flags.PostCommands(
		GetPayForSubscriptionCmd(cdc),
		GetPayForVPNCmd(cdc),
		GetPayForStorageCmd(cdc),
		util.LineBreak(),
		GetSetTokenRateCmd(cdc),
	)...)

	return subscriptionTxCmd
}

func getPrices(cdc *codec.Codec, cliCtx *context.CLIContext) (*types.QueryPricesRes, error) {
	res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryPrices))
	if err != nil {
		return nil, err
	}

	var out types.QueryPricesRes
	cdc.MustUnmarshalJSON(res, &out)

	return &out, nil
}

// GetSetProfileCmd will create a send tx and sign it with the given key.
func GetPayForSubscriptionCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pay",
		Short: "Create and sign a pay for subscription tx",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			// Get prices
			prices, err := getPrices(cdc, &cliCtx)

			if err != nil {
				return err
			}

			//prices.VPN = prices.Subscription
			txBldr = txBldr.WithFees(
				util.CalculateFee(sdk.NewInt(prices.Subscription)).String() + util.ConfigMainDenom)

			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgPaySubscription(cliCtx.FromAddress, amount)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetSetProfileCmd will create a send tx and sign it with the given key.
func GetPayForVPNCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn [amount in bytes]",
		Short: "Create and sign a pay for vpn tx",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			// Get prices
			prices, err := getPrices(cdc, &cliCtx)

			if err != nil {
				return err
			}

			// fee
			txBldr = txBldr.WithFees(
				util.CalculateFee(sdk.NewInt(amount*prices.VPN/util.GBSize)).String() + util.ConfigMainDenom)

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgPayVPN(cliCtx.FromAddress, amount)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetPayForStorageCmd will create a send tx and sign it with the given key.
func GetPayForStorageCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage [amount in bytes]",
		Short: "Create and sign a pay for vpn tx",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			// Get prices
			prices, err := getPrices(cdc, &cliCtx)

			if err != nil {
				return err
			}

			// fee
			txBldr = txBldr.WithFees(
				util.CalculateFee(sdk.NewInt(amount*prices.Storage/util.GBSize)).String() + util.ConfigMainDenom)

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgPayStorage(cliCtx.FromAddress, amount)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetSetTokenRateCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-rate [0.01 USD to uARTRs]",
		Short: "Set ARTR exchange rate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			rate, err := strconv.ParseUint(args[0], 0, 32)
			if err != nil { return err }

			msg := types.NewMsgSetTokenRate(cliCtx.FromAddress, uint32(rate))
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}