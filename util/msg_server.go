package util

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TagTx adds a `message` event with `module` and `action` attributes to the context EventManager.
func TagTx(ctx sdk.Context, module string, tx proto.Message) {
	EmitEvent(ctx,
		&EventMessage{
			Module: module,
			Action: proto.MessageName(tx),
		},
	)
}

func (EventMessage) XXX_MessageName() string { return "message" }


// EmitEvent emits a typed event like the sdk.EventManager.EmitTypedEvent method does, but it unquotes string
// attribute values so Tendermint could search txs by them.
//
// See https://github.com/tendermint/tendermint/issues/6809
func EmitEvent(ctx sdk.Context, tev proto.Message) {
	//goland:noinspection GoDeprecation
	ctx.EventManager().EmitEvent(patchEvent(tev))
}

// EmitEvents emits typed events like the sdk.EventManager.EmitTypedEvents method does, but it unquotes string
// attribute values so Tendermint could search txs by them.
//
// See https://github.com/tendermint/tendermint/issues/6809
func EmitEvents(ctx sdk.Context, tevs... proto.Message) {
	events := make([]sdk.Event, 0, len(tevs))
	for _, tev := range tevs {
		events = append(events, patchEvent(tev))
	}
	//goland:noinspection GoDeprecation
	ctx.EventManager().EmitEvents(events)
}

func patchEvent(tev proto.Message) sdk.Event {
	evtType := proto.MessageName(tev)
	evtJSON, err := codec.ProtoMarshalJSON(tev, nil)
	if err != nil {
		panic(errors.Wrap(err, "cannot marshal proto"))
	}

	var attrMap map[string]json.RawMessage
	err = json.Unmarshal(evtJSON, &attrMap)
	if err != nil {
		panic(errors.Wrap(err, "cannot unmarshal JSON"))
	}

	attrs := make([]abci.EventAttribute, 0, len(attrMap))
	for k, v := range attrMap {
		if bytes.Equal(v, []byte("null")) {
			continue
		}
		m, err := regexp.Match(`^".*"$`, v)
		if err != nil {
			panic(errors.Wrap(err, "regexp failed"))
		}
		if m {
			s := string(v)
			s = s[1 : len(s)-1]
			s = strings.ReplaceAll(s, `\"`, `"`)
			s = strings.ReplaceAll(s, `\\`, `\`)
			v = []byte(s)
		}

		attrs = append(attrs, abci.EventAttribute{
			Key:   []byte(k),
			Value: v,
		})
	}

	return sdk.Event{
		Type:       evtType,
		Attributes: attrs,
	}
}
