package keeper

import (
	"bytes"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/arterynetwork/artr/x/referral/types"
)

type kvRecord struct {
	key   []byte
	value []byte
}

type callback struct {
	event string
	acc   string
}

func (x callback) Eq(y callback) bool {
	return x.event == y.event && x.acc == y.acc
}

type callbacks []callback

func (cbz *callbacks) Len() int {
	return len(*cbz)
}
func (cbz *callbacks) Less(i, j int) bool {
	x := (*cbz)[i]
	y := (*cbz)[j]

	res := strings.Compare(x.acc, y.acc)
	if res < 0 {
		return true
	} else if res > 0 {
		return false
	} else {
		return x.event < y.event
	}
}
func (cbz *callbacks) Swap(i, j int) {
	tmp := (*cbz)[i]
	(*cbz)[i] = (*cbz)[j]
	(*cbz)[j] = tmp
}

type bunchUpdater struct {
	k         Keeper
	ctx       sdk.Context
	data      []kvRecord
	callbacks callbacks
}

func newBunchUpdater(k Keeper, ctx sdk.Context) *bunchUpdater {
	return &bunchUpdater{
		k:         k,
		ctx:       ctx,
		data:      nil,
		callbacks: nil,
	}
}

func (bu *bunchUpdater) set(acc string, value types.Info) error {
	keyBytes := []byte(acc)
	valueBytes, err := bu.k.cdc.MarshalBinaryBare(&value)
	if err != nil {
		return err
	}
	for i, record := range bu.data {
		if bytes.Equal(record.key, keyBytes) {
			bu.data[i].value = valueBytes
			return nil
		}
	}
	bu.data = append(bu.data, kvRecord{
		key:   keyBytes,
		value: valueBytes,
	})
	return nil
}

//TODO: Refactor to mitigate string <-> AccAddress casting
func (bu *bunchUpdater) get(acc string) (types.Info, error) {
	var (
		keyBytes   = []byte(acc)
		valueBytes = []byte(nil)

		value types.Info
	)
	for _, record := range bu.data {
		if bytes.Equal(record.key, keyBytes) {
			valueBytes = record.value
			break
		}
	}
	if valueBytes == nil {
		store := bu.ctx.KVStore(bu.k.storeKey)
		valueBytes = store.Get(keyBytes)
	}
	err := bu.k.cdc.UnmarshalBinaryBare(valueBytes, &value)
	return value, err
}

func (bu bunchUpdater) StatusDowngradeAfter() time.Duration {
	return bu.k.scheduleKeeper.OneMonth(bu.ctx)
}

func (bu *bunchUpdater) update(acc string, checkForStatusUpdate bool, callback func(value *types.Info) error) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(store.ErrorOutOfGas); ok {
				panic(e)
			} else if er, ok := e.(error); ok {
				err = errors.Wrap(er, "update paniced")
			} else {
				err = errors.Errorf("update paniced: %s", e)
			}
		}
	}()
	value, err := bu.get(acc)
	if err != nil {
		bu.k.Logger(bu.ctx).Info("Cannot update, no such account", "addr", acc)
		return nil
	}
	value.Normalize()
	err = callback(&value)
	if err != nil {
		return errors.Wrap(err, "callback failed")
	}
	if checkForStatusUpdate {
		checkResult, err := checkStatusRequirements(value.Status, value, bu)
		if err != nil {
			return err
		}
		if !checkResult.Overall {
			if value.StatusDowngradeAt == nil {
				downgradeAt := bu.ctx.BlockTime().Add(bu.StatusDowngradeAfter())
				value.StatusDowngradeAt = &downgradeAt
				bu.k.scheduleKeeper.ScheduleTask(bu.ctx, downgradeAt, StatusDowngradeHookName, []byte(acc))
				if err := bu.ctx.EventManager().EmitTypedEvent(
					&types.EventStatusWillBeDowngraded{
						Address: acc,
						Time:    downgradeAt,
					},
				); err != nil { panic(err) }
			}
		} else {
			if value.StatusDowngradeAt != nil {
				value.StatusDowngradeAt = nil
				if err := bu.ctx.EventManager().EmitTypedEvent(
					&types.EventStatusDowngradeCanceled{
						Address: acc,
					},
				); err != nil { panic(err) }
			}
			var nextStatus = value.Status
			for {
				if nextStatus == types.MaximumStatus {
					break
				}
				nextStatus++
				checkResult, err = checkStatusRequirements(nextStatus, value, bu)
				if err != nil {
					return err
				}
				if !checkResult.Overall {
					nextStatus--
					break
				}
			}
			if nextStatus > value.Status {
				if err := bu.ctx.EventManager().EmitTypedEvent(
					&types.EventStatusUpdated{
						Address: acc,
						Before:  value.Status,
						After:   nextStatus,
					},
				); err != nil { panic(err) }
				bu.k.setStatus(bu.ctx, &value, nextStatus, acc)
				bu.addCallback(StatusUpdatedCallback, acc)
			}
		}
	}
	if err := bu.set(acc, value); err != nil { return err }
	return nil
}

func (bu *bunchUpdater) addCallback(eventName string, acc string) {
	bu.callbacks = append(bu.callbacks, callback{event: eventName, acc: acc})
}

func (bu *bunchUpdater) commit() error {
	store := bu.ctx.KVStore(bu.k.storeKey)
	for _, pair := range bu.data {
		store.Set(pair.key, pair.value)
	}
	sort.Sort(&bu.callbacks)
	for i, cb := range bu.callbacks {
		if i > 0 && bu.callbacks[i-1].Eq(cb) {
			continue
		}
		if err := bu.k.callback(cb.event, bu.ctx, cb.acc); err != nil {
			return sdkerrors.Wrap(err, cb.event+" callback failed for "+cb.acc)
		}
	}
	return nil
}
