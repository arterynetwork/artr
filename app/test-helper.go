// +build testing

package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/kv"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/cosmos/cosmos-sdk/x/upgrade"

	"github.com/arterynetwork/artr/x/bank"
	"github.com/arterynetwork/artr/x/delegating"
	"github.com/arterynetwork/artr/x/earning"
	"github.com/arterynetwork/artr/x/noding"
	"github.com/arterynetwork/artr/x/profile"
	"github.com/arterynetwork/artr/x/referral"
	"github.com/arterynetwork/artr/x/schedule"
	"github.com/arterynetwork/artr/x/storage"
	"github.com/arterynetwork/artr/x/subscription"
	"github.com/arterynetwork/artr/x/voting"
	"github.com/arterynetwork/artr/x/vpn"
)

func init() {
	InitConfig()
	initDefaultGenesisUsers()
}

const verbose      = false
const printGenesis = false

func (app ArteryApp) GetKeys() map[string]*sdk.KVStoreKey { return app.keys }
func (app ArteryApp) GetTransientKeys() map[string]*sdk.TransientStoreKey { return app.tKeys }
func (app ArteryApp) GetSubspaces() map[string]params.Subspace { return app.subspaces }

func (app ArteryApp) GetAccountKeeper() auth.AccountKeeper { return app.accountKeeper }
func (app ArteryApp) GetBankKeeper() bank.Keeper { return app.bankKeeper }
func (app ArteryApp) GetSupplyKeeper() supply.Keeper { return app.supplyKeeper }
func (app ArteryApp) GetParamsKeeper() params.Keeper { return app.paramsKeeper }
func (app ArteryApp) GetUpgradeKeeper() upgrade.Keeper { return app.upgradeKeeper }
func (app ArteryApp) GetReferralKeeper() referral.Keeper { return app.referralKeeper }
func (app ArteryApp) GetProfileKeeper() profile.Keeper { return app.profileKeeper }
func (app ArteryApp) GetScheduleKeeper() schedule.Keeper { return app.scheduleKeeper }
func (app ArteryApp) GetDelegatingKeeper() delegating.Keeper { return app.delegatingKeeper }
func (app ArteryApp) GetVpnKeeper() vpn.Keeper { return app.vpnKeeper }
func (app ArteryApp) GetStorageKeeper() storage.Keeper { return app.storageKeeper }
func (app ArteryApp) GetSubscriptionKeeper() subscription.Keeper { return app.subscriptionKeeper }
func (app ArteryApp) GetVotingKeeper() voting.Keeper { return app.votingKeeper }
func (app ArteryApp) GetNodingKeeper() noding.Keeper { return app.nodingKeeper }
func (app ArteryApp) GetEarningKeeper() earning.Keeper { return app.earningKeeper }

func NewAppFromGenesis(genesis []byte) (app *ArteryApp, cleanup func()) {
	var logger log.Logger
	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}
	dir, _  := ioutil.TempDir("", "goleveldb-app-sim")
	db, _   := sdk.NewLevelDB("Simulation", dir)

	cleanup = func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}

	app = NewArteryApp(logger, db, nil, true, 0, fauxMerkleModeOpt)

	if genesis == nil {
		genesis = []byte(defaultGenesis)
	}

	var (
		genesisDoc   tmtypes.GenesisDoc
		genesisState simapp.GenesisState
	)
	app.Codec().MustUnmarshalJSON(genesis, &genesisDoc)
	app.Codec().MustUnmarshalJSON(genesisDoc.AppState, &genesisState)

	ctx := app.NewContext(true, abci.Header{})

	app.mm.InitGenesis(ctx, genesisState)

	return app, cleanup
}

func NewTestConsPubAddress() (crypto.PrivKey, crypto.PubKey, sdk.ConsAddress) {
	privKey := ed25519.GenPrivKey()
	pubKey  := privKey.PubKey()
	addr    := pubKey.Address()

	return privKey, pubKey, sdk.ConsAddress(addr.Bytes())
}

type Decoder func(bz []byte) (string, error)

func AccAddressDecoder(bz []byte) (string, error) {
	err := sdk.VerifyAddressFormat(bz)
	if err != nil {
		return "", err
	}
	return sdk.AccAddress(bz).String(), nil
}

