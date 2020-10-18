package keeper

import (
	"github.com/arterynetwork/artr/util"
	"github.com/arterynetwork/artr/x/referral/types"
	"bytes"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"sort"
)

type kvRecord struct {
	key   []byte
	value []byte
}

type callback struct {
	event string
	acc   sdk.AccAddress
}
func (x callback) Eq(y callback) bool {
	return x.event == y.event && x.acc.Equals(y.acc)
}

type callbacks []callback

func (cbz *callbacks) Len() int {
	return len(*cbz)
}
func (cbz *callbacks) Less(i, j int) bool {
	x := (*cbz)[i]
	y := (*cbz)[j]

	res := bytes.Compare(x.acc, y.acc)
	if res < 0 {
		return true
	} else if res > 0 {
		return false
	} else {
		return x.event < y.event
	}
}
func (cbz *callbacks) Swap (i, j int) {
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

func newBunchUpdater(k Keeper, ctx sdk.Context) bunchUpdater {
	return bunchUpdater{
		k:         k,
		ctx:       ctx,
		data:      nil,
		callbacks: nil,
	}
}

func (bu *bunchUpdater) set(acc sdk.AccAddress, value types.R) error {
	keyBytes := []byte(acc)
	valueBytes, err := bu.k.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		return err
	}
	bu.data = append(bu.data, kvRecord{
		key:   keyBytes,
		value: valueBytes,
	})
	return nil
}

func (bu *bunchUpdater) get(acc sdk.AccAddress) (types.R, error) {
	var (
		keyBytes   = []byte(acc)
		valueBytes = []byte(nil)

		value types.R
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
	err := bu.k.cdc.UnmarshalBinaryLengthPrefixed(valueBytes, &value)
	return value, err
}

const StatusDowngradeAfter = util.BlocksOneMonth
func (bu *bunchUpdater) update(acc sdk.AccAddress, checkForStatusUpdate bool, callback func(value *types.R)) error {
	value, err := bu.get(acc)
	if err != nil {
		bu.k.Logger(bu.ctx).Info("Cannot update, no such account", "addr", acc)
		return nil
	}
	callback(&value)
	if checkForStatusUpdate {
		checkResult, err := statusRequirements[value.Status](value, *bu)
		if err != nil { return err }
		if !checkResult.Overall {
			if value.StatusDowngradeAt == -1 {
				downgradeAt := bu.ctx.BlockHeight() + StatusDowngradeAfter
				value.StatusDowngradeAt = downgradeAt
				payload := []byte(acc)
				err = bu.k.scheduleKeeper.ScheduleTask(bu.ctx, uint64(downgradeAt), "referral/downgrade", &payload)
				if err != nil {
					return err
				}
				bu.ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeStatusWillBeDowngraded,
						sdk.NewAttribute(types.AttributeKeyAddress, acc.String()),
						sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", downgradeAt)),
					),
				)
			}
		} else {
			if value.StatusDowngradeAt != -1 {
				value.StatusDowngradeAt = -1
				bu.ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeStatusDowngradeCanceled,
						sdk.NewAttribute(types.AttributeKeyAddress, acc.String()),
					),
				)
			}
			var nextStatus = value.Status
			for {
				if nextStatus == types.MaximumStatus {
					break
				}
				nextStatus++
				checkResult, err = statusRequirements[nextStatus](value, *bu)
				if err != nil {
					return err
				}
				if !checkResult.Overall {
					nextStatus--
					break
				}
			}
			if nextStatus > value.Status {
				bu.ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeStatusUpdated,
						sdk.NewAttribute(types.AttributeKeyAddress, acc.String()),
						sdk.NewAttribute(types.AttributeKeyStatusBefore, value.Status.String()),
						sdk.NewAttribute(types.AttributeKeyStatusAfter, nextStatus.String()),
					),
				)
				bu.k.setStatus(bu.ctx, &value, nextStatus, acc)
				bu.addCallback(StatusUpdatedCallback, acc)
			}
		}
	}
	valueBytes, err := bu.k.cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		bu.k.Logger(bu.ctx).Error("Cannot marshal", "value", value)
		return err
	}
	bu.data = append(bu.data, kvRecord{
		key:   []byte(acc),
		value: valueBytes,
	})
	return nil
}

func (bu *bunchUpdater) addCallback(eventName string, acc sdk.AccAddress) {
	bu.callbacks = append(bu.callbacks, callback{event: eventName, acc: acc})
}

func (bu *bunchUpdater) commit() error {
	store := bu.ctx.KVStore(bu.k.storeKey)
	for _, pair := range bu.data {
		store.Set(pair.key, pair.value)
	}
	sort.Sort(&bu.callbacks)
	for i, cb := range bu.callbacks {
		if i > 0 && bu.callbacks[i-1].Eq(cb) { continue }
		if err := bu.k.callback(cb.event, bu.ctx, cb.acc); err != nil {
			return sdkerrors.Wrap(err, cb.event +" callback failed for "+cb.acc.String())
		}
	}
	return nil
}
