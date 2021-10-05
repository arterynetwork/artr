package types

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/x/referral/types"
)

func (args *PriceArgs) Validate() error           { return nil }
func (args *DelegationAwardArgs) Validate() error { return args.Award.Validate() }
func (args *NetworkAwardArgs) Validate() error    { return args.Award.Validate() }
func (args *AddressArgs) Validate() error {
	_, err := sdk.AccAddressFromBech32(args.Address)
	return err
}
func (args *SoftwareUpgradeArgs) Validate() error {
	if args.Name == "" {
		return errors.New("empty upgrade name")
	}
	if args.Height > 0 {
		return errors.New("upgrade height is deprecated, use time instead")
	}
	if args.Time == nil {
		return errors.New("upgrade time is nil")
	}
	return nil
}
func (args *MinAmountArgs) Validate() error { return nil }
func (args *CountArgs) Validate() error     { return nil }
func (args *StatusArgs) Validate() error {
	if _, ok := types.Status_name[int32(args.Status)]; !ok || args.Status == types.STATUS_UNSPECIFIED {
		return errors.New("enum value out of range")
	}
	return nil
}
func (args *PeriodArgs) Validate() error {
	if args.Days < 1 {
		return errors.New("period must be at least one day")
	}
	return nil
}

func (args *AddressArgs) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(args.Address)
	if err != nil {
		panic(err)
	}
	return addr
}