func Uint64Decoder(bz []byte) (string, error) {
	if len(bz) != 8 {
		return "", fmt.Errorf("wrong uint64 length")
	}
	return fmt.Sprintf("%d", binary.BigEndian.Uint64(bz)), nil
}

func DummyDecoder(_ []byte) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (app ArteryApp) CheckExportImport(t *testing.T, storeKeys []string, keyDecoders, valueDecoders map[string]Decoder, ignorePrefixes map[string][][]byte) {
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight()})
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	appState, _, err := app.ExportAppStateAndValidators(false, nil)
	assert.NoError(t, err)

	if printGenesis {
		fmt.Println(string(appState))
	}

	var logger log.Logger
	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}
	dir, _  := ioutil.TempDir("", "goleveldb-app-sim-2")
	db, _   := sdk.NewLevelDB("Simulation-2", dir)

	defer func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}()

	app2 := NewArteryApp(logger, db, nil, true, 0, fauxMerkleModeOpt)
	var genesisState simapp.GenesisState
	app2.Codec().MustUnmarshalJSON(appState, &genesisState)
	ctx2 := app2.NewContext(true, abci.Header{Height: app2.LastBlockHeight()})
	app2.mm.InitGenesis(ctx2, genesisState)

	for _, key := range storeKeys {
		store1 := ctx.KVStore(app.GetKeys()[key])
		store2 := ctx2.KVStore(app2.GetKeys()[key])
		kvA, kvB := diffKVStores(store1, store2, ignorePrefixes[key])
		dkvA := decodeKVPairs(kvA, keyDecoders[key], valueDecoders[key])
		dkvB := decodeKVPairs(kvB, keyDecoders[key], valueDecoders[key])
		assert.Empty(t, dkvA, "bad pair(s) in original %s kvstore", key)
		assert.Empty(t, dkvB, "bad pair(s) in imported %s kvstore", key)
	}
}

//---------------------------------------------------------------------------------------------------

// Pass this in as an option to use a dbStoreAdapter instead of an IAVLStore for simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

func decodeKVPairs(kvz []kv.Pair, keyDecoder func([]byte)(string, error), valDecoder func([]byte)(string, error)) []string {
	result := make([]string, len(kvz))
	for i, kv := range kvz {
		var keyStr, valStr string

		s, err := keyDecoder(kv.Key)
		if err == nil {
			keyStr = s
		} else {
			keyStr = fmt.Sprintf("%v", kv.Key)
		}

		s, err = valDecoder(kv.Value)
		if err == nil {
			valStr = s
		} else {
			valStr = fmt.Sprintf("%v", kv.Value)
		}

		result[i] = fmt.Sprintf("{%s -> %s}", keyStr, valStr)
	}
	return result
}

func diffKVStores(a, b sdk.KVStore, ignore [][]byte) (kvAs, kvBs []kv.Pair) {
	iterA := a.Iterator(nil, nil)
	iterB := b.Iterator(nil, nil)

	for {
		shouldIgnore := func(key []byte) bool {
			for _, prefix := range ignore {
				if len(key) >= len(prefix) && bytes.Equal(prefix, key[:len(prefix)]) {
					return true
				}
			}
			return false
		}
		for iterA.Valid() && shouldIgnore(iterA.Key()) { iterA.Next() }
		for iterB.Valid() && shouldIgnore(iterB.Key()) { iterB.Next() }

		if !iterA.Valid() && !iterB.Valid() { break }

		var kvA, kvB kv.Pair
		if !iterA.Valid() {
			for ; iterB.Valid(); iterB.Next() {
				kvBs = append(kvBs, kv.Pair{Key: iterB.Key(), Value: iterB.Value()})
			}
			break
		}
		if !iterB.Valid() {
			for ; iterA.Valid(); iterA.Next() {
				kvAs = append(kvAs, kv.Pair{Key: iterA.Key(), Value: iterA.Value()})
			}
			break
		}

		kvA = kv.Pair{Key: iterA.Key(), Value: iterA.Value()}
		iterA.Next()

		kvB = kv.Pair{Key: iterB.Key(), Value: iterB.Value()}
		iterB.Next()

		if !bytes.Equal(kvA.Key, kvB.Key) || !bytes.Equal(kvA.Value, kvB.Value) {
			kvAs = append(kvAs, kvA)
			kvBs = append(kvBs, kvB)
		}
	}
	return kvAs, kvBs
}
