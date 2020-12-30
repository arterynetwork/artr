package noding

import (
	"github.com/arterynetwork/artr/x/noding/keeper"
	"github.com/arterynetwork/artr/x/noding/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	IdxStoreKey       = types.IdxSoreKey
	DefaultParamspace = types.DefaultParamspace
	QuerierRoute      = types.QuerierRoute
	SwitchOnConst     = types.SwitchOnConst
	SwitchOffConst    = types.SwitchOffConst
	UnjailConst       = types.UnjailConst
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	NewMsgSwitchOn      = types.NewMsgSwitchOn
	NewMsgSwitchOff     = types.NewMsgSwitchOff
	NewMsgUnjail        = types.NewMsgUnjail
	RegisterCodec       = types.RegisterCodec
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

	MsgSwitchOn  = types.MsgSwitchOn
	MsgSwitchOff = types.MsgSwitchOff
	MsgUnjail    = types.MsgUnjail
)
