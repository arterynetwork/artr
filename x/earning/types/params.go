package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"

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
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().
		RegisterParamSet(&Params{}).
		RegisterParamSet(&StateParams{})
}


// Params - used for initializing default parameter for earning at genesis
type Params struct {
	Signers []sdk.AccAddress `json:"signers"`
}

// StateParams - used for storing keeper inner state and exporting it to genesis if needed
type StateParams struct {
	Locked           bool           `json:"locked,omitempty"`
	VpnPointCost     util.Fraction  `json:"vpn_point_cost,omitempty"`
	StoragePointCost util.Fraction  `json:"storage_point_cost,omitempty"`
	ItemsPerBlock    uint16         `json:"items_per_block,omitempty"`
}

// NewParams creates a new Params object
func NewStateLocked(vpnPointCost util.Fraction, storagePointCost util.Fraction, itemsPerBlock uint16) StateParams {
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

func (p Params) String() string { return fmt.Sprintf("Signers: %v", p.Signers) }

// String implements the stringer interface for Params
func (p StateParams) String() string {
	return fmt.Sprintf(`
Locked: %t
VpnPointCost: %v
StoragePointCost: %v
ItemsPerBlock: %d
`, p.Locked, p.VpnPointCost, p.StoragePointCost, p.ItemsPerBlock)
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeySigners, &p.Signers, validateSigners),
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *StateParams) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyLocked, &p.Locked, validateLocked),
		params.NewParamSetPair(KeyVpnPointCost, &p.VpnPointCost, validatePointCost),
		params.NewParamSetPair(KeyStoragePointCost, &p.StoragePointCost, validatePointCost),
		params.NewParamSetPair(KeyItemsPerBlock, &p.ItemsPerBlock, validateItemsPerBlock),
	}
}
//
//// DefaultParams defines the parameters for this module
//func DefaultParams() Params {
//	return NewParamsUnlocked()
//}

func (p Params) Validate() error {
	if err := validateSigners(p.Signers); err != nil { return err }
	return nil
}

func validateSigners(value interface{}) error {
	accz, ok := value.([]sdk.AccAddress)
	if !ok {
		return fmt.Errorf("unexpected Signers type: %T", value)
	}
	if len(accz) == 0 {
		return fmt.Errorf("signers list is empty")
	}
	for i, acc := range accz {
		if acc.Empty() {
			return fmt.Errorf("signer address is empty (#%d)", i)
		}
	}
	return nil
}

func (p StateParams) Validate() error {
	if err := validateLocked(p.Locked); err != nil { return err }
	if err := validatePointCost(p.VpnPointCost); err != nil { return err }
	if p.Locked && p.VpnPointCost.IsNullValue() {
		return errors.New("missing VpnPointCost")
	}
	if err := validatePointCost(p.StoragePointCost); err != nil { return err }
	if p.Locked && p.StoragePointCost.IsNullValue() {
		return errors.New("missing StoragePointCost")
	}
	if err := validateItemsPerBlock(p.ItemsPerBlock); err != nil { return err }
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
	if !ok { return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unexpected point cost type: %T", value)) }
	if !q.IsNullValue() && q.IsNegative() { return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("point cost must be non-negative: %v", q)) }
	return nil
}

func validateItemsPerBlock(value interface{}) error {
	_, ok := value.(uint16)
	if !ok { return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unexpected item count type: %T", value)) }
	// Value can be empty if Locked is false
	//if n < 1 { return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("items per block number must be positive: %d", n)) }
	return nil
}
