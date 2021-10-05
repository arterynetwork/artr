package app

import (
	"encoding/json"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *ArteryApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailWhiteList []string,
) (serverTypes.ExportedApp, error) {

	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})

	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	height := app.LastBlockHeight() + 1

	if forZeroHeight {
		height = 0
		app.prepForZeroHeightGenesis(ctx, jailWhiteList)
	}

	genState := app.mm.ExportGenesis(ctx, app.ec.Marshaler)
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return serverTypes.ExportedApp{}, err
	}

	// We should never have genesis validators per se.
	// All validators should be added via noding module instead.
	return serverTypes.ExportedApp{
		AppState:        appState,
		Validators:      nil,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, nil
}

// prepare for fresh start at zero height
// NOTE zero height genesis is a temporary feature which will be deprecated
//      in favour of export at a block height
func (app *ArteryApp) prepForZeroHeightGenesis(ctx sdk.Context, jailWhiteList []string) {
	panic("export to zero height genesis is not supported")
	// Almost every module schedules something and has block height somewhere in its data.
	// All these heights must be carefully patched if we want this feature implemented.
}
