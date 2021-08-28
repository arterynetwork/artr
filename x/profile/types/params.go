package types

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

// Default parameter namespace
const (
	DefaultParamspace = ModuleName

	DefaultFee               int64  = 1000000
	DefaultCardMagic         uint64 = 0x1A21A4B61B
	DefaultSubscriptionPrice uint32 = 1990
	DefaultVpnGbPrice        uint32 = 10
	DefaultStorageGbPrice    uint32 = 10
	DefaultBaseVpnGb         uint32 = 7
	DefaultBaseStorageGb     uint32 = 5
)

// Parameter store keys
var (
	DefaultCreators         []sdk.AccAddress = nil
	DefaultStorageSigners   []sdk.AccAddress = nil
	DefaultVpnSigners       []sdk.AccAddress = nil
	DefaultTokenRateSigners []sdk.AccAddress = nil
	DefaultTokenRate                         = util.FractionInt(100000)

	KeyCreators          = []byte("Creators")
	KeyFee               = []byte("RenamePrice")
	KeyCardMagic         = []byte("CardMagic")
	KeyStorageSigners    = []byte("StorageSigners")
	KeyVpnSigners        = []byte("VpnSigners")
	KeyTokenRate         = []byte("TokenRate")
	KeySubscriptionPrice = []byte("SubscriptionPrice")
	KeyVpnGbPrice        = []byte("VpnGbPrice")
	KeyStorageGbPrice    = []byte("StorageGbPrice")
	KeyBaseVpnGb         = []byte("BaseVpnGb")
	KeyBaseStorageGb     = []byte("BaseStorageGb")
	KeyTokenRateSigners  = []byte("TokenRateSigners")
)

// ParamKeyTable for profile module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p Params) GetCreators() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.Creators))
	for i, bech32 := range p.Creators {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p *Params) SetCreators(creators []sdk.AccAddress) {
	p.Creators = make([]string, len(creators))
	for i, addr := range creators {
		p.Creators[i] = addr.String()
	}
}

func (p Params) GetStorageSigners() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.StorageSigners))
	for i, bech32 := range p.StorageSigners {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p *Params) SetStorageSigners(signers []sdk.AccAddress) {
	p.StorageSigners = make([]string, len(signers))
	for i, addr := range signers {
		p.StorageSigners[i] = addr.String()
	}
}

func (p Params) GetVpnSigners() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.VpnSigners))
	for i, bech32 := range p.StorageSigners {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p *Params) SetVpnSigners(signers []sdk.AccAddress) {
	p.VpnSigners = make([]string, len(signers))
	for i, addr := range signers {
		p.VpnSigners[i] = addr.String()
	}
}

func (p Params) GetTokenRateSigners() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.TokenRateSigners))
	for i, bech32 := range p.TokenRateSigners {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p *Params) SetTokenRateSigners(signers []sdk.AccAddress) {
	p.TokenRateSigners = make([]string, len(signers))
	for i, addr := range signers {
		p.TokenRateSigners[i] = addr.String()
	}
}

// NewParams creates a new Params object
func NewParams(
	creators []sdk.AccAddress,
	renamePrice int64,
	cardMagic uint64,
	storageSigners, vpnSigners []sdk.AccAddress,
	tokenRate util.Fraction,
	subscriptionPrice uint32,
	vpnGbPrice, storageGbPrice uint32,
	baseVpnGb, baseStorageGb uint32,
	tokenRateSigners []sdk.AccAddress) *Params {
	p := &Params{
		RenamePrice:       renamePrice,
		CardMagic:         cardMagic,
		TokenRate:         tokenRate,
		SubscriptionPrice: subscriptionPrice,
		VpnGbPrice:        vpnGbPrice,
		StorageGbPrice:    storageGbPrice,
		BaseVpnGb:         baseVpnGb,
		BaseStorageGb:     baseStorageGb,
	}
	p.SetCreators(creators)
	p.SetStorageSigners(storageSigners)
	p.SetVpnSigners(vpnSigners)
	p.SetTokenRateSigners(tokenRateSigners)
	return p
}

