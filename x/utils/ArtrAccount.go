package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/tendermint/tendermint/crypto"
)

//-----------------------------------------------------------------------------
// ArtrAccount

var _ exported.Account = (*ArtrAccount)(nil)
var _ exported.GenesisAccount = (*ArtrAccount)(nil)

// ArtrAccount - a base account structure.
// This can be extended by embedding within in your AppAccount.
// However one doesn't have to use ArtrAccount as long as your struct
// implements Account.
type ArtrAccount struct {
	auth.BaseAccount
	ActiveUntil    uint64 `json:"active_until" yaml:"active_until"`
	Noding         bool   `json:"noding" yaml:"noding"`
}

// NewArtrAccount creates a new ArtrAccount object
func NewArtrAccount(address sdk.AccAddress, coins sdk.Coins,
	pubKey crypto.PubKey, accountNumber uint64, sequence uint64) *ArtrAccount {

	return &ArtrAccount{
		BaseAccount: auth.BaseAccount{
			Address:       address,
			Coins:         coins,
			PubKey:        pubKey,
			AccountNumber: accountNumber,
			Sequence:      sequence,
		},
		ActiveUntil:    0,
		Noding:         false,
	}
}

// ProtoArtrAccount - a prototype function for ArtrAccount
func ProtoArtrAccount() exported.Account {
	return &ArtrAccount{}
}

// NewArtrAccountWithAddress - returns a new base account with a given address
func NewArtrAccountWithAddress(addr sdk.AccAddress) ArtrAccount {
	return ArtrAccount{
		BaseAccount: auth.BaseAccount{Address: addr},
	}
}
