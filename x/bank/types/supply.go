package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewSupply creates a new Supply instance
func NewSupply(total sdk.Coins) *Supply {
	return &Supply{Total: total}
}

// DefaultSupply creates an empty Supply
func DefaultSupply() *Supply {
	return NewSupply(sdk.NewCoins())
}

// SetTotal sets the total supply.
func (supply *Supply) SetTotal(total sdk.Coins) {
	supply.Total = total
}

// GetTotal returns the supply total.
func (supply Supply) GetTotal() sdk.Coins {
	return supply.Total
}

// Inflate adds coins to the total supply
func (supply *Supply) Inflate(amount sdk.Coins) {
	supply.Total = sdk.Coins(supply.Total).Add(amount...)
}

// Deflate subtracts coins from the total supply.
func (supply *Supply) Deflate(amount sdk.Coins) {
	supply.Total = sdk.Coins(supply.Total).Sub(amount)
}

// String returns a human readable string representation of a supplier.
func (supply Supply) String() string {
	bz, _ := yaml.Marshal(supply)
	return string(bz)
}

// ValidateBasic validates the Supply coins and returns error if invalid
func (supply Supply) ValidateBasic() error {
	if !sdk.Coins(supply.Total).IsValid() {
		return fmt.Errorf("invalid total supply: %s", sdk.Coins(supply.Total).String())
	}

	return nil
}
