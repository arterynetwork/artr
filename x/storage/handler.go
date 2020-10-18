package storage

import (
	"github.com/arterynetwork/artr/x/storage/types"
	"fmt"

	b64 "encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates an sdk.Handler for all the storage type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSetStorageData:
			return handleMsgSetStorageData(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// handleMsgPaySubscription process payments for subscription
func handleMsgSetStorageData(ctx sdk.Context, k Keeper, msg types.MsgSetStorageData) (*sdk.Result, error) {
	bz, err := b64.StdEncoding.DecodeString(msg.Data)

	if len(bz) > 10*1024 {
		return nil, types.ErrDataToLong
	}

	if err != nil {
		return nil, err
	}

	k.SetData(ctx, msg.Address, bz)
	k.SetCurrent(ctx, msg.Address, msg.Size)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
