package types

import (
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (p Proposal) GetAuthor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Author)
	if err != nil {
		panic(err)
	}
	return addr
}

func (p Proposal) String() string {
	bz, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

func (p Proposal) Validate() error {
	if p.Name == "" {
		return errors.New("invalid name: empty string")
	}
	if _, err := sdk.AccAddressFromBech32(p.Author); err != nil {
		return errors.Wrap(err, "invalid author")
	}
	switch p.Type {
	case
		PROPOSAL_TYPE_CANCEL_SOFTWARE_UPGRADE,
		PROPOSAL_TYPE_GENERAL_AMNESTY:

		if p.Args != nil {
			return errors.New("args unexpected")
		}
	case
		PROPOSAL_TYPE_ENTER_PRICE,
		PROPOSAL_TYPE_PRODUCT_VPN_BASE_PRICE,
		PROPOSAL_TYPE_PRODUCT_STORAGE_BASE_PRICE,
		PROPOSAL_TYPE_TRANSITION_PRICE:

		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_Price expected")
		}
		if args, ok := p.Args.(*Proposal_Price); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_Price expected", p.Args)
		} else {
			if err := args.Price.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case PROPOSAL_TYPE_DELEGATION_AWARD:
		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_DelegationAward expected")
		}
		if args, ok := p.Args.(*Proposal_DelegationAward); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_DelegationAward expected", p.Args)
		} else {
			if err := args.DelegationAward.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case
		PROPOSAL_TYPE_DELEGATION_NETWORK_AWARD,
		PROPOSAL_TYPE_PRODUCT_NETWORK_AWARD:

		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_NetworkAward expected")
		}
		if args, ok := p.Args.(*Proposal_NetworkAward); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_NetworkAward expected", p.Args)
		} else {
			if err := args.NetworkAward.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case
		PROPOSAL_TYPE_GOVERNMENT_ADD,
		PROPOSAL_TYPE_GOVERNMENT_REMOVE,
		PROPOSAL_TYPE_FREE_CREATOR_ADD,
		PROPOSAL_TYPE_FREE_CREATOR_REMOVE,
		PROPOSAL_TYPE_STAFF_VALIDATOR_ADD,
		PROPOSAL_TYPE_STAFF_VALIDATOR_REMOVE,
		PROPOSAL_TYPE_EARNING_SIGNER_ADD,
		PROPOSAL_TYPE_EARNING_SIGNER_REMOVE,
		PROPOSAL_TYPE_TOKEN_RATE_SIGNER_ADD,
		PROPOSAL_TYPE_TOKEN_RATE_SIGNER_REMOVE,
		PROPOSAL_TYPE_VPN_SIGNER_ADD,
		PROPOSAL_TYPE_VPN_SIGNER_REMOVE,
		PROPOSAL_TYPE_STORAGE_SIGNER_ADD,
		PROPOSAL_TYPE_STORAGE_SIGNER_REMOVE:

		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_Address expected")
		}
		if args, ok := p.Args.(*Proposal_Address); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_Address expected", p.Args)
		} else {
			if err := args.Address.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case PROPOSAL_TYPE_SOFTWARE_UPGRADE:
		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_SoftwareUpgrade expected")
		}
		if args, ok := p.Args.(*Proposal_SoftwareUpgrade); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_SoftwareUpgrade expected", p.Args)
		} else {
			if err := args.SoftwareUpgrade.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case
		PROPOSAL_TYPE_MIN_SEND,
		PROPOSAL_TYPE_MIN_DELEGATE:

		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_MinAmount expected")
		}
		if args, ok := p.Args.(*Proposal_MinAmount); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_MinAmount expected", p.Args)
		} else {
			if err := args.MinAmount.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case
		PROPOSAL_TYPE_MAX_VALIDATORS,
		PROPOSAL_TYPE_LUCKY_VALIDATORS:

		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_Count expected")
		}
		if args, ok := p.Args.(*Proposal_Count); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_Count expected", p.Args)
		} else {
			if err := args.Count.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	case
		PROPOSAL_TYPE_JAIL_AFTER:

		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_Count expected")
		}
		if args, ok := p.Args.(*Proposal_Count); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_Count expected", p.Args)
		} else {
			if err := args.Count.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
			if args.Count.Count <= 0 {
				return errors.New("positive number expected")
			}
		}
	case PROPOSAL_TYPE_VALIDATOR_MINIMAL_STATUS:
		if p.Args == nil {
			return errors.New("invalid args: nil, *Proposal_Status expected")
		}
		if args, ok := p.Args.(*Proposal_Status); !ok {
			return errors.Errorf("invalid args: %T, *Proposal_Status expected", p.Args)
		} else {
			if err := args.Status.Validate(); err != nil {
				return errors.Wrap(err, "invalid args")
			}
		}
	default:
		return errors.Errorf("invalid type: %s", p.Type)
	}
	return nil
}

func (g Government) GetMembers() []sdk.AccAddress {
	addrz := make([]sdk.AccAddress, len(g.Members))
	for i, bech32 := range g.Members {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		addrz[i] = addr
	}
	return addrz
}

func (g Government) GetMember(i int) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(g.Members[i])
	if err != nil {
		panic(err)
	}
	return addr
}

func (g Government) String() string {
	bz, err := yaml.Marshal(g.Members)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

func (g Government) Strings() []string { return g.Members }

func (g Government) Contains(addr sdk.AccAddress) bool {
	bech32 := addr.String()
	for _, elem := range g.Members {
		if elem == bech32 {
			return true
		}
	}

	return false
}

func (g *Government) Remove(addr sdk.AccAddress) {
	bech32 := addr.String()
	for index, elem := range g.Members {
		if elem == bech32 {
			g.Members = append(g.Members[:index], g.Members[index+1:]...)
			return
		}
	}
}

func (g *Government) Append(addr sdk.AccAddress) {
	g.Members = append(g.Members, addr.String())
}

func (r ProposalHistoryRecord) GetGovernment() *Government {
	return &Government{Members: r.Government}
}

func (r ProposalHistoryRecord) GetAgreed() *Government {
	return &Government{Members: r.Agreed}
}

func (r ProposalHistoryRecord) GetDisagreed() *Government {
	return &Government{Members: r.Disagreed}
}

func (r ProposalHistoryRecord) Validate() error {
	if err := r.Proposal.Validate(); err != nil {
		return errors.Wrap(err, "invalid proposal")
	}
	if r.Government == nil {
		return errors.New("invalid government: empty list")
	}
	for i, bech32 := range r.Government {
		if _, err := sdk.AccAddressFromBech32(bech32); err != nil {
			return errors.Wrapf(err, "invalid government (item #%d)", i)
		}
	}
	for i, bech32 := range r.Agreed {
		if _, err := sdk.AccAddressFromBech32(bech32); err != nil {
			return errors.Wrapf(err, "invalid agreed (item #%d)", i)
		}
	}
	for i, bech32 := range r.Disagreed {
		if _, err := sdk.AccAddressFromBech32(bech32); err != nil {
			return errors.Wrapf(err, "invalid disagreed (item #%d)", i)
		}
	}
	if r.Started <= 0 {
		return errors.New("invalid started: must be positive")
	}
	if r.Finished <= 0 {
		return errors.New("invalid finished: must be positive")
	}
	return nil
}
