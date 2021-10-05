package bank

// nolint

import (
	"github.com/arterynetwork/artr/x/bank/internal/keeper"
	"github.com/arterynetwork/artr/x/bank/types"
)

const (
	QueryBalance       = keeper.QueryBalance
	QueryParams        = keeper.QueryParams
	ModuleName         = types.ModuleName
	QuerierRoute       = types.QuerierRoute
	RouterKey          = types.RouterKey
	StoreKey           = types.StoreKey
	DefaultParamspace  = types.DefaultParamspace
)

var (
	RegisterInvariants          = keeper.RegisterInvariants
	NonnegativeBalanceInvariant = keeper.NonnegativeBalanceInvariant
	NewBaseKeeper               = keeper.NewBaseKeeper
	NewBaseSendKeeper           = keeper.NewBaseSendKeeper
	NewBaseViewKeeper           = keeper.NewBaseViewKeeper
	NewQuerier                  = keeper.NewQuerier
	RegisterLegacyAminoCodec    = types.RegisterLegacyAminoCodec
	ErrNoInputs                 = types.ErrNoInputs
	ErrNoOutputs                = types.ErrNoOutputs
	ErrInputOutputMismatch      = types.ErrInputOutputMismatch
	ErrSendDisabled             = types.ErrSendDisabled
	NewGenesisState             = types.NewGenesisState
	DefaultGenesisState         = types.DefaultGenesisState
	ValidateGenesis             = types.ValidateGenesis
	NewMsgSend                  = types.NewMsgSend
	ParamKeyTable               = types.ParamKeyTable
	NewQueryBalanceParams       = types.NewQueryBalanceParams
	ModuleCdc                   = types.ModuleCdc
	NewInput                    = types.NewInput
	NewOutput                   = types.NewOutput
)

type (
	Keeper             = keeper.Keeper
	BaseKeeper         = keeper.BaseKeeper
	SendKeeper         = keeper.SendKeeper
	BaseSendKeeper     = keeper.BaseSendKeeper
	ViewKeeper         = keeper.ViewKeeper
	BaseViewKeeper     = keeper.BaseViewKeeper
	GenesisState       = types.GenesisState
	MsgSend            = types.MsgSend
	QueryBalanceParams = types.QueryBalanceParams
	Input              = types.Input
	Output             = types.Output
	Supply			   = types.Supply
	Params             = types.Params
)
