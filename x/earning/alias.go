package earning

import (
	"github.com/arterynetwork/artr/x/earning/keeper"
	"github.com/arterynetwork/artr/x/earning/types"
)

const (
	ModuleName           = types.ModuleName
	RouterKey            = types.RouterKey
	StoreKey             = types.StoreKey
	DefaultParamspace    = types.DefaultParamspace
	QuerierRoute         = types.QuerierRoute
	VpnCollectorName     = types.VpnCollectorName
	StorageCollectorName = types.StorageCollectorName
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	NewEarner           = types.NewEarner
	NewTimestamps       = types.NewTimestamps

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
	Params       = types.Params
	Earner       = types.Earner
	Timestamps   = types.Timestamps
)
