package cli

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/voting/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	votingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	votingTxCmd.AddCommand(flags.PostCommands(
		GetCmdEnterPrice(cdc),
		GetCmdDelegationAward(cdc),
		GetCmdDelegationNetworkAward(cdc),
		GetCmdSubscriptionNetworkAward(cdc),
		GetCmdAddGovernor(cdc),
		GetCmdRemoveGovernor(cdc),
		GetCmdProductVpnBasePrice(cdc),
		GetCmdProductStorageBasePrice(cdc),
		GetCmdAddFreeCreator(cdc),
		GetCmdRemoveFreeCreator(cdc),
		GetCmdUpgradeSoftware(cdc),
		GetCmdCancelSoftwareUpgrade(cdc),
		GetCmdStaffValidatorAdd(cdc),
		GetCmdStaffValidatorRemove(cdc),
		GetCmdEarningSignerAdd(cdc),
		GetCmdEarningSignerRemove(cdc),
		GetCmdCourseChangeSignerAdd(cdc),
		GetCmdCourseChangeSignerRemove(cdc),
		GetCmdVpnCurrentSignerAdd(cdc),
		GetCmdVpnCurrentSignerRemove(cdc),
		getCmdAccountTransitionPrice(cdc),
		getCmdSetMinSend(cdc),
		getCmdSetMinDelegate(cdc),
		getCmdSetMaxValidators(cdc),
		getCmdSetLotteryValidators(cdc),
		getCmdGeneralAmnesty(cdc),
		getCmdSetValidatorMinStatus(cdc),
		util.LineBreak(),
		GetCmdVote(cdc),
	)...)

	return votingTxCmd
}

func GetCmdEnterPrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-subscription-price <price> <proposal name>",
		Aliases: []string{"set_subscription_price", "ssp"},
		Short:   "Propose to change the subscription price",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			params := types.PriceProposalParams{Price: price}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeEnterPrice,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdDelegationAward(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-delegation-award <minimal> <1K+> <10K+> <100K+> <proposal name>",
		Aliases: []string{"set_delegation_award", "sda"},
		Short:   "Propose to change an award for delegating funds",
		Example: `artrcli tx voting set-delegation-award 21 24 27 30 "return to default values" --from ivan`,
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[4]

			var percent [4]uint8
			for i := 0; i < 4; i++ {
				n, err := strconv.ParseUint(args[i], 0, 8)
				if err != nil {
					return err
				}
				percent[i] = uint8(n)
			}

			params := types.DelegationAwardProposalParams{
				Minimal:      percent[0],
				ThousandPlus: percent[1],
				TenKPlus:     percent[2],
				HundredKPlus: percent[3],
			}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeDelegationAward,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdDelegationNetworkAward(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-delegation-network-award <company> <lvl 1> <lvl 2> ... <lvl 10> <proposal name>",
		Aliases: []string{"set_delegation_network_award", "sdna"},
		Short:   "Propose to change the network commission for delegations",
		Example: `artrcli tx voting set-delegation-network-award 5/1000 5% 1% 1% 2% 1% 1% 1% 1% 1% 5/1000 "return to default values" --from ivan`,
		Args:    cobra.ExactArgs(12),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[11]

			params, err := parseNetworkAward(args[:11])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeDelegationNetworkAward,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdSubscriptionNetworkAward(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-subscription-network-award <company> <lvl 1> <lvl 2> ... <lvl 10> <proposal name>",
		Aliases: []string{"set_subscription_network_award", "ssna"},
		Short:   "Propose to change the network commission for subscription payments",
		Example: `artrcli tx voting set-subscription-network-award 10% 15% 10% 7% 7% 7% 7% 7% 5% 2% 2% "return to default values" --from ivan`,
		Args:    cobra.ExactArgs(12),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[11]

			params, err := parseNetworkAward(args[:11])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeProductNetworkAward,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func parseNetworkAward(args []string) (types.NetworkAwardProposalParams, error) {
	var percent [11]util.Fraction
	for i := 0; i < 11; i++ {
		n, err := util.ParseFraction(args[i])
		if err != nil {
			return types.NetworkAwardProposalParams{}, err
		}
		percent[i] = n
	}

	params := types.NetworkAwardProposalParams{Award: referral.NetworkAward{Company: percent[0]}}
	copy(params.Award.Network[:], percent[1:])

	return params, nil
}

// GetCmdAddGovernor is the CLI command for creating AddGovernor proposal
func GetCmdAddGovernor(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-governor <address> <proposal name>",
		Aliases: []string{"add_governor", "ag"},
		Short:   "Propose to add an account to the government",
		Args:    cobra.ExactArgs(2), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				args[1],
				types.ProposalTypeGovernmentAdd,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdRemoveGovernor is the CLI command for creating Remove proposal
func GetCmdRemoveGovernor(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "remove-governor <address> <proposal name>",
		Aliases: []string{"remove_governor", "rg"},
		Short:   "Propose to remove an account from the government",
		Args:    cobra.ExactArgs(2), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				args[1],
				types.ProposalTypeGovernmentRemove,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdProductVpnBasePrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-vpn-gb-price <price> <proposal name>",
		Aliases: []string{"set_vpn_gb_price", "svgp"},
		Short:   "Propose to change the VPN base price",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			params := types.PriceProposalParams{Price: price}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeProductVpnBasePrice,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdProductStorageBasePrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-storage-gb-price <price> <proposal name>",
		Aliases: []string{"set_storage_gb_price", "ssgp"},
		Short:   "Propose to change the storage base price",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			params := types.PriceProposalParams{Price: price}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeProductStorageBasePrice,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdAddFreeCreator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-free-creator <address> <proposal name>",
		Aliases: []string{"add_free_creator", "afc"},
		Short:   "Propose to allow an account to create new accounts for free",
		Args:    cobra.ExactArgs(2), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				args[1],
				types.ProposalTypeAddFreeCreator,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdRemoveFreeCreator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "remove-free-creator <address> <proposal name>",
		Aliases: []string{"remove_free_creator", "rfc"},
		Short:   "Propose to disallow an account to create new accounts for free",
		Args:    cobra.ExactArgs(2), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				args[1],
				types.ProposalTypeRemoveFreeCreator,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdUpgradeSoftware is the CLI command for creating software upgrade proposal
func GetCmdUpgradeSoftware(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "upgrade-software <upgrade name> <height> <JSON URI with checksum> <proposal name>",
		Aliases: []string{"upgrade_software", "upgrade", "us"},
		Short:   "Propose to upgrade the blockchain software",
		Example: `artrcli tx voting upgrade-software v2.0 1000 https://example.com/updates/v2.0/info.json?checksum=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 "update to v2.0 on block height 1000" --from ivan`,
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[3]

			upgradeName := args[0]
			height, err := strconv.ParseInt(args[1], 0, 64)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid height "+args[1])
			}
			info := args[2]
			params := types.SoftwareUpgradeProposalParams{
				Name:   upgradeName,
				Height: height,
				Info:   info,
			}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeSoftwareUpgrade,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdUpgradeSoftware is the CLI command for creating scheduled software upgrade cancellation proposal
func GetCmdCancelSoftwareUpgrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "cancel-upgrade-software <proposal name>",
		Aliases: []string{"cancel-upgrade", "cancel_upgrade_software", "cus"},
		Short:   "Propose to cancel a previously scheduled blockchain software upgrade",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[0]

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeCancelSoftwareUpgrade,
				types.EmptyProposalParams{},
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})

		},
	}
}

