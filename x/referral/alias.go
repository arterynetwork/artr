package referral

import (
	"github.com/arterynetwork/artr/x/referral/keeper"
	"github.com/arterynetwork/artr/x/referral/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	IndexStoreKey     = types.IndexStoreKey
	DefaultParamspace = types.DefaultParamspace
	QuerierRoute      = types.QuerierRoute

	StatusLucky            = types.STATUS_LUCKY
	StatusLeader           = types.STATUS_LEADER
	StatusMaster           = types.STATUS_MASTER
	StatusChampion         = types.STATUS_CHAMPION
	StatusBusinessman      = types.STATUS_BUSINESSMAN
	StatusProfessional     = types.STATUS_PROFESSIONAL
	StatusTopLeader        = types.STATUS_TOP_LEADER
	StatusAbsoluteChampion = types.STATUS_ABSOLUTE_CHAMPION

	StatusUpdatedCallback = keeper.StatusUpdatedCallback
	StakeChangedCallback  = keeper.StakeChangedCallback
	BanishedCallback      = keeper.BanishedCallback

	StatusDowngradeHookName   = keeper.StatusDowngradeHookName
	CompressionHookName       = keeper.CompressionHookName
	TransitionTimeoutHookName = keeper.TransitionTimeoutHookName
	BanishHookName            = keeper.BanishHookName
	StatusBonusHookName       = keeper.StatusBonusHookName
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper            = keeper.Keeper
	GenesisState      = types.GenesisState
	ReferralFee       = types.ReferralFee
	Params            = types.Params
	CompanyAccounts   = types.CompanyAccounts
	NetworkAward      = types.NetworkAward
	StatusCheckResult = types.StatusCheckResult
	Status            = types.Status
	DataRecord        = types.Info
)
