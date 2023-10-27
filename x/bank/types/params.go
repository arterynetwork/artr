package types

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramTypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/arterynetwork/artr/util"
)

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName

	DefaultMinSend        = 1000
	DefaultDustDelegation = 0

	DefaultMaxTransactionFee = 10_000000
)

var (
	DefaultTransactionFee            = util.Permille(3)
	DefaultTransactionFeeSplitRatios = TransactionFeeSplitRatios{
		ForProposer: util.FractionInt(1),
		ForCompany:  util.FractionZero(),
	}
	DefaultBlockedSenders []sdk.AccAddress = nil
)

var (
	ParamStoreKeyMinSend                   = []byte("minsend")
	ParamStoreKeyDustDelegation            = []byte("dustd")
	ParamStoreKeyTransactionFee            = []byte("txfee")
	ParamStoreKeyMaxTransactionFee         = []byte("maxtxfee")
	ParamStoreKeyTransactionFeeSplitRatios = []byte("txfeesplitratios")
	ParamStoreKeyCompanyAccount            = []byte("companyaccount")
	ParamStoreKeyBlockedSenders            = []byte("blockedsenders")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramTypes.KeyTable {
	return paramTypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p TransactionFeeSplitRatios) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p Params) GetBlockedSenders() []sdk.AccAddress {
	res := make([]sdk.AccAddress, len(p.BlockedSenders))
	for i, bech32 := range p.BlockedSenders {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			panic(err)
		}
		res[i] = addr
	}
	return res
}

func (p *Params) SetBlockedSenders(blockedSenders []sdk.AccAddress) {
	p.BlockedSenders = make([]string, len(blockedSenders))
	for i, addr := range blockedSenders {
		p.BlockedSenders[i] = addr.String()
	}
}

func (p *Params) ParamSetPairs() paramTypes.ParamSetPairs {
	return paramTypes.ParamSetPairs{
		paramTypes.NewParamSetPair(ParamStoreKeyMinSend, &p.MinSend, validateMinSend),
		paramTypes.NewParamSetPair(ParamStoreKeyDustDelegation, &p.DustDelegation, validateDustDelegation),
		paramTypes.NewParamSetPair(ParamStoreKeyTransactionFee, &p.TransactionFee, validateTransactionFee),
		paramTypes.NewParamSetPair(ParamStoreKeyMaxTransactionFee, &p.MaxTransactionFee, validateMaxTransactionFee),
		paramTypes.NewParamSetPair(ParamStoreKeyTransactionFeeSplitRatios, &p.TransactionFeeSplitRatios, validateTransactionFeeSplitRatios),
		paramTypes.NewParamSetPair(ParamStoreKeyCompanyAccount, &p.CompanyAccount, validateCompanyAccount),
		paramTypes.NewParamSetPair(ParamStoreKeyBlockedSenders, &p.BlockedSenders, validateAccounts),
	}
}

// NewParams creates a new parameter configuration for the bank module
func NewParams(minSend int64, dust int64, fee util.Fraction, maxFee int64, feeSplitRatios TransactionFeeSplitRatios, companyAccount string, blockedSenders []sdk.AccAddress) Params {
	p := Params{
		MinSend:                   minSend,
		DustDelegation:            dust,
		TransactionFee:            fee,
		MaxTransactionFee:         maxFee,
		TransactionFeeSplitRatios: feeSplitRatios,
		CompanyAccount:            companyAccount,
	}
	p.SetBlockedSenders(blockedSenders)

	return p
}

// DefaultParams defines the parameters for the bank module
func DefaultParams() Params {
	p := Params{
		MinSend:                   DefaultMinSend,
		DustDelegation:            DefaultDustDelegation,
		TransactionFee:            DefaultTransactionFee,
		MaxTransactionFee:         DefaultMaxTransactionFee,
		TransactionFeeSplitRatios: DefaultTransactionFeeSplitRatios,
	}
	p.SetBlockedSenders(DefaultBlockedSenders)

	return p
}

// Validate all bank module parameters
func (p Params) Validate() error {
	if err := validateMinSend(p.MinSend); err != nil {
		return errors.Wrap(err, "invalid min_send")
	}
	if err := validateDustDelegation(p.DustDelegation); err != nil {
		return errors.Wrap(err, "invalid dust_delegation")
	}
	if err := validateTransactionFee(p.TransactionFee); err != nil {
		return errors.Wrap(err, "invalid transaction_fee")
	}
	if err := validateMaxTransactionFee(p.MaxTransactionFee); err != nil {
		return errors.Wrap(err, "invalid max_transaction_fee")
	}
	if err := validateTransactionFeeSplitRatios(p.TransactionFeeSplitRatios); err != nil {
		return errors.Wrap(err, "invalid transaction_fee_slit_ratios")
	}
	if err := validateCompanyAccount(p.CompanyAccount); err != nil {
		return errors.Wrap(err, "invalid company_account")
	}
	if err := validateAccounts(p.BlockedSenders); err != nil {
		return errors.Wrap(err, "invalid blocked_senders")
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMinSend(i interface{}) error {
	_, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDustDelegation(i interface{}) error {
	dt, ok := i.(int64)
	if !ok {
		return errors.Errorf("invalid DustDelegation parameter type: %T", i)
	}
	if dt < 0 {
		return errors.New("DustDelegation must be non-negative")
	}
	return nil
}

func validateTransactionFee(i interface{}) error {
	dt, ok := i.(util.Fraction)
	if !ok {
		return errors.Errorf("invalid TransactionFee parameter type: %T", i)
	}
	if dt.IsNegative() {
		return errors.New("TransactionFee must be non-negative")
	}
	return nil
}

func validateMaxTransactionFee(i interface{}) error {
	dt, ok := i.(int64)
	if !ok {
		return errors.Errorf("invalid MaxTransactionFee parameter type: %T", i)
	}
	if dt < 0 {
		return errors.New("MaxTransactionFee must be non-negative")
	}
	return nil
}

func validateTransactionFeeSplitRatios(i interface{}) error {
	dt, ok := i.(TransactionFeeSplitRatios)
	if !ok {
		return errors.Errorf("invalid TransactionFeeSplitRatios parameter type: %T", i)
	}
	if dt.ForProposer.IsNegative() {
		return errors.New("TransactionFeeSplitRatios.ForProposer must be non-negative")
	}
	if dt.ForCompany.IsNegative() {
		return errors.New("TransactionFeeSplitRatios.ForCompany must be non-negative")
	}
	if dt.ForProposer.Add(dt.ForCompany).GT(util.FractionInt(1)) {
		return errors.New("TransactionFeeSplitRatios sums must be less than or equal 1")
	}
	if util.CalculateTransactionFeeSplitRatiosLCM(dt.ForProposer, dt.ForCompany).GT(sdk.NewInt(util.TransactionFeeSplitRatiosMaxLcm)) {
		return errors.Errorf("TransactionFeeSplitRatios LCM must be less than or equal %d", util.TransactionFeeSplitRatiosMaxLcm)
	}
	return nil
}

func validateCompanyAccount(i interface{}) error {
	dt, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if _, err := sdk.AccAddressFromBech32(dt); err != nil {
		return errors.Wrap(err, "cannot parse account address")
	}

	return nil
}

func validateAccounts(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for i, bech32 := range v {
		_, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			return errors.Wrapf(err, "invalid acc address #%d", i)
		}
	}

	return nil
}
