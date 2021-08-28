package types

import (
	"bytes"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, balances []Balance, supply sdk.Coins) *GenesisState {
	return &GenesisState{
		Params:   params,
		Balances: balances,
		Supply:   supply,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(Params{MinSend: 1000}, nil, nil)
}

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error { return nil }

// SanitizeGenesisAccounts sorts addresses and coin sets.
func SanitizeGenesisBalances(balances []Balance) []Balance {
	sort.Slice(balances, func(i, j int) bool {
		addr1, _ := sdk.AccAddressFromBech32(balances[i].Address)
		addr2, _ := sdk.AccAddressFromBech32(balances[j].Address)
		return bytes.Compare(addr1.Bytes(), addr2.Bytes()) < 0
	})

	for _, balance := range balances {
		balance.Coins = sdk.Coins(balance.Coins).Sort()
	}

	return balances
}
