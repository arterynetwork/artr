package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// module name
	ModuleName   = "artrbank"
	QuerierRoute = ModuleName
	StoreKey     = ModuleName
)

// KVStore keys
var (
	BalancesPrefix      = []byte("balances")
	SupplyKey           = []byte{0x00}
	DenomMetadataPrefix = []byte{0x1}
)

// DenomMetadataKey returns the denomination metadata key.
func DenomMetadataKey(denom string) []byte {
	d := []byte(denom)
	return append(DenomMetadataPrefix, d...)
}

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the perfix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
func AddressFromBalancesStore(key []byte) sdk.AccAddress {
	addr := key[:sdk.AddrLen]
	if len(addr) != sdk.AddrLen {
		panic(fmt.Sprintf("unexpected account address key length; got: %d, expected: %d", len(addr), sdk.AddrLen))
	}

	return sdk.AccAddress(addr)
}
