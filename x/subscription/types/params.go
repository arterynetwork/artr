package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const DefaultParamspace = ModuleName

// Default parameters
const (
	// 1 RUB = 100000UARTR
	DefaultTokenCourse uint32 = 100000
	// All prices in RUB
	DefaultSubscriptionPrice uint32 = 1990
	DefaultVPNGbPrice        uint32 = 10
	DefaultStorageGbPrice    uint32 = 10
	DefaultBaseVPNGb         uint32 = 7
	DefaultBaseStorageGb     uint32 = 5
)

// Parameter store keys
var (
	// KeyParamName          = []byte("ParamName")
	KeyTokenCourse         = []byte("TokenCourse")
	KeySubscriptionPrice   = []byte("SubscriptionPrice")
	KeyVPNGbPrice          = []byte("VPNGbPrice")
	KeyStorageGbPrice      = []byte("StorageGbPrice")
	KeyBaseVPNGb           = []byte("BaseVPNGb")
	KeyBaseStorageGb       = []byte("BaseStorageGb")
	KeyCourseChangeSigners = []byte("CourseChangeSigners")
)

// ParamKeyTable for subscription module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

var _ subspace.ParamSet = &Params{}

// Params - used for initializing default parameter for subscription at genesis
type Params struct {
	// KeyParamName string `json:"key_param_name"`
	TokenCourse         uint32           `json:"token_course" yaml:"token_course"`
	SubscriptionPrice   uint32           `json:"subscription_price" yaml:"subscription_price"`
	VPNGBPrice          uint32           `json:"vpn_gb_price" yaml:"vpn_gb_price"`
	StorageGBPrice      uint32           `json:"storage_gb_price" yaml:"storage_gb_price"`
	BaseVPNGb           uint32           `json:"base_vpn_gb" yaml:"base_vpn_gb"`
	BaseStorageGb       uint32           `json:"base_storage_gb" yaml:"base_storage_gb"`
	CourseChangeSigners []sdk.AccAddress `json:"course_change_signers" yaml:"course_change_signers"`
}

// NewParams creates a new Params object
func NewParams(tokenCourse, subscriptionPrice, VPNGBPrice,
	storageGBPrice, baseVPNGb, baseStorageGb uint32, courseSigners []sdk.AccAddress) Params {
	return Params{
		TokenCourse:         tokenCourse,
		SubscriptionPrice:   subscriptionPrice,
		VPNGBPrice:          VPNGBPrice,
		StorageGBPrice:      storageGBPrice,
		BaseVPNGb:           baseVPNGb,
		BaseStorageGb:       baseStorageGb,
		CourseChangeSigners: courseSigners[:],
	}
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(
		"TokenCourse: %d\n"+
			"SubscriptionPrice: %d\n"+
			"VPNGBPrice: %d\n"+
			"StorageGBPrice: %d\n"+
			"BaseVPNGb: %d\n"+
			"BaseStorageGb: %d\n"+
			"CouseChangeSigners: %v\n",
		p.TokenCourse,
		p.SubscriptionPrice,
		p.VPNGBPrice,
		p.StorageGBPrice,
		p.BaseVPNGb,
		p.BaseStorageGb,
		p.CourseChangeSigners,
	)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyTokenCourse, &p.TokenCourse, validateTokenCourse),
		params.NewParamSetPair(KeySubscriptionPrice, &p.SubscriptionPrice, validateSubscriptionPrice),
		params.NewParamSetPair(KeyVPNGbPrice, &p.VPNGBPrice, validateVPNGBPrice),
		params.NewParamSetPair(KeyStorageGbPrice, &p.StorageGBPrice, validateStorageGBPrice),
		params.NewParamSetPair(KeyBaseVPNGb, &p.BaseVPNGb, validateBaseVPNGb),
		params.NewParamSetPair(KeyBaseStorageGb, &p.BaseStorageGb, validateBaseStorageGb),
		params.NewParamSetPair(KeyCourseChangeSigners, &p.CourseChangeSigners, validateCourseChangeSigners),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultTokenCourse,
		DefaultSubscriptionPrice,
		DefaultVPNGbPrice,
		DefaultStorageGbPrice,
		DefaultBaseVPNGb,
		DefaultBaseStorageGb,
		nil,
	)
}

func validateTokenCourse(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid token course: %d", v)
	}

	return nil
}

func validateSubscriptionPrice(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid subscription price: %d", v)
	}

	return nil
}

func validateVPNGBPrice(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid vpn price: %d", v)
	}

	return nil
}

func validateStorageGBPrice(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid storage price: %d", v)
	}

	return nil
}

func validateBaseVPNGb(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid base VPN Bb volume: %d", v)
	}

	return nil
}

func validateBaseStorageGb(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid base storage Gb volume: %d", v)
	}

	return nil
}

func validateCourseChangeSigners(i interface{}) error {
	v, ok := i.([]sdk.AccAddress)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return fmt.Errorf("empty exchange rate signer list")
	}

	for i, signer := range v {
		if signer.Empty() {
			return fmt.Errorf("empty exchange rate signer account address (#%d)", i)
		}
	}

	return nil
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateTokenCourse(p.TokenCourse); err != nil {
		return err
	}
	if err := validateSubscriptionPrice(p.SubscriptionPrice); err != nil {
		return err
	}
	if err := validateVPNGBPrice(p.VPNGBPrice); err != nil {
		return err
	}
	if err := validateStorageGBPrice(p.StorageGBPrice); err != nil {
		return err
	}
	if err := validateBaseVPNGb(p.BaseVPNGb); err != nil {
		return err
	}
	if err := validateBaseStorageGb(p.BaseStorageGb); err != nil {
		return err
	}
	if err := validateCourseChangeSigners(p.CourseChangeSigners); err != nil {
		return err
	}

	return nil
}
