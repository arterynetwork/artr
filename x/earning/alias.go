package earning

import (
	"github.com/arterynetwork/artr/x/earning/keeper"
	"github.com/arterynetwork/artr/x/earning/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
	QuerierRoute      = types.QuerierRoute
	StartHookName     = types.StartHookName
	ContinueHookName  = types.ContinueHookName
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	RegisterCodec       = types.RegisterCodec
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	NewEarner           = types.NewEarner
	NewPoints           = types.NewPoints

	// variable aliases
	ModuleCdc = types.ModuleCdc

	ErrAlreadyListed = types.ErrAlreadyListed
	ErrTooLate       = types.ErrTooLate
	ErrLocked        = types.ErrLocked
	ErrNotLocked     = types.ErrNotLocked
	ErrNoMoney       = types.ErrNoMoney
)

type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
	Params       = types.Params
	StateParams  = types.StateParams
	Earner       = types.Earner
	Points       = types.Points
)
