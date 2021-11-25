package delegating

import (
	"github.com/arterynetwork/artr/x/delegating/keeper"
	"github.com/arterynetwork/artr/x/delegating/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	MainStoreKey      = types.MainStoreKey
	DefaultParamspace = types.DefaultParamspace
	QuerierRoute      = types.QuerierRoute
	RevokeHookName    = types.RevokeHookName
	AccrueHookName    = types.AccrueHookName
)

var (
	// functions aliases
	NewKeeper                = keeper.NewKeeper
	NewQuerier               = keeper.NewQuerier
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec
	NewGenesisState          = types.NewGenesisState
	DefaultGenesisState      = types.DefaultGenesisState
	ValidateGenesis          = types.ValidateGenesis
	NewPercentage            = types.NewPercentage

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
	Params       = types.Params
	Percentage   = types.Percentage
	MsgDelegate  = types.MsgDelegate
	MsgRevoke    = types.MsgRevoke
)
