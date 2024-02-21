package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/delegating"
	noding "github.com/arterynetwork/artr/x/noding/types"
	referral "github.com/arterynetwork/artr/x/referral/types"
	"github.com/arterynetwork/artr/x/voting/types"
)

// GetTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	votingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	votingTxCmd.AddCommand(
		cmdEnterPrice(),
		cmdAddGovernor(),
		cmdRemoveGovernor(),
		cmdProductVpnBasePrice(),
		cmdProductStorageBasePrice(),
		cmdAddFreeCreator(),
		cmdRemoveFreeCreator(),
		cmdUpgradeSoftware(),
		cmdCancelSoftwareUpgrade(),
		cmdStaffValidatorAdd(),
		cmdStaffValidatorRemove(),
		cmdEarningSignerAdd(),
		cmdEarningSignerRemove(),
		cmdCourseChangeSignerAdd(),
		cmdCourseChangeSignerRemove(),
		cmdVpnCurrentSignerAdd(),
		cmdVpnCurrentSignerRemove(),
		cmdAccountTransitionPrice(),
		cmdSetMinSend(),
		cmdSetMinDelegate(),
		cmdSetMaxValidators(),
		cmdSetLotteryValidators(),
		cmdGeneralAmnesty(),
		cmdSetValidatorMinCriteria(),
		cmdSetJailAfter(),
		cmdSetRevokePeriod(),
		cmdSetDustDelegation(),
		cmdSetVotingPower(),
		cmdSetTransactionFee(),
		cmdSetBurnOnRevoke(),
		cmdSetMaxTransactionFee(),
		cmdSetTransactionFeeSplitRatios(),
		cmdSetAccruePercentageTable(),
		cmdAddBlockedSender(),
		cmdRemoveBlockedSender(),
		util.LineBreak(),
		cmdVote(),
		util.LineBreak(),
		cmdStartPoll(),
		cmdAnswerPoll(),
	)

	return votingTxCmd
}

func cmdEnterPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-subscription-price <price> <proposal name> <author key or address>",
		Aliases: []string{"set_subscription_price", "ssp"},
		Short:   "Propose to change the subscription price",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_ENTER_PRICE,
					Args: &types.Proposal_Price{
						Price: &types.PriceArgs{
							Price: price,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddGovernor is the CLI command for creating AddGovernor proposal
func cmdAddGovernor() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-governor <address> <proposal name> <author key or address>",
		Aliases: []string{"add_governor", "ag"},
		Short:   "Propose to add an account to the government",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_GOVERNMENT_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRemoveGovernor is the CLI command for creating Remove proposal
func cmdRemoveGovernor() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-governor <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_governor", "rg"},
		Short:   "Propose to remove an account from the government",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_GOVERNMENT_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdProductVpnBasePrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-vpn-gb-price <price> <proposal name> <author key or address>",
		Aliases: []string{"set_vpn_gb_price", "svgp"},
		Short:   "Propose to change the VPN base price",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_PRODUCT_VPN_BASE_PRICE,
					Args: &types.Proposal_Price{
						Price: &types.PriceArgs{
							Price: price,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdProductStorageBasePrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-storage-gb-price <price> <proposal name> <author key or address>",
		Aliases: []string{"set_storage_gb_price", "ssgp"},
		Short:   "Propose to change the storage base price",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_PRODUCT_STORAGE_BASE_PRICE,
					Args: &types.Proposal_Price{
						Price: &types.PriceArgs{
							Price: price,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdAddFreeCreator() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-free-creator <address> <proposal name> <author key or address>",
		Aliases: []string{"add_free_creator", "afc"},
		Short:   "Propose to allow an account to create new accounts for free",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_FREE_CREATOR_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdRemoveFreeCreator() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-free-creator <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_free_creator", "rfc"},
		Short:   "Propose to disallow an account to create new accounts for free",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_FREE_CREATOR_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpgradeSoftware is the CLI command for creating software upgrade proposal
func cmdUpgradeSoftware() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade-software <upgrade name> <time> <JSON URI with checksum> <proposal name> <author key or address>",
		Aliases: []string{"upgrade_software", "upgrade", "us"},
		Short:   "Propose to upgrade the blockchain software",
		Example: `artrcli tx voting upgrade-software 3.0.0 2023-01-01T03:00:00Z https://example.com/updates/3.0.0/info.json?checksum=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 "update to v3 Jan 1st at 03:00 AM GMT" ivan`,
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[4]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[3]

			upgradeName := args[0]

			var t time.Time
			if stamp, err := runtime.Timestamp(fmt.Sprintf(`"%s"`, args[1])); err != nil {
				return errors.Wrap(err, "cannot parse time")
			} else {
				t = stamp.AsTime()
			}

			info := args[2]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_SOFTWARE_UPGRADE,
					Args: &types.Proposal_SoftwareUpgrade{
						SoftwareUpgrade: &types.SoftwareUpgradeArgs{
							Name: upgradeName,
							Time: &t,
							Info: info,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpgradeSoftware is the CLI command for creating scheduled software upgrade cancellation proposal
func cmdCancelSoftwareUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-upgrade-software <proposal name> <author key or address>",
		Aliases: []string{"cancel-upgrade", "cancel_upgrade_software", "cus"},
		Short:   "Propose to cancel a previously scheduled blockchain software upgrade",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[1]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_CANCEL_SOFTWARE_UPGRADE,
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)

		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdStaffValidatorAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-staff-validator <address> <proposal name> <author key or address>",
		Aliases: []string{"add_staff_validator", "asv"},
		Short:   "Propose to allow an account to become a validator even if it doesn't fulfill requirements",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_STAFF_VALIDATOR_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdStaffValidatorRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-staff-validator <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_staff_validator", "rsv"},
		Short:   "Propose to disallow an account to be a validator if it doesn't fulfill requirements",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_STAFF_VALIDATOR_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdEarningSignerAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-earning-signer <address> <proposal name> <author key or address>",
		Aliases: []string{"add_earning_signer", "aes"},
		Short:   "Propose to allow an account to schedule VPN & storage awards",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_EARNING_SIGNER_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdEarningSignerRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-earning-signer <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_earning_signer", "res"},
		Short:   "Propose to disallow an account to schedule VPN & storage awards",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_EARNING_SIGNER_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdCourseChangeSignerAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-exchange-rate-signer <address> <proposal name> <author key or address>",
		Aliases: []string{"add_exchange_rate_signer", "axrs"},
		Short:   "Propose to allow an account to set token exchange rate",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_TOKEN_RATE_SIGNER_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdCourseChangeSignerRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-exchange-rate-signer <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_exchange_rate_signer", "rxrs"},
		Short:   "Propose to disallow an account to set token exchange rate",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_TOKEN_RATE_SIGNER_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdVpnCurrentSignerAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-vpn-current-signer <address> <proposal name> <author key or address>",
		Aliases: []string{"add_vpn_current_signer", "avcs"},
		Short:   "Propose to allow an account to update accounts' current VPN traffic value",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_VPN_SIGNER_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdVpnCurrentSignerRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-vpn-current-signer <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_vpn_current_signer", "rvcs"},
		Short:   "Propose to disallow an account to update accounts' current VPN traffic value",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_VPN_SIGNER_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdAccountTransitionPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-account-transition-price <price> <proposal name> <author key or address>",
		Example: `artrcli tx voting set-account-transition-price 2000000 "2 ARTR for transition" ivan`,
		Aliases: []string{"set_account_transition_price", "satp"},
		Short:   "Propose to change an account transition price (in uARTR)",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_TRANSITION_PRICE,
					Args: &types.Proposal_Price{
						Price: &types.PriceArgs{
							Price: price,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetMinSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-min-send <amount> <proposal name> <author key or address>",
		Example: `artrcli tx voting set-min-send 1000 "0.001 ARTR minimum" ivan`,
		Aliases: []string{"set_min_send", "sms"},
		Short:   "Propose to change minimum amount allowed to send (in uARTR)",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var n int64
			{
				var err error
				n, err = strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					return err
				}
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_MIN_SEND,
					Args: &types.Proposal_MinAmount{
						MinAmount: &types.MinAmountArgs{
							MinAmount: n,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetMinDelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-min-delegate <amount> <proposal name> <author key or address>",
		Example: `artrcli tx voting set-min-delegate 1000 "0.001 ARTR minimum" ivan`,
		Aliases: []string{"set_min_delegate", "smd"},
		Short:   "Propose to change minimum amount allowed to delegate (in uARTR)",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var n int64
			{
				var err error
				n, err = strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					return err
				}
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_MIN_DELEGATE,
					Args: &types.Proposal_MinAmount{
						MinAmount: &types.MinAmountArgs{
							MinAmount: n,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetMaxValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-max-validators <count> <proposal name> <author key or address>",
		Example: `artrcli tx voting set-max-validators 200 "let's double the count" ivan`,
		Aliases: []string{"set_max_validators", "smv"},
		Short:   "Propose to change maximum validator count",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var count uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				count = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_MAX_VALIDATORS,
					Args: &types.Proposal_Count{
						Count: &types.CountArgs{
							Count: count,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetLotteryValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-lottery-validators <count> <proposal name> <author key or address>",
		Example: `artrcli tx voting set-lottery-validators 20 "lucky 20" ivan`,
		Aliases: []string{"set_lottery_validators", "slv"},
		Short:   `Propose to change the count of "lucky" (aka "lottery") validators`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var count uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				count = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_LUCKY_VALIDATORS,
					Args: &types.Proposal_Count{
						Count: &types.CountArgs{
							Count: count,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdGeneralAmnesty() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "general-amnesty <proposal name> <author key or address>",
		Aliases: []string{"general_amnesty"},
		Short:   "Zero all users' missed block count and jail count",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[1]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_GENERAL_AMNESTY,
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetValidatorMinCriteria() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-validator-min-criteria <status (number)> <self stake> <total stake> <proposal name> <author key or address>",
		Aliases: []string{"set_validator_min_criteria", "svmc"},
		Short:   `Propose to set minimal criteria required for validation (stakes set in uARTR)`,
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[4]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[3]

			n, err := strconv.ParseUint(args[0], 0, 8)
			if err != nil {
				return err
			}
			status := referral.Status(n)

			selfStake, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			totalStake, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_VALIDATOR_MINIMAL_CRITERIA,
					Args: &types.Proposal_MinCriteria{
						MinCriteria: &types.MinCriteriaArgs{
							MinCriteria: &noding.MinCriteria{
								Status:     status,
								SelfStake:  selfStake,
								TotalStake: totalStake,
							},
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetJailAfter() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-jail-after <count> <proposal name> <author key or address>",
		Aliases: []string{"set_jail_after", "sja"},
		Short:   `Propose to set a number of blocks, a validator is jailed after missing which in row`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var count uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				count = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_JAIL_AFTER,
					Args: &types.Proposal_Count{
						Count: &types.CountArgs{
							Count: count,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetDustDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-dust-delegation <amount> <proposal name> <author key or address>",
		Example: `artrd tx voting set-dust-delegation 849999 "Ignore delegation lower than 0.85 ARTR" ivan`,
		Aliases: []string{"set_dust_delegation", "sdd"},
		Short:   "Propose to change dust delegation threshold (in uARTR, an exactly equal delegation counts as dust)",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var n int64
			{
				var err error
				n, err = strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					return err
				}
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_DUST_DELEGATION,
					Args: &types.Proposal_MinAmount{
						MinAmount: &types.MinAmountArgs{
							MinAmount: n,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetRevokePeriod() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-revoke-period <days> <proposal name> <author key or address>",
		Aliases: []string{"set_revoke_period", "srp"},
		Short:   `Set a number of days, coins are returned from delegation after`,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			var days uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				days = uint32(n)
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_REVOKE_PERIOD,
					Args: &types.Proposal_Period{
						Period: &types.PeriodArgs{
							Days: days,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetVotingPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:     `set-voting-power <part>:<voting power> [<part>:<voting power> [...]] <"luckies" voting power> <proposal name> <author key or address>`,
		Example: `artrd tx voting set-voting-power 15%:3 85%:2 2 "reduce voting power" ivan`,
		Aliases: []string{"set_voting_power", "svp"},
		Short:   "Propose to change validator voting power distribution",
		Args:    cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[len(args)-1]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[len(args)-2]

			value := noding.Distribution{}

			n, err := strconv.ParseInt(args[len(args)-3], 0, 64)
			if err != nil {
				return errors.Wrap(err, `cannot parse "luckies" voting power`)
			}
			value.LuckiesVotingPower = n

			for i := 0; i < len(args)-3; i++ {
				parts := strings.Split(args[i], ":")
				if len(parts) != 2 {
					return errors.Errorf("cannot parse the slice #%d: exactly one colon expected", i)
				}
				f, err := util.ParseFraction(parts[0])
				if err != nil {
					return errors.Wrapf(err, "cannot parse the slice #%d: invalid part", i)
				}
				n, err = strconv.ParseInt(parts[1], 0, 64)
				if err != nil {
					return errors.Wrapf(err, "cannot parse the slice #%d: invalid power", i)
				}
				value.Slices = append(value.Slices, noding.Distribution_Slice{Part: f, VotingPower: n})
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_VOTING_POWER,
					Args: &types.Proposal_VotingPower{
						VotingPower: &value,
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetTransactionFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-transaction-fee <amount> <proposal name> <author key or address>",
		Example: `artrd tx voting set-transaction-fee 3/1000 "Set fee of 0.3%" ivan`,
		Aliases: []string{"set_transaction_fee", "stf"},
		Short:   "Propose to change transaction fee",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			q, err := util.ParseFraction(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_TRANSACTION_FEE,
					Args: &types.Proposal_Portion{
						Portion: &types.PortionArgs{
							Fraction: q,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetBurnOnRevoke() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-burn-on-revoke <amount> <proposal name> <author key or address>",
		Example: `artrd tx voting set-burn-on-revoke 5% "Set burn on revoke percent of 5%" ivan`,
		Aliases: []string{"set_burn_on_revoke", "sbr"},
		Short:   "Propose to change burn on revoke percent",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			q, err := util.ParseFraction(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_BURN_ON_REVOKE,
					Args: &types.Proposal_Portion{
						Portion: &types.PortionArgs{
							Fraction: q,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetMaxTransactionFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-max-transaction-fee <amount> <proposal name> <author key or address>",
		Example: `artrd tx voting set-max-transaction-fee 10_000000 "Set max fee of 10_000000" ivan`,
		Aliases: []string{"set_max_transaction_fee", "smtf"},
		Short:   "Propose to change max transaction fee (in uARTR)",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]

			q, err := strconv.ParseInt(args[0], 0, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_MAX_TRANSACTION_FEE,
					Args: &types.Proposal_MinAmount{
						MinAmount: &types.MinAmountArgs{
							MinAmount: q,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetTransactionFeeSplitRatios() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-transaction-fee-split-ratios <for proposal ratio> <for company ratio> <proposal name> <author key or address>",
		Example: `artrd tx voting set-transaction-fee-split-ratios 2/5 1/5 "Set transaction fee split ratios: 2 part for proposal, 1 part for company and 2 part for burning" ivan`,
		Aliases: []string{"set_transaction_fee_split_ratios", "stfsr"},
		Short:   "Propose to change transaction fee split ratios (in fractions for proposal and company)",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[3]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[2]

			forProposal, err := util.ParseFraction(args[0])
			if err != nil {
				return err
			}

			forCompany, err := util.ParseFraction(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_TRANSACTION_FEE_SPLIT_RATIOS,
					Args: &types.Proposal_Portions{
						Portions: &types.PortionsArgs{
							Fractions: []util.Fraction{forProposal, forCompany},
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdSetAccruePercentageTable() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-accrue-percentage-table <range start in uartr>:<base percentage>:<validator bonus percentage>:<subscription bonus percentage>:<vpn bonus percentage>:<storage bonus percentage> [...] <proposal name> <author key or address>",
		Example: `artrcli tx voting set-accrue-percentage-table 0:21%:2%:1%:4%:4% 1000000000:24%:2%:1%:3%:3% 10000000000:27%:2%:1%:2%:2% 100000000000:30%:2%:1%:1%:1% "return to default values" ivan`,
		Aliases: []string{"set_accrue_percentage_table", "sapt"},
		Short:   "Propose to change an award for delegating funds in five categories: base, validator bonus, subscription bonus, vpn bonus, storage bonus",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[len(args)-1]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[len(args)-2]

			value := []delegating.PercentageListRange(nil)

			for i := 0; i < len(args)-2; i++ {
				parts := strings.Split(args[i], ":")
				if len(parts) != 6 {
					return errors.Errorf("cannot parse the range #%d: exactly five colon expected", i+1)
				}
				n, err := strconv.ParseUint(parts[0], 0, 64)
				if err != nil {
					return errors.Wrapf(err, "cannot parse the range #%d: invalid range start", i+1)
				}
				fl := []util.Fraction(nil)
				for j := 1; j < len(parts); j++ {
					f, err := util.ParseFraction(parts[j])
					if err != nil {
						return errors.Wrapf(err, "cannot parse the range #%d: invalid percentage #%d", i+1, j)
					}
					fl = append(fl, f)
				}
				value = append(value, delegating.PercentageListRange{Start: n, PercentList: fl})
			}

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_ACCRUE_PERCENTAGE_TABLE,
					Args: &types.Proposal_AccruePercentageTable{
						AccruePercentageTable: &types.AccruePercentageTableArgs{
							AccruePercentageTable: value,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdAddBlockedSender() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-blocked-sender <address> <proposal name> <author key or address>",
		Aliases: []string{"add_blocked_sender", "abs"},
		Short:   "Propose to add an account to the blocked senders list",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_BLOCKED_SENDER_ADD,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdRemoveBlockedSender() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-blocked-sender <address> <proposal name> <author key or address>",
		Aliases: []string{"remove_blocked_sender", "rbs"},
		Short:   "Propose to remove an account from the blocked senders list",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[2]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			author := clientCtx.GetFromAddress().String()
			proposalName := args[1]
			addr := args[0]

			msg := &types.MsgPropose{
				Proposal: types.Proposal{
					Author: author,
					Name:   proposalName,
					Type:   types.PROPOSAL_TYPE_BLOCKED_SENDER_REMOVE,
					Args: &types.Proposal_Address{
						Address: &types.AddressArgs{
							Address: addr,
						},
					},
				},
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote agree|disagree <voter_key_or_address>",
		Short: "Vote for/against the current proposal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[1]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			voter := clientCtx.GetFromAddress().String()
			agree := strings.ToLower(args[0]) == "agree"
			if !agree && strings.ToLower(args[0]) != "disagree" {
				return errors.New("cannot parse aggree/disagree flag")
			}

			msg := &types.MsgVote{
				Voter: voter,
				Agree: agree,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdStartPoll() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start-poll <author_key_or_address> validators|status:<status> <name> <text> [quorum]",
		Aliases: []string{"start_poll", "sp"},
		Short:   "Start a public poll",
		Example: `start-poll ivan validators Halving "Should we decrease all awards by a half next Monday?" 2/3`,
		Args:    cobra.RangeArgs(4, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			poll := types.Poll{
				Name:     args[2],
				Author:   clientCtx.GetFromAddress().String(),
				Question: args[3],
			}

			if req := args[1]; req == "validators" {
				poll.Requirements = &types.Poll_CanValidate{CanValidate: &types.Poll_Unit{}}
			} else if m := regexp.MustCompile(`/^status:(\d+)|([A-Za-z_]+)$/`).FindStringSubmatch(req); m != nil {
				var status referral.Status
				if len(m[1]) > 0 {
					if s, err := strconv.Atoi(m[1]); err != nil {
						return errors.Wrap(err, "cannot parse status")
					} else {
						status = referral.Status(s)
					}
				} else {
					name := strings.ToUpper(m[2])
					if !strings.HasPrefix(name, "STATUS_") {
						name = "STATUS_" + name
					}
					if status, err = referral.ParseStatus(name); err != nil {
						return err
					}
					if err = status.Validate(); err != nil {
						return err
					}
				}

				poll.Requirements = &types.Poll_MinStatus{MinStatus: status}
			} else {
				return errors.New("cannot parse requirements")
			}

			if len(args) > 4 {
				if q, err := util.ParseFraction(args[4]); err != nil {
					return errors.Wrap(err, "cannot parse quorum")
				} else {
					poll.Quorum = &q
				}
			}

			msg := types.MsgStartPoll{Poll: poll}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}

func cmdAnswerPoll() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "answer yes|no <respondent_key_or_address>",
		Aliases: []string{"ans", "a", "answer-poll", "answer_poll"},
		Short:   "Answer the current public poll",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[1]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var yes bool
			if ans := strings.ToLower(args[0]); ans == "yes" {
				yes = true
			} else if ans != "no" {
				return errors.New("cannot parse answer")
			}

			msg := types.MsgAnswerPoll{
				Respondent: clientCtx.GetFromAddress().String(),
				Yes:        yes,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	util.AddTxFlagsToCmd(cmd)
	return cmd
}
