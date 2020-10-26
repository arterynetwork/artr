package cli

import (
	"github.com/arterynetwork/artr/util"
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"strings"

	"github.com/arterynetwork/artr/x/profile/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	profileTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	profileTxCmd.AddCommand(
		//flags.PostCommands(
		GetSetProfileCmd(cdc),
		//GetCreateAccountCmd(cdc),
		GetCreateAccountWithProfileCmd(cdc),
		// GetCmd<Action>(cdc)
		//)...
	)

	return profileTxCmd
}

// GetSetProfileCmd will create a send tx and sign it with the given key.
func GetSetProfileCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [from_key_or_address]",
		Short: "Create and sign a set profile tx",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// query current profile
			params := types.NewQueryProfileParams(addr)
			bz, err := cliCtx.Codec.MarshalJSON(params)

			res, _, _ := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/profile", "profile"), bz)

			var out types.QueryResProfile
			cdc.MustUnmarshalJSON(res, &out)

			profile := out.Profile

			if len(args) > 1 {
				for _, val := range args[1:] {
					com := strings.Split(strings.TrimSpace(val), ":")

					if len(com) != 2 {
						return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "invalid parameter string "+val)
					}

					switch strings.ToLower(com[0]) {
					case "nickname":
						profile.Nickname = com[1]
					case "autopay":
						profile.AutoPay = com[1] == "yes"
					case "noding":
						profile.Noding = com[1] == "yes"
					case "storage":
						profile.Storage = com[1] == "yes"
					case "validator":
						profile.Validator = com[1] == "yes"
					case "vpn":
						profile.VPN = com[1] == "yes"
					}

				}
			}

			//if len(strings.TrimSpace(out.Profile.Nickname)) != 0 {
			//	if strings.ToLower(out.Profile.Nickname) != strings.ToLower(profile.Nickname) {
			//		txBldr = txBldr.WithFees("1000000uartr")
			//	}
			//}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgSetProfile(addr, profile)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = flags.PostCommands(cmd)[0]

	return cmd
}

// GetSetProfileCmd will create a send tx and sign it with the given key.
func GetCreateAccountCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_account [from_key_or_address] [new_account_address] [referral_account_address]",
		Short: "Create and sign a create account tx",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			txBldr = txBldr.WithFees("1000000" + util.ConfigMainDenom)
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			newAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			refAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgCreateAccount(addr, newAddr, refAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = flags.PostCommands(cmd)[0]

	return cmd
}

func GetCreateAccountWithProfileCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_account_with_profile [from_key_or_address] [new_account_address] [referral_account_address] [params]",
		Short: "Create and sign a create account with profile tx",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			newAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			refAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			profile := types.Profile{}

			if len(args) > 3 {
				for _, val := range args[3:] {
					com := strings.Split(strings.TrimSpace(val), ":")

					if len(com) != 2 {
						return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "invalid parameter string "+val)
					}

					switch strings.ToLower(com[0]) {
					case "nickname":
						profile.Nickname = com[1]
					case "autopay":
						profile.AutoPay = com[1] == "yes"
					case "noding":
						profile.Noding = com[1] == "yes"
					case "storage":
						profile.Storage = com[1] == "yes"
					case "validator":
						profile.Validator = com[1] == "yes"
					case "vpn":
						profile.VPN = com[1] == "yes"
					}

				}
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgCreateAccountWithProfile(addr, newAddr, refAddr, profile)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = flags.PostCommands(cmd)[0]

	return cmd
}
