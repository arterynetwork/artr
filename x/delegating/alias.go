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
	ValidatePercentageRanges = types.ValidatePercentageRanges
	ValidatePercentageTable  = types.ValidatePercentageTable

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper              = keeper.Keeper
	GenesisState        = types.GenesisState
	Params              = types.Params
	Percentage          = types.Percentage
	PercentageRange     = types.PercentageRange
	PercentageListRange = types.PercentageListRange
	MsgDelegate         = types.MsgDelegate
	MsgRevoke           = types.MsgRevoke
)