func GetCmdStaffValidatorAdd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-staff-validator <address> <proposal name>",
		Aliases: []string{"add_staff_validator", "asv"},
		Short:   "Propose to allow an account to become a validator even if it doesn't fulfill requirements",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeStaffValidatorAdd,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdStaffValidatorRemove(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "remove-staff-validator <address> <proposal name>",
		Aliases: []string{"remove_staff_validator", "rsv"},
		Short:   "Propose to disallow an account to be a validator if it doesn't fulfill requirements",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeStaffValidatorRemove,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdEarningSignerAdd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-earning-signer <address> <proposal name>",
		Aliases: []string{"add_earning_signer", "aes"},
		Short:   "Propose to allow an account to schedule VPN & storage awards",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeEarningSignerAdd,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdEarningSignerRemove(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "remove-earning-signer <address> <proposal name>",
		Aliases: []string{"remove_earning_signer", "res"},
		Short:   "Propose to disallow an account to schedule VPN & storage awards",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeEarningSignerRemove,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdCourseChangeSignerAdd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-exchange-rate-signer <address> <proposal name>",
		Aliases: []string{"add_exchange_rate_signer", "axrs"},
		Short:   "Propose to allow an account to set token exchange rate",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeRateChangeSignerAdd,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdCourseChangeSignerRemove(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "remove-exchange-rate-signer <address> <proposal name>",
		Aliases: []string{"remove_exchange_rate_signer", "rxrs"},
		Short:   "Propose to disallow an account to set token exchange rate",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeRateChangeSignerRemove,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVpnCurrentSignerAdd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-vpn-current-signer <address> <proposal name>",
		Aliases: []string{"add_vpn_current_signer", "avcs"},
		Short:   "Propose to allow an account to update accounts' current VPN traffic value",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeVpnCurrentSignerAdd,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVpnCurrentSignerRemove(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "remove-vpn-current-signer <address> <proposal name>",
		Aliases: []string{"remove_vpn_current_signer", "rvcs"},
		Short:   "Propose to disallow an account to update accounts' current VPN traffic value",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.AddressProposalParams{Address: addr}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeVpnCurrentSignerRemove,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdAccountTransitionPrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-account-transition-price <price> <proposal name>",
		Example: `artrcli tx voting set-account-transition-price 2000000 "2 ARTR for transition" --from ivan`,
		Aliases: []string{"set_account_transition_price", "satp"},
		Short:   "Propose to change an account transition price (in uARTR)",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var price uint32
			{
				n, err := strconv.ParseUint(args[0], 0, 32)
				if err != nil {
					return err
				}
				price = uint32(n)
			}

			params := types.PriceProposalParams{Price: price}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeTransitionCost,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdSetMinSend(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-min-send <amount> <proposal name>",
		Example: `artrcli tx voting set-min-send 1000 "0.001 ARTR minimum" --from ivan`,
		Aliases: []string{"set_min_send", "sms"},
		Short:   "Propose to change minimum amount allowed to send (in uARTR)",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var n int64
			{
				var err error
				n, err = strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					return err
				}
			}

			params := types.MinAmountProposalParams{MinAmount: n}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeMinSend,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdSetMinDelegate(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-min-delegate <amount> <proposal name>",
		Example: `artrcli tx voting set-min-delegate 1000 "0.001 ARTR minimum" --from ivan`,
		Aliases: []string{"set_min_delegate", "smd"},
		Short:   "Propose to change minimum amount allowed to delegate (in uARTR)",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var n int64
			{
				var err error
				n, err = strconv.ParseInt(args[0], 0, 64)
				if err != nil {
					return err
				}
			}

			params := types.MinAmountProposalParams{MinAmount: n}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeMinDelegate,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdSetMaxValidators(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-max-validators <count> <proposal name>",
		Example: `artrcli tx voting set-max-validators 200 "let's double the count" --from ivan`,
		Aliases: []string{"set_max_validators", "smv"},
		Short:   "Propose to change maximum validator count",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var count uint16
			{
				n, err := strconv.ParseUint(args[0], 0, 16)
				if err != nil {
					return err
				}
				count = uint16(n)
			}

			params := types.ShortCountProposalParams{Count: count}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeMaxValidators,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdSetLotteryValidators(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-lottery-validators <count> <proposal name>",
		Example: `artrcli tx voting set-lottery-validators 20 "lucky 20" --from ivan`,
		Aliases: []string{"set_lottery_validators", "slv"},
		Short:   `Propose to change the count of "lucky" (aka "lottery") validators`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]

			var count uint16
			{
				n, err := strconv.ParseUint(args[0], 0, 16)
				if err != nil {
					return err
				}
				count = uint16(n)
			}

			params := types.ShortCountProposalParams{Count: count}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeLotteryValidators,
				params,
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdGeneralAmnesty(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "general-amnesty <proposal name>",
		Aliases: []string{"general_amnesty"},
		Short:   "Zero all users' missed block count and jail count",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[0]

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeGeneralAmnesty,
				types.EmptyProposalParams{},
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdSetValidatorMinStatus(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-validator-min-status <status (number)> <proposal name>",
		Aliases: []string{"set_validator_min_status", "svms"},
		Short:   `Propose to set minimal status required for validation`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			proposalName := args[1]


			var status uint8
			{
				n, err := strconv.ParseUint(args[0], 0, 8)
				if err != nil {
					return err
				}
				status = uint8(n)
			}

			msg := types.NewMsgCreateProposal(
				cliCtx.GetFromAddress(),
				proposalName,
				types.ProposalTypeValidatorMinimalStatus,
				types.StatusProposalParams{Status: status},
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdVote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "vote [agree/disagree]",
		Short: "Vote for/against the current proposal",
		Args:  cobra.ExactArgs(1), // Does your request require arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgProposalVote(
				cliCtx.GetFromAddress(),
				strings.ToLower(args[0]) == "agree",
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
