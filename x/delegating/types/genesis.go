package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/arterynetwork/artr/util"
)

// GenesisState - all delegating state that must be provided at genesis
type GenesisState struct {
	Params   Params    `json:"params"`
	Clusters []Cluster `json:"clusters"`
	Revoking []Revoke  `json:"revoking"`
}

type Cluster struct {
	Modulo   uint16           `json:"modulo"`
	Accounts []sdk.AccAddress `json:"accounts"`
}

type Revoke struct {
	Account sdk.AccAddress `json:"account"`
	Amount  int64          `json:"amount"`
	Height  int64          `json:"height"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, clusters []Cluster, revoking []Revoke) GenesisState {
	return GenesisState{
		Params:   params,
		Clusters: clusters,
		Revoking: revoking,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the delegating genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	for i, cluster := range data.Clusters {
		if cluster.Modulo >= util.BlocksOneDay {
			return fmt.Errorf("modulo must be less than %d", util.BlocksOneDay)
		}
		for j, account := range cluster.Accounts {
			if account.Empty() {
				return fmt.Errorf("account is empty (#%d.%d)", i, j)
			}
		}
	}
	return nil
}
