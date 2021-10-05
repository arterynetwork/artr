package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/profile/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	profileTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	profileTxCmd.AddCommand(
		cmdCreateAccount(),
		cmdUpdateProfile(),
		cmdStorageCurrent(),
		cmdVpnCurrent(),

		cmdPayTariff(),
		cmdBuyStorage(),
		cmdGiveUpStorage(),
		cmdBuyVpn(),
		cmdSetRate(),
	)

	return profileTxCmd
}

func cmdCreateAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create_account <from_key_or_address> <new_account_address> <referral_account_address> [[nickname:<nickname>] [autopay:yes|no] [noding:yes|no] [storage:yes|no] [validator:yes|no] [vpn:yes|no]]",
		Aliases: []string{"create-account", "ca", "c"},
		Short:   "Create and sign a create account with profile tx",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateAccount{
				Creator:  clientCtx.GetFromAddress().String(),
				Address:  args[1],
				Referrer: args[2],
			}

			if len(args) > 3 {
				profile := types.Profile{}
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
						profile.Vpn = com[1] == "yes"
					}
				}
				msg.Profile = &profile
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

// cmdUpdateProfile will create a send tx and sign it with the given key.
func cmdUpdateProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <from_key_or_address> [nickname:<nickname>] [autopay:yes|no] [noding:yes|no] [storage:yes|no] [validator:yes|no] [vpn:yes|no]",
		Aliases: []string{"u"},
		Short:   "Create and sign a set profile tx",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			addr := clientCtx.GetFromAddress()

			msg := &types.MsgUpdateProfile{
				Address: addr.String(),
				Updates: make([]types.MsgUpdateProfile_Update, 0, 6),
			}
			if len(args) > 1 {
				for _, val := range args[1:] {
					com := strings.Split(strings.TrimSpace(val), ":")

					if len(com) != 2 {
						return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "invalid parameter string "+val)
					}

					switch strings.ToLower(com[0]) {
					case "nickname":
						msg.Updates = append(msg.Updates, types.MsgUpdateProfile_Update{
							Field: types.MsgUpdateProfile_Update_FIELD_NICKNAME,
							Value: &types.MsgUpdateProfile_Update_String_{
								String_: com[1],
							},
						})
					case "autopay":
						msg.Updates = append(msg.Updates, types.MsgUpdateProfile_Update{
							Field: types.MsgUpdateProfile_Update_FIELD_AUTO_PAY,
							Value: &types.MsgUpdateProfile_Update_Bool{
								Bool: com[1] == "yes",
							},
						})
					case "noding":
						msg.Updates = append(msg.Updates, types.MsgUpdateProfile_Update{
							Field: types.MsgUpdateProfile_Update_FIELD_NODING,
							Value: &types.MsgUpdateProfile_Update_Bool{
								Bool: com[1] == "yes",
							},
						})
					case "storage":
						msg.Updates = append(msg.Updates, types.MsgUpdateProfile_Update{
							Field: types.MsgUpdateProfile_Update_FIELD_STORAGE,
							Value: &types.MsgUpdateProfile_Update_Bool{
								Bool: com[1] == "yes",
							},
						})
					case "validator":
						msg.Updates = append(msg.Updates, types.MsgUpdateProfile_Update{
							Field: types.MsgUpdateProfile_Update_FIELD_VALIDATOR,
							Value: &types.MsgUpdateProfile_Update_Bool{
								Bool: com[1] == "yes",
							},
						})
					case "vpn":
						msg.Updates = append(msg.Updates, types.MsgUpdateProfile_Update{
							Field: types.MsgUpdateProfile_Update_FIELD_VPN,
							Value: &types.MsgUpdateProfile_Update_Bool{
								Bool: com[1] == "yes",
							},
						})
					}

				}
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdStorageCurrent() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "storage_current <from_key_or_address> <address> <value>",
		Aliases: []string{"storage-current", "sc"},
		Short:   "Update current Artery Storage consumption",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			value, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return errors.Wrap(err, "cannot parse value")
			}

			msg := &types.MsgSetStorageCurrent{
				Sender:  sender.String(),
				Address: args[1],
				Value:   value,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdVpnCurrent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn_current <from_key_or_address> <address> <value>",
		Short: "Update current Artery VPN traffic usage",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			value, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return errors.Wrap(err, "cannot parse value")
			}

			msg := &types.MsgSetVpnCurrent{
				Sender:  sender.String(),
				Address: args[1],
				Value:   value,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdPayTariff() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pay_tariff <from_key_or_address> <storage_GBs>",
		Aliases: []string{"pay", "pt", "p"},
		Short:   "Pay for subscription",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			var storage uint32
			if n, err := strconv.ParseUint(args[1], 0, 32); err != nil {
				return errors.Wrap(err, "cannot parse value")
			} else {
				storage = uint32(n)
			}

			msg := &types.MsgPayTariff{
				Address:       sender.String(),
				StorageAmount: storage,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdBuyStorage() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "buy_storage <from_key_or_address> <extra_GBs>",
		Aliases: []string{"buy-storage", "bs"},
		Short:   "Buy some additional storage amount",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			var storage uint32
			if n, err := strconv.ParseUint(args[1], 0, 32); err != nil {
				return errors.Wrap(err, "cannot parse value")
			} else {
				storage = uint32(n)
			}

			msg := &types.MsgBuyStorage{
				Address:      sender.String(),
				ExtraStorage: storage,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdGiveUpStorage() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "give_up_storage <from_key_or_address> <GBs>",
		Aliases: []string{"give-up-storage", "gs"},
		Short:   "Give up some of unused storage amount. No refunds",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			var storage uint32
			if n, err := strconv.ParseUint(args[1], 0, 32); err != nil {
				return errors.Wrap(err, "cannot parse value")
			} else {
				storage = uint32(n)
			}

			msg := &types.MsgGiveStorageUp{
				Address: sender.String(),
				Amount:  storage,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdBuyVpn() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "buy_vpn <from_key_or_address> <extra_GBs>",
		Aliases: []string{"buy-vpn", "bv"},
		Short:   "Buy some additional VPN traffic",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			var traffic uint32
			if n, err := strconv.ParseUint(args[1], 0, 32); err != nil {
				return errors.Wrap(err, "cannot parse value")
			} else {
				traffic = uint32(n)
			}

			msg := &types.MsgBuyVpn{
				Address:      sender.String(),
				ExtraTraffic: traffic,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set_rate <from_key_or_address> <value>",
		Aliases: []string{"set-rate", "sr"},
		Short:   "Set the coin rate (how much uARTRs does 0.01 USD cost)",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			var traffic uint32
			if n, err := strconv.ParseUint(args[1], 0, 32); err != nil {
				return errors.Wrap(err, "cannot parse value")
			} else {
				traffic = uint32(n)
			}

			msg := &types.MsgBuyVpn{
				Address:      sender.String(),
				ExtraTraffic: traffic,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	util.AddTxFlagsToCmd(cmd)
	return cmd
}
