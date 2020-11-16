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

	StatusLucky            = types.Lucky
	StatusLeader           = types.Leader
	StatusMaster           = types.Master
	StatusChampion         = types.Champion
	StatusBusinessman      = types.Businessman
	StatusProfessional     = types.Professional
	StatusTopLeader        = types.TopLeader
	StatusHero             = types.Hero
	StatusAbsoluteChampion = types.AbsoluteChampion

	CompressionPeriod    = keeper.CompressionPeriod
	StatusDowngradeAfter = keeper.StatusDowngradeAfter

	StatusUpdatedCallback = keeper.StatusUpdatedCallback
	StakeChangedCallback  = keeper.StakeChangedCallback

	StatusDowngradeHookName = keeper.StatusDowngradeHookName
	CompressionHookName     = keeper.CompressionHookName
)

var (
	// functions aliases
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	RegisterCodec       = types.RegisterCodec
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis

	// variable aliases
	ModuleCdc     = types.ModuleCdc
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
	DataRecord        = types.R
)
