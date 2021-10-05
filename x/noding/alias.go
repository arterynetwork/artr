package noding

import (
	"github.com/arterynetwork/artr/x/noding/keeper"
	"github.com/arterynetwork/artr/x/noding/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	IdxStoreKey       = types.IdxStoreKey
	DefaultParamspace = types.DefaultParamspace
	QuerierRoute      = types.QuerierRoute
	SwitchOnConst     = types.SwitchOnConst
	SwitchOffConst    = types.SwitchOffConst
	UnjailConst       = types.UnjailConst

	ValidatorStateOff   = types.VALIDATOR_STATE_OFF
	ValidatorStateBan   = types.VALIDATOR_STATE_BAN
	ValidatorStateJail  = types.VALIDATOR_STATE_JAIL
	ValidatorStateSpare = types.VALIDATOR_STATE_SPARE
	ValidatorStateLucky = types.VALIDATOR_STATE_LUCKY
	ValidatorStateTop   = types.VALIDATOR_STATE_TOP
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis

	// variable aliases
	ModuleCdc            = types.ModuleCdc
	ErrNotQualified      = types.ErrNotQualified
	ErrPubkeyBusy        = types.ErrPubkeyBusy
	ErrNotFound          = types.ErrNotFound
	ErrNotJailed         = types.ErrNotJailed
	ErrJailPeriodNotOver = types.ErrJailPeriodNotOver
	ErrBannedForLifetime = types.ErrBannedForLifetime
	ErrAlreadyOn         = types.ErrAlreadyOn
)

type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
	Params       = types.Params

	ValidatorState = types.ValidatorState
)
