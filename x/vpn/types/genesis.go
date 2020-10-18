package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisVpnInfo struct {
	VpnInfo
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

// GenesisState - all vpn state that must be provided at genesis
type GenesisState struct {
	Params    Params           `json:"params" yaml:"params"`
	VpnStatus []GenesisVpnInfo `json:"vpn_statuses" yaml:"vpn_statuses"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, vpnStatus []GenesisVpnInfo) GenesisState {
	return GenesisState{
		Params:    params,
		VpnStatus: vpnStatus,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:    DefaultParams(),
		VpnStatus: nil,
	}
}

// ValidateGenesis validates the vpn genesis parameters
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil { return err }
	return nil
}