// String implements the stringer interface for Params
func (p Params) String() string {
	out, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyCreators, &p.Creators, validateAccounts),
		paramtypes.NewParamSetPair(KeyFee, &p.RenamePrice, validateFee),
		paramtypes.NewParamSetPair(KeyCardMagic, &p.CardMagic, validateCardMagic),
		paramtypes.NewParamSetPair(KeyStorageSigners, &p.StorageSigners, validateAccounts),
		paramtypes.NewParamSetPair(KeyVpnSigners, &p.VpnSigners, validateAccounts),
		paramtypes.NewParamSetPair(KeyTokenRate, &p.TokenRate, validateTokenRate),
		paramtypes.NewParamSetPair(KeySubscriptionPrice, &p.SubscriptionPrice, validatePositive),
		paramtypes.NewParamSetPair(KeyVpnGbPrice, &p.VpnGbPrice, validatePositive),
		paramtypes.NewParamSetPair(KeyStorageGbPrice, &p.StorageGbPrice, validatePositive),
		paramtypes.NewParamSetPair(KeyBaseVpnGb, &p.BaseVpnGb, validatePositive),
		paramtypes.NewParamSetPair(KeyBaseStorageGb, &p.BaseStorageGb, validatePositive),
		paramtypes.NewParamSetPair(KeyTokenRateSigners, &p.TokenRateSigners, validateAccounts),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() *Params {
	return NewParams(
		DefaultCreators,
		DefaultFee,
		DefaultCardMagic,
		DefaultStorageSigners,
		DefaultVpnSigners,
		DefaultTokenRate,
		DefaultSubscriptionPrice,
		DefaultVpnGbPrice,
		DefaultStorageGbPrice,
		DefaultBaseVpnGb,
		DefaultBaseStorageGb,
		DefaultTokenRateSigners,
	)
}

func (p Params) Validate() error {
	if err := validateAccounts(p.Creators); err != nil {
		return errors.Wrap(err, "invalid creators")
	}
	if err := validateFee(p.RenamePrice); err != nil {
		return errors.Wrap(err, "invalid rename_price")
	}
	if err := validateCardMagic(p.CardMagic); err != nil {
		return errors.Wrap(err, "invalid card_magic")
	}
	if err := validateAccounts(p.StorageSigners); err != nil {
		return errors.Wrap(err, "invalid storage_signers")
	}
	if err := validateAccounts(p.VpnSigners); err != nil {
		return errors.Wrap(err, "invalid vpn_signers")
	}
	if err := validateTokenRate(p.TokenRate); err != nil {
		return errors.Wrap(err, "invalid token_rate")
	}
	if err := validatePositive(p.SubscriptionPrice); err != nil {
		return errors.Wrap(err, "invalid subscription_price")
	}
	if err := validatePositive(p.VpnGbPrice); err != nil {
		return errors.Wrap(err, "invalid vpn_gb_price")
	}
	if err := validatePositive(p.StorageGbPrice); err != nil {
		return errors.Wrap(err, "invalid storage_gb_price")
	}
	if err := validatePositive(p.BaseVpnGb); err != nil {
		return errors.Wrap(err, "invalid base_vpn_gb")
	}
	if err := validatePositive(p.BaseStorageGb); err != nil {
		return errors.Wrap(err, "invalid base_storage_gb")
	}
	if err := validateAccounts(p.TokenRateSigners); err != nil {
		return errors.Wrap(err, "invaid token_rate_signers")
	}

	return nil
}

func validateAccounts(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return fmt.Errorf("account list empty")
	}

	for i, bech32 := range v {
		_, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			return errors.Wrapf(err, "invalid acc address #%d", i)
		}
	}

	return nil
}

func validateFee(i interface{}) error {
	_, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateCardMagic(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid CardMagic parameter type: %T", i)
	}

	return nil
}

func validateTokenRate(i interface{}) error {
	val, ok := i.(util.Fraction)
	if !ok {
		return errors.Errorf("invalid TokenRate parameter type: %T", i)
	}
	if !val.IsPositive() {
		return errors.New("TokenRate must be positive")
	}
	return nil
}

func validatePositive(i interface{}) error {
	val, ok := i.(uint32)
	if !ok {
		return errors.Errorf("invalid parameter type: %i (uint32 expected)", i)
	}
	if val <= 0 {
		return errors.New("parameter value must be positive")
	}
	return nil
}
