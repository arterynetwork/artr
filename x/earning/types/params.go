package types

import (
	"fmt"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter store keys
var (
	// default paramspace keys
	KeySigners = []byte("Signers")

	// state paramspace keys
	KeyLocked           = []byte("Locked")
	KeyVpnPointCost     = []byte("VpnCost")
	KeyStoragePointCost = []byte("StorageCost")
	KeyItemsPerBlock    = []byte("PerBlock")
)

// ParamKeyTable for earning module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().
		RegisterParamSet(&Params{}).
		RegisterParamSet(&StateParams{})
}

func NewParams(signers []sdk.AccAddress) *Params {
	res := &Params{
		Signers: make([]string, len(signers)),
	}
	for i, acc := range signers {
		res.Signers[i] = acc.String()
	}
	return res
}

func (p Params) GetSigners() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.Signers))
	for i, bech32 := range p.Signers {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p Params) String() string {
	str, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(str)
}

// NewParams creates a new Params object
func NewStateLocked(vpnPointCost util.Fraction, storagePointCost util.Fraction, itemsPerBlock uint32) StateParams {
	return StateParams{
		Locked:           true,
		VpnPointCost:     vpnPointCost,
		StoragePointCost: storagePointCost,
		ItemsPerBlock:    itemsPerBlock,
	}
}

func NewStateUnlocked() StateParams {
	return StateParams{}
}

// String implements the stringer interface for Params
func (p StateParams) String() string {
	str, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(str)
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySigners, &p.Signers, validateSigners),
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *StateParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyLocked, &p.Locked, validateLocked),
		paramtypes.NewParamSetPair(KeyVpnPointCost, &p.VpnPointCost, validatePointCost),
		paramtypes.NewParamSetPair(KeyStoragePointCost, &p.StoragePointCost, validatePointCost),
		paramtypes.NewParamSetPair(KeyItemsPerBlock, &p.ItemsPerBlock, validateItemsPerBlock),
	}
}

//
//// DefaultParams defines the parameters for this module
//func DefaultParams() Params {
//	return NewParamsUnlocked()
//}

func (p Params) Validate() error {
	if err := validateSigners(p.Signers); err != nil {
		return err
	}
	return nil
}

func validateSigners(value interface{}) error {
	accz, ok := value.([]string)
	if !ok {
		return fmt.Errorf("unexpected Signers type: %T", value)
	}
	if len(accz) == 0 {
		return fmt.Errorf("signers list is empty")
	}
	for i, acc := range accz {
		if _, err := sdk.AccAddressFromBech32(acc); err != nil {
			return errors.Wrapf(err, "invalid signer address (#%d)", i)
		}
	}
	return nil
}

func (p StateParams) Validate() error {
	if err := validateLocked(p.Locked); err != nil {
		return err
	}
	if err := validatePointCost(p.VpnPointCost); err != nil {
		return err
	}
	if p.Locked && p.VpnPointCost.IsNullValue() {
		return errors.New("missing VpnPointCost")
	}
	if err := validatePointCost(p.StoragePointCost); err != nil {
		return err
	}
	if p.Locked && p.StoragePointCost.IsNullValue() {
		return errors.New("missing StoragePointCost")
	}
	if err := validateItemsPerBlock(p.ItemsPerBlock); err != nil {
		return err
	}
	if p.Locked && p.ItemsPerBlock < 1 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("items per block number must be positive: %d", p.ItemsPerBlock))
	}
	return nil
}

func validateLocked(value interface{}) error {
	if _, ok := value.(bool); !ok {
		return fmt.Errorf("unexpected Locked type: %T", value)
	}
	return nil
}

func validatePointCost(value interface{}) error {
	q, ok := value.(util.Fraction)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unexpected point cost type: %T", value))
	}
	if !q.IsNullValue() && q.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("point cost must be non-negative: %v", q))
	}
	return nil
}

func validateItemsPerBlock(value interface{}) error {
	_, ok := value.(uint32)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unexpected item count type: %T", value))
	}
	// Value can be empty if Locked is false
	//if n < 1 { return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("items per block number must be positive: %d", n)) }
	return nil
}
