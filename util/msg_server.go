package util

import (
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TagTx adds a `message` event with `module` and `action` attributes to the context EventManager.
func TagTx(ctx sdk.Context, module string, tx proto.Message) {
	if err := ctx.EventManager().EmitTypedEvent(
		&EventMessage{
			Module: module,
			Action: proto.MessageName(tx),
		},
	); err != nil {
		panic(err)
	}
}

func (EventMessage) XXX_MessageName() string { return "message" }
